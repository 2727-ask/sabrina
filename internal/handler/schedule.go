package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

const maxBody = 1 << 20 // 1 MB

func (h *Handler) schedule(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}
	if !json.Valid(body) {
		http.Error(w, "body must be valid JSON", http.StatusBadRequest)
		return
	}

	if err := h.queue.Publish(r.Context(), "jobs", body); err != nil {
		http.Error(w, "queue unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}