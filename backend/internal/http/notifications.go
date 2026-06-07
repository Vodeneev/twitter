package http

import "net/http"

func (a *api) listNotifications(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	page, err := a.Notifications.List(r.Context(), me.ID, r.URL.Query().Get("cursor"), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load notifications")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *api) unreadCount(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	n, err := a.Notifications.UnreadCount(r.Context(), me.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load count")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": n})
}

func (a *api) markNotificationsRead(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	if err := a.Notifications.MarkAllRead(r.Context(), me.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to mark read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
