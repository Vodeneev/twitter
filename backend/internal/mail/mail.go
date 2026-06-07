// Package mail provides a minimal transactional-email interface over SMTP.
// Resend is supported as an SMTP provider (smtp.resend.com:465, user "resend",
// password = API key), so no provider-specific dependency is required.
package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type Message struct {
	From     string
	To       []string
	Subject  string
	TextBody string
	HTMLBody string
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	// TLSMode: "none" | "starttls" | "tls".
	TLSMode string
}

type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	if deadline, ok := ctx.Deadline(); ok {
		dialer.Deadline = deadline
	}

	var conn net.Conn
	var err error
	if s.cfg.TLSMode == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: s.cfg.Host})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if s.cfg.TLSMode == "starttls" {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if s.cfg.Username != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	from := msg.From
	if from == "" {
		from = s.cfg.From
	}

	envelopeFrom, err := extractAddress(from)
	if err != nil {
		return fmt.Errorf("parse from: %w", err)
	}
	if err := client.Mail(envelopeFrom); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, to := range msg.To {
		a, err := extractAddress(to)
		if err != nil {
			return fmt.Errorf("parse rcpt %q: %w", to, err)
		}
		if err := client.Rcpt(a); err != nil {
			return fmt.Errorf("rcpt %q: %w", to, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(renderMIME(from, msg)); err != nil {
		_ = w.Close()
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return client.Quit()
}

func renderMIME(from string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(msg.To, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" && msg.TextBody != "" {
		boundary := "yapper-boundary-ch42"
		fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)
		fmt.Fprintf(&b, "--%s\r\n", boundary)
		b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		b.WriteString(msg.TextBody)
		b.WriteString("\r\n\r\n")
		fmt.Fprintf(&b, "--%s\r\n", boundary)
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		b.WriteString(msg.HTMLBody)
		b.WriteString("\r\n\r\n")
		fmt.Fprintf(&b, "--%s--\r\n", boundary)
	} else if msg.HTMLBody != "" {
		b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		b.WriteString(msg.HTMLBody)
	} else {
		b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		b.WriteString(msg.TextBody)
	}
	return []byte(b.String())
}

// extractAddress returns the bare mailbox for SMTP envelope commands
// (MAIL FROM / RCPT TO reject "Display Name <addr>" form).
func extractAddress(s string) (string, error) {
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}

// NullSender logs instead of sending; used when SMTP is not configured.
type NullSender struct{}

func (NullSender) Send(_ context.Context, msg Message) error {
	slog.Info("mail.null: drop message", "to", msg.To, "subject", msg.Subject)
	return nil
}
