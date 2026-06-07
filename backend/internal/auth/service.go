package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type Mailer interface {
	SendVerification(ctx context.Context, u *User, link string, locale string) error
	SendPasswordReset(ctx context.Context, u *User, link string, locale string) error
}

type Service struct {
	users          *UserRepository
	sessions       *SessionRepository
	verifications  *VerificationRepository
	passwordResets *PasswordResetRepository
	mailer         Mailer
	siteBaseURL    string
}

func NewService(
	users *UserRepository,
	sessions *SessionRepository,
	verifications *VerificationRepository,
	passwordResets *PasswordResetRepository,
	mailer Mailer,
	siteBaseURL string,
) *Service {
	if mailer == nil {
		mailer = nullMailer{}
	}
	return &Service{
		users:          users,
		sessions:       sessions,
		verifications:  verifications,
		passwordResets: passwordResets,
		mailer:         mailer,
		siteBaseURL:    siteBaseURL,
	}
}

type nullMailer struct{}

func (nullMailer) SendVerification(context.Context, *User, string, string) error  { return nil }
func (nullMailer) SendPasswordReset(context.Context, *User, string, string) error { return nil }

// Register does not issue a session: the user must verify their email first.
func (s *Service) Register(ctx context.Context, in RegisterInput, locale string) (*User, error) {
	in.Email = NormalizeEmail(in.Email)
	in.Username = normalizeUsername(in.Username)
	if err := ValidateRegister(in); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.users.Create(ctx, in.Username, in.Email, string(hash), in.DisplayName)
	if err != nil {
		return nil, err
	}

	if err := s.SendVerificationEmail(ctx, u, locale); err != nil {
		return u, fmt.Errorf("send verification: %w", err)
	}
	return u, nil
}

func (s *Service) SendVerificationEmail(ctx context.Context, u *User, locale string) error {
	if u.EmailVerified {
		return ErrAlreadyVerified
	}
	token, err := s.verifications.Issue(ctx, u.ID)
	if err != nil {
		return err
	}
	link := s.buildLink(locale, "verify-email", token)
	return s.mailer.SendVerification(ctx, u, link, locale)
}

// ResendVerification never reveals whether the email exists.
func (s *Service) ResendVerification(ctx context.Context, email string, locale string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return nil
	}
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil
		}
		return err
	}
	if u.EmailVerified {
		return nil
	}
	_ = s.SendVerificationEmail(ctx, u, locale)
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) (*User, error) {
	u, err := s.verifications.Consume(ctx, token)
	if err != nil && !errors.Is(err, ErrAlreadyVerified) {
		return nil, err
	}
	return u, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string, locale string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return nil
	}
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil
		}
		return err
	}
	token, err := s.passwordResets.Issue(ctx, u.ID)
	if err != nil {
		return err
	}
	link := s.buildLink(locale, "reset-password", token)
	return s.mailer.SendPasswordReset(ctx, u, link, locale)
}

// ResetPassword consumes a token and sets a new password; all sessions are revoked.
func (s *Service) ResetPassword(ctx context.Context, token string, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.passwordResets.ConsumeAndSetPassword(ctx, token, string(hash))
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*User, *Session, error) {
	in.Identifier = normalizeIdentifier(in.Identifier)
	if in.Identifier == "" || in.Password == "" {
		return nil, nil, ErrInvalidCredentials
	}

	u, hash, err := s.users.GetByLogin(ctx, in.Identifier)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Constant-time dummy compare to avoid leaking user existence via timing.
			_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$invalidinvalidinvalidinvalidinvalidinvalidinvalidinvalidinv"), []byte(in.Password))
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	if u.IsBanned {
		return nil, nil, ErrUserBanned
	}
	if !u.EmailVerified {
		return u, nil, ErrEmailNotVerified
	}

	ttl := DefaultSessionTTL
	if in.RememberMe {
		ttl = RememberMeSessionTTL
	}
	sess, err := s.sessions.Create(ctx, u.ID, ttl, in.RememberMe, in.UserAgent, in.IP)
	if err != nil {
		return nil, nil, err
	}
	return u, sess, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Delete(ctx, sessionID)
}

func (s *Service) Authenticate(ctx context.Context, sessionID uuid.UUID) (*User, *Session, error) {
	sess, u, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	if u.IsBanned {
		_ = s.sessions.Delete(ctx, sess.ID)
		return nil, nil, ErrUserBanned
	}
	return u, sess, nil
}

// buildLink builds a localized frontend URL like https://host/ru/verify-email?token=...
func (s *Service) buildLink(locale, page, token string) string {
	normalizedLocale := NormalizeLocale(locale)
	base := strings.TrimRight(s.siteBaseURL, "/")

	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Sprintf("%s/%s/%s?token=%s", base, normalizedLocale, page, token)
	}

	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "/ru" || u.Path == "/en" {
		u.Path = ""
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/" + normalizedLocale + "/" + page

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}

// NormalizeLocale maps anything to a supported locale; default is English.
func NormalizeLocale(raw string) string {
	l := strings.ToLower(strings.TrimSpace(raw))
	if len(l) >= 2 {
		l = l[:2]
	}
	switch l {
	case "ru", "en":
		return l
	default:
		return "en"
	}
}
