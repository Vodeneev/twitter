package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Vodeneev/twitter/backend/internal/auth"
	"github.com/Vodeneev/twitter/backend/internal/social"
)

func (a *api) getProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := a.Users.GetByUsernameWithViewer(r.Context(), username, viewerID(r.Context()))
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.presentPublicUser(u)})
}

func (a *api) updateMe(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	var req struct {
		DisplayName *string `json:"displayName"`
		Bio         *string `json:"bio"`
		Location    *string `json:"location"`
		Website     *string `json:"website"`
		AvatarKey   *string `json:"avatarKey"`
		HeaderKey   *string `json:"headerKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	u, err := a.Users.UpdateProfile(r.Context(), me.ID, auth.UpdateProfileInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		Location:    req.Location,
		Website:     req.Website,
		AvatarKey:   req.AvatarKey,
		HeaderKey:   req.HeaderKey,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.presentUser(u)})
}

func (a *api) listFollowers(w http.ResponseWriter, r *http.Request) {
	a.followList(w, r, false)
}

func (a *api) listFollowing(w http.ResponseWriter, r *http.Request) {
	a.followList(w, r, true)
}

func (a *api) followList(w http.ResponseWriter, r *http.Request, following bool) {
	username := chi.URLParam(r, "username")
	target, err := a.Users.GetByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	limit := queryInt(r, "limit", 50)
	var users []*auth.User
	if following {
		users, err = a.Users.Following(r.Context(), target.ID, viewerID(r.Context()), limit)
	} else {
		users, err = a.Users.Followers(r.Context(), target.ID, viewerID(r.Context()), limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load users")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": a.presentPublicUsers(users)})
}

func (a *api) searchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	users, err := a.Users.Search(r.Context(), q, viewerID(r.Context()), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "search failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": a.presentPublicUsers(users)})
}

func (a *api) suggestions(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	users, err := a.Users.Suggestions(r.Context(), me.ID, queryInt(r, "limit", 5))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load suggestions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": a.presentPublicUsers(users)})
}

func (a *api) follow(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	target, err := a.Users.GetByUsername(r.Context(), chi.URLParam(r, "username"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	created, err := a.Social.Follow(r.Context(), me.ID, target.ID)
	if err != nil {
		if errors.Is(err, social.ErrSelfFollow) {
			writeError(w, http.StatusBadRequest, "self_follow", "cannot follow yourself")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to follow")
		return
	}
	if created {
		a.notify(r, target.ID, me.ID, "follow", nil)
	}
	writeJSON(w, http.StatusOK, map[string]any{"following": true})
}

func (a *api) unfollow(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	target, err := a.Users.GetByUsername(r.Context(), chi.URLParam(r, "username"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	if err := a.Social.Unfollow(r.Context(), me.ID, target.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to unfollow")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"following": false})
}

func queryInt(r *http.Request, key string, fallback int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
