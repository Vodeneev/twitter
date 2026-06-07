package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Vodeneev/twitter/backend/internal/dm"
	"github.com/Vodeneev/twitter/backend/internal/realtime"
)

func (a *api) listConversations(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	items, err := a.DM.ListConversations(r.Context(), me.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load conversations")
		return
	}
	if items == nil {
		items = []dm.Conversation{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) openConversation(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	other, err := a.Users.GetByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	convID, err := a.DM.GetOrCreate(r.Context(), me.ID, other.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot_open", "cannot open conversation")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"conversationId": convID, "other": a.presentPublicUser(other)})
}

func (a *api) listMessages(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid conversation id")
		return
	}
	items, next, err := a.DM.ListMessages(r.Context(), convID, me.ID, r.URL.Query().Get("cursor"), queryInt(r, "limit", 40))
	if err != nil {
		if errors.Is(err, dm.ErrNotParticipant) {
			writeError(w, http.StatusForbidden, "forbidden", "not a participant")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load messages")
		return
	}
	if items == nil {
		items = []dm.Message{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "nextCursor": next})
}

func (a *api) sendMessage(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid conversation id")
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	msg, err := a.DM.SendMessage(r.Context(), convID, me.ID, req.Body)
	if err != nil {
		switch {
		case errors.Is(err, dm.ErrNotParticipant):
			writeError(w, http.StatusForbidden, "forbidden", "not a participant")
		case errors.Is(err, dm.ErrEmptyMessage):
			writeError(w, http.StatusBadRequest, "empty", "message is empty")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to send message")
		}
		return
	}

	// Push to the other participant(s) in real time.
	if a.Hub != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if parts, err := a.DM.Participants(ctx, convID); err == nil {
			for _, uid := range parts {
				a.Hub.Publish(uid, realtime.Event{Type: "message", Data: msg})
			}
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": msg})
}

func (a *api) markConversationRead(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid conversation id")
		return
	}
	if err := a.DM.MarkRead(r.Context(), convID, me.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to mark read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
