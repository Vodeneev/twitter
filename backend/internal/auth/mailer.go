package auth

import (
	"context"
	"fmt"

	"github.com/Vodeneev/twitter/backend/internal/mail"
)

type VerificationMailer struct {
	M        mail.Mailer
	From     string
	SiteName string
}

func displayName(u *User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Username
}

func (m VerificationMailer) SendVerification(ctx context.Context, u *User, link string, locale string) error {
	var subject, text, html string
	name := displayName(u)
	switch NormalizeLocale(locale) {
	case "ru":
		subject = fmt.Sprintf("[%s] Подтвердите адрес e-mail", m.SiteName)
		text = fmt.Sprintf(`Привет, %s!

Добро пожаловать в %s. Подтвердите свой e-mail, открыв ссылку ниже:

%s

Ссылка действует 24 часа. Если вы не регистрировались — просто проигнорируйте это письмо.
`, name, m.SiteName, link)
		html = fmt.Sprintf(emailHTML,
			"Добро пожаловать в "+m.SiteName,
			fmt.Sprintf("Привет, %s! Подтвердите адрес e-mail, чтобы активировать аккаунт.", name),
			link, "Подтвердить e-mail",
			"Ссылка действует 24 часа. Если вы не регистрировались, проигнорируйте это письмо.")
	default:
		subject = fmt.Sprintf("[%s] Confirm your email address", m.SiteName)
		text = fmt.Sprintf(`Hi %s,

Welcome to %s. Please confirm your email address by opening the link below:

%s

The link expires in 24 hours. If you did not sign up, ignore this email.
`, name, m.SiteName, link)
		html = fmt.Sprintf(emailHTML,
			"Welcome to "+m.SiteName,
			fmt.Sprintf("Hi %s, please confirm your email address to activate your account.", name),
			link, "Confirm my email",
			"The link expires in 24 hours. If you did not sign up, ignore this message.")
	}

	return m.M.Send(ctx, mail.Message{From: m.From, To: []string{u.Email}, Subject: subject, TextBody: text, HTMLBody: html})
}

func (m VerificationMailer) SendPasswordReset(ctx context.Context, u *User, link string, locale string) error {
	var subject, text, html string
	name := displayName(u)
	switch NormalizeLocale(locale) {
	case "ru":
		subject = fmt.Sprintf("[%s] Сброс пароля", m.SiteName)
		text = fmt.Sprintf(`Привет, %s!

Мы получили запрос на сброс пароля для %s. Откройте ссылку ниже, чтобы задать новый пароль:

%s

Ссылка действует один час. Если это были не вы — проигнорируйте письмо, пароль останется прежним.
`, name, m.SiteName, link)
		html = fmt.Sprintf(emailHTML,
			"Сброс пароля",
			fmt.Sprintf("Привет, %s! Нажмите кнопку ниже, чтобы задать новый пароль для аккаунта %s.", name, m.SiteName),
			link, "Задать новый пароль",
			"Ссылка действует один час. Если вы не запрашивали сброс, проигнорируйте письмо.")
	default:
		subject = fmt.Sprintf("[%s] Reset your password", m.SiteName)
		text = fmt.Sprintf(`Hi %s,

We received a request to reset your password for %s. Open the link below to choose a new password:

%s

The link expires in one hour. If you did not ask for this, ignore this email.
`, name, m.SiteName, link)
		html = fmt.Sprintf(emailHTML,
			"Reset your password",
			fmt.Sprintf("Hi %s, use the button below to set a new password for your %s account.", name, m.SiteName),
			link, "Choose a new password",
			"This link expires in one hour. If you did not request a reset, ignore this message.")
	}

	return m.M.Send(ctx, mail.Message{From: m.From, To: []string{u.Email}, Subject: subject, TextBody: text, HTMLBody: html})
}

const emailHTML = `<!doctype html>
<html><body style="font-family:system-ui,Segoe UI,Roboto,sans-serif;color:#0f1419;background:#f7f9f9;padding:24px;">
  <div style="max-width:520px;margin:0 auto;background:#fff;border:1px solid #eff3f4;border-radius:16px;padding:28px;">
    <h2 style="margin:0 0 12px;color:#FC3F1D">%s</h2>
    <p style="font-size:15px;line-height:1.5">%s</p>
    <p style="margin:24px 0"><a href="%s" style="background:#FC3F1D;color:#fff;padding:12px 20px;border-radius:9999px;text-decoration:none;font-weight:600">%s</a></p>
    <p style="color:#536471;font-size:12px">%s</p>
  </div>
</body></html>`
