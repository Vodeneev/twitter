package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Vodeneev/twitter/backend/internal/auth"
	"github.com/Vodeneev/twitter/backend/internal/dm"
	"github.com/Vodeneev/twitter/backend/internal/notifications"
	"github.com/Vodeneev/twitter/backend/internal/realtime"
	"github.com/Vodeneev/twitter/backend/internal/social"
	"github.com/Vodeneev/twitter/backend/internal/storage"
	"github.com/Vodeneev/twitter/backend/internal/yaps"
)

type Deps struct {
	Pool          *pgxpool.Pool
	AuthService   *auth.Service
	Users         *auth.UserRepository
	Yaps          *yaps.Repository
	Social        *social.Repository
	Notifications *notifications.Repository
	DM            *dm.Repository
	Hub           *realtime.Hub
	Storage       *storage.Client
	PhotoURL      func(string) string
	SiteName      string
	SiteBaseURL   string
	CORSOrigins   []string
	CookieSecure  bool
}

type api struct {
	Deps
}

func NewRouter(d Deps) http.Handler {
	if d.PhotoURL == nil {
		d.PhotoURL = func(string) string { return "" }
	}
	a := &api{Deps: d}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware(d.CORSOrigins))

	if d.AuthService != nil {
		r.Use(loadSession(d.AuthService))
	}

	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(d.Pool))

	r.Route("/api", func(r chi.Router) {
		r.Get("/ping", ping)

		if d.AuthService != nil {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", a.register)
				r.Post("/login", a.login)
				r.Post("/logout", a.logout)
				r.Get("/me", a.me)
				r.Post("/verify-email", a.verifyEmail)
				r.Post("/resend-verification", a.resendVerification)
				r.Post("/forgot-password", a.forgotPassword)
				r.Post("/reset-password", a.resetPassword)
			})
		}

		// Public reads (viewer optional).
		r.Get("/users/{username}", a.getProfile)
		r.Get("/users/{username}/followers", a.listFollowers)
		r.Get("/users/{username}/following", a.listFollowing)
		r.Get("/users/{username}/yaps", a.userYaps)
		r.Get("/users/{username}/replies", a.userReplies)
		r.Get("/users/{username}/media", a.userMedia)
		r.Get("/users/{username}/likes", a.userLikes)
		r.Get("/search/users", a.searchUsers)
		r.Get("/search/yaps", a.searchYaps)
		r.Get("/timeline/global", a.globalTimeline)
		r.Get("/hashtags/{tag}", a.hashtagTimeline)
		r.Get("/yaps/{id}", a.getYap)
		r.Get("/yaps/{id}/replies", a.yapReplies)
		r.Get("/yaps/{id}/thread", a.yapThread)

		// WebSocket (auth via session cookie inside the handler).
		r.Get("/ws", a.websocket)

		// Authenticated writes/reads.
		r.Group(func(r chi.Router) {
			r.Use(requireAuth)

			r.Patch("/me", a.updateMe)
			r.Get("/me/suggestions", a.suggestions)
			r.Get("/timeline/home", a.homeTimeline)
			r.Get("/bookmarks", a.listBookmarks)

			r.Post("/media/presign", a.presign)

			r.Post("/yaps", a.createYap)
			r.Delete("/yaps/{id}", a.deleteYap)
			r.Put("/yaps/{id}/like", a.like)
			r.Delete("/yaps/{id}/like", a.unlike)
			r.Put("/yaps/{id}/repost", a.repost)
			r.Delete("/yaps/{id}/repost", a.unrepost)
			r.Put("/yaps/{id}/bookmark", a.bookmark)
			r.Delete("/yaps/{id}/bookmark", a.unbookmark)

			r.Put("/users/{username}/follow", a.follow)
			r.Delete("/users/{username}/follow", a.unfollow)

			r.Get("/notifications", a.listNotifications)
			r.Get("/notifications/unread-count", a.unreadCount)
			r.Post("/notifications/read", a.markNotificationsRead)

			r.Get("/conversations", a.listConversations)
			r.Post("/conversations", a.openConversation)
			r.Get("/conversations/{id}/messages", a.listMessages)
			r.Post("/conversations/{id}/messages", a.sendMessage)
			r.Post("/conversations/{id}/read", a.markConversationRead)
		})
	})

	return r
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyz(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "no_db"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unreachable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func ping(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
