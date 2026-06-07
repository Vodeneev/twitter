package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var allowedImageTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/gif":  "gif",
}

// presign returns a pre-signed PUT URL so the browser can upload an image
// directly to object storage. The API only ever stores the resulting key.
func (a *api) presign(w http.ResponseWriter, r *http.Request) {
	me, _ := UserFromContext(r.Context())
	if a.Storage == nil {
		writeError(w, http.StatusServiceUnavailable, "storage_disabled", "media storage is not configured")
		return
	}
	var req struct {
		ContentType string `json:"contentType"`
		Kind        string `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}
	ext, ok := allowedImageTypes[req.ContentType]
	if !ok {
		writeError(w, http.StatusBadRequest, "unsupported_type", "only jpeg, png, webp and gif are allowed")
		return
	}
	folder := "yap"
	switch req.Kind {
	case "avatar":
		folder = "avatar"
	case "header":
		folder = "header"
	}
	key := fmt.Sprintf("%s/%s/%s.%s", folder, me.ID.String(), uuid.NewString(), ext)

	url, err := a.Storage.PresignPut(r.Context(), key, req.ContentType, 10*time.Minute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to presign upload")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uploadUrl": url,
		"key":       key,
		"publicUrl": a.PhotoURL(key),
	})
}
