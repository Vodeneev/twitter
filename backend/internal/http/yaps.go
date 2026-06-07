package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Vodeneev/twitter/backend/internal/yaps"
)

func (a *api) createYap(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	var req struct {
		Content   string   `json:"content"`
		ReplyToID *string  `json:"replyToId"`
		QuoteOfID *string  `json:"quoteOfId"`
		MediaKeys []string `json:"mediaKeys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	in := yaps.CreateInput{AuthorID: me.ID, Content: req.Content, MediaKeys: req.MediaKeys}
	if id, ok := parseOptionalUUID(req.ReplyToID); ok {
		in.ReplyToID = id
	}
	if id, ok := parseOptionalUUID(req.QuoteOfID); ok {
		in.QuoteOfID = id
	}

	yap, mentioned, err := a.Yaps.Create(r.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, yaps.ErrEmpty):
			writeError(w, http.StatusBadRequest, "empty", "yap must have text or media")
		case errors.Is(err, yaps.ErrTooLong):
			writeError(w, http.StatusBadRequest, "too_long", "yap exceeds 280 characters")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create yap")
		}
		return
	}

	if in.ReplyToID != nil && yap.QuoteOf == nil {
		if parent, err := a.Yaps.GetByID(r.Context(), *in.ReplyToID, &me.ID); err == nil {
			a.notify(r, parent.Author.ID, me.ID, "reply", &yap.ID)
		}
	}
	if in.QuoteOfID != nil {
		if quoted, err := a.Yaps.GetByID(r.Context(), *in.QuoteOfID, &me.ID); err == nil {
			a.notify(r, quoted.Author.ID, me.ID, "quote", &yap.ID)
		}
	}
	for _, uid := range mentioned {
		a.notify(r, uid, me.ID, "mention", &yap.ID)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"yap": yap})
}

func (a *api) getYap(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid yap id")
		return
	}
	yap, err := a.Yaps.GetByID(r.Context(), id, viewerID(r.Context()))
	if err != nil {
		if errors.Is(err, yaps.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "yap not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load yap")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"yap": yap})
}

func (a *api) deleteYap(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid yap id")
		return
	}
	if err := a.Yaps.Delete(r.Context(), id, me.ID); err != nil {
		switch {
		case errors.Is(err, yaps.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "yap not found")
		case errors.Is(err, yaps.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", "not your yap")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete yap")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) homeTimeline(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	page, err := a.Yaps.HomeTimeline(r.Context(), me.ID, r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load timeline")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) globalTimeline(w http.ResponseWriter, r *http.Request) {
	page, err := a.Yaps.GlobalTimeline(r.Context(), viewerID(r.Context()), r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load timeline")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) yapReplies(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid yap id")
		return
	}
	page, err := a.Yaps.Replies(r.Context(), id, viewerID(r.Context()), r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load replies")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) yapThread(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid yap id")
		return
	}
	viewer := viewerID(r.Context())
	yap, err := a.Yaps.GetByID(r.Context(), id, viewer)
	if err != nil {
		if errors.Is(err, yaps.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "yap not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load thread")
		return
	}
	ancestors, err := a.Yaps.Ancestors(r.Context(), id, viewer)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load thread")
		return
	}
	if ancestors == nil {
		ancestors = []yaps.Yap{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"yap": yap, "ancestors": ancestors})
}

func (a *api) hashtagTimeline(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	page, err := a.Yaps.HashtagTimeline(r.Context(), tag, viewerID(r.Context()), r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load hashtag")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) searchYaps(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page, err := a.Yaps.SearchYaps(r.Context(), q, viewerID(r.Context()), r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "search failed")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) listBookmarks(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	page, err := a.Yaps.Bookmarks(r.Context(), me.ID, r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load bookmarks")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// --- profile timelines ----------------------------------------------------

func (a *api) profileFeed(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, owner uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (yaps.Page, error)) {
	owner, err := a.Users.GetByUsername(r.Context(), chi.URLParam(r, "username"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	page, err := fn(r.Context(), owner.ID, viewerID(r.Context()), r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load feed")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) userYaps(w http.ResponseWriter, r *http.Request) {
	a.profileFeed(w, r, a.Yaps.UserYaps)
}
func (a *api) userReplies(w http.ResponseWriter, r *http.Request) {
	a.profileFeed(w, r, a.Yaps.UserReplies)
}
func (a *api) userMedia(w http.ResponseWriter, r *http.Request) {
	a.profileFeed(w, r, a.Yaps.UserMedia)
}
func (a *api) userLikes(w http.ResponseWriter, r *http.Request) {
	a.profileFeed(w, r, a.Yaps.UserLikes)
}

func parseOptionalUUID(s *string) (*uuid.UUID, bool) {
	if s == nil || *s == "" {
		return nil, false
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, false
	}
	return &id, true
}
