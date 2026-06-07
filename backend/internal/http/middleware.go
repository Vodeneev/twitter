package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Vodeneev/twitter/backend/internal/auth"
)

type ctxKey int

const ctxKeyUser ctxKey = iota

func UserFromContext(ctx context.Context) (*auth.User, bool) {
	u, ok := ctx.Value(ctxKeyUser).(*auth.User)
	return u, ok
}

// viewerID returns the current user's id, or nil for anonymous requests.
func viewerID(ctx context.Context) *uuid.UUID {
	if u, ok := UserFromContext(ctx); ok {
		id := u.ID
		return &id
	}
	return nil
}

func loadSession(svc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(sessionCookieName)
			if err != nil || c.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			id, err := uuid.Parse(c.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			u, _, err := svc.Authenticate(r.Context(), id)
			if err != nil {
				if errors.Is(err, auth.ErrSessionExpired) || errors.Is(err, auth.ErrUserBanned) {
					clearCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKeyUser, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
