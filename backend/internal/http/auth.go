package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Vodeneev/twitter/backend/internal/auth"
)

const sessionCookieName = "yapper_session"

type registerRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	Locale      string `json:"locale"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

func (a *api) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	u, err := a.AuthService.Register(r.Context(), auth.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	}, preferredLocale(req.Locale, r.Header.Get("Accept-Language")))
	if err != nil {
		if u != nil {
			slog.Warn("registration ok but verification email failed", "user_id", u.ID, "error", err)
			writeJSON(w, http.StatusCreated, map[string]any{"user": a.presentUser(u), "verificationEmailSent": false})
			return
		}
		a.translateAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": a.presentUser(u), "verificationEmailSent": true})
}

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	u, sess, err := a.AuthService.Login(r.Context(), auth.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
		RememberMe: req.RememberMe,
		UserAgent:  r.UserAgent(),
		IP:         clientIP(r),
	})
	if err != nil {
		if errors.Is(err, auth.ErrEmailNotVerified) {
			writeJSON(w, http.StatusForbidden, map[string]any{
				"error": map[string]string{"code": "email_not_verified", "message": "email not verified"},
				"email": u.Email,
			})
			return
		}
		a.translateAuthError(w, err)
		return
	}
	a.setSessionCookie(w, sess)
	writeJSON(w, http.StatusOK, map[string]any{"user": a.presentUser(u)})
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		if id, err := uuid.Parse(c.Value); err == nil {
			_ = a.AuthService.Logout(r.Context(), id)
		}
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"user": nil})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.presentUser(u)})
}

func (a *api) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "missing or malformed token")
		return
	}
	u, err := a.AuthService.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, auth.ErrVerificationInvalid) {
			writeError(w, http.StatusBadRequest, "invalid_token", "token is invalid or expired")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "verification failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.presentUser(u)})
}

func (a *api) resendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email  string `json:"email"`
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	_ = a.AuthService.ResendVerification(r.Context(), req.Email, preferredLocale(req.Locale, r.Header.Get("Accept-Language")))
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email  string `json:"email"`
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	if err := a.AuthService.RequestPasswordReset(r.Context(), req.Email, preferredLocale(req.Locale, r.Header.Get("Accept-Language"))); err != nil {
		slog.Warn("forgot-password: send failed", "error", err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "missing token")
		return
	}
	if err := a.AuthService.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		var verr *auth.ValidationError
		if errors.As(err, &verr) {
			writeValidation(w, verr)
			return
		}
		if errors.Is(err, auth.ErrPasswordResetInvalid) {
			writeError(w, http.StatusBadRequest, "invalid_token", "token is invalid or expired")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "password reset failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeValidation(w http.ResponseWriter, verr *auth.ValidationError) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error": map[string]any{"code": "validation", "message": verr.Message, "field": verr.Field},
	})
}

func (a *api) translateAuthError(w http.ResponseWriter, err error) {
	var verr *auth.ValidationError
	switch {
	case errors.As(err, &verr):
		writeValidation(w, verr)
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email_taken", "email is already registered")
	case errors.Is(err, auth.ErrUsernameTaken):
		writeError(w, http.StatusConflict, "username_taken", "username is already taken")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
	case errors.Is(err, auth.ErrUserBanned):
		writeError(w, http.StatusForbidden, "account_banned", "account is blocked")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "authentication failed")
	}
}

func (a *api) setSessionCookie(w http.ResponseWriter, s *auth.Session) {
	c := &http.Cookie{
		Name:     sessionCookieName,
		Value:    s.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   a.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	if s.RememberMe {
		c.Expires = s.ExpiresAt
		c.MaxAge = int(time.Until(s.ExpiresAt).Seconds())
	}
	http.SetCookie(w, c)
}

func (a *api) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func preferredLocale(explicit, acceptLanguage string) string {
	if l := auth.NormalizeLocale(explicit); l != "" && explicit != "" {
		return l
	}
	for _, part := range strings.Split(acceptLanguage, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		if idx := strings.Index(token, ";"); idx >= 0 {
			token = token[:idx]
		}
		lower := strings.ToLower(token)
		if strings.HasPrefix(lower, "ru") {
			return "ru"
		}
		if strings.HasPrefix(lower, "en") {
			return "en"
		}
	}
	return "en"
}
