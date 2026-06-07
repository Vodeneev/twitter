package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Vodeneev/twitter/backend/internal/realtime"
	"github.com/Vodeneev/twitter/backend/internal/social"
)

func yapIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid yap id")
		return uuid.Nil, false
	}
	return id, true
}

func mapInteractionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, social.ErrYapNotFound):
		writeError(w, http.StatusNotFound, "not_found", "yap not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "operation failed")
	}
}

func (a *api) like(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	created, author, err := a.Social.Like(r.Context(), me.ID, id)
	if err != nil {
		mapInteractionError(w, err)
		return
	}
	if created {
		a.notify(r, author, me.ID, "like", &id)
	}
	writeJSON(w, http.StatusOK, map[string]any{"liked": true})
}

func (a *api) unlike(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	if err := a.Social.Unlike(r.Context(), me.ID, id); err != nil {
		mapInteractionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"liked": false})
}

func (a *api) repost(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	created, author, err := a.Social.Repost(r.Context(), me.ID, id)
	if err != nil {
		mapInteractionError(w, err)
		return
	}
	if created {
		a.notify(r, author, me.ID, "repost", &id)
	}
	writeJSON(w, http.StatusOK, map[string]any{"reposted": true})
}

func (a *api) unrepost(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	if err := a.Social.Unrepost(r.Context(), me.ID, id); err != nil {
		mapInteractionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reposted": false})
}

func (a *api) bookmark(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	if err := a.Social.Bookmark(r.Context(), me.ID, id); err != nil {
		mapInteractionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookmarked": true})
}

func (a *api) unbookmark(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, ok := yapIDParam(w, r)
	if !ok {
		return
	}
	if err := a.Social.Unbookmark(r.Context(), me.ID, id); err != nil {
		mapInteractionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookmarked": false})
}

// notify persists a notification and pushes it to the recipient over WebSocket.
// Runs detached so a cancelled request context cannot drop the side effect.
func (a *api) notify(_ *http.Request, userID, actorID uuid.UUID, typ string, yapID *uuid.UUID) {
	if a.Notifications == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n, err := a.Notifications.Create(ctx, userID, actorID, typ, yapID)
	if err != nil || n == nil {
		return
	}
	if a.Hub != nil {
		a.Hub.Publish(userID, realtime.Event{Type: "notification", Data: n})
	}
}
