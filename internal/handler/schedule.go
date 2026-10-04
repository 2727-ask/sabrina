package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/2727-ask/jobstar/internal/model"
)

const maxBody = 1 << 20 // 1 MB

func (h *Handler) schedule(w http.ResponseWriter, r *http.Request) {
	var req model.ScheduleRequest

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Image = strings.TrimSpace(req.Image)

	if req.Name == "" || req.Image == "" {
		writeError(w, http.StatusBadRequest, "name and image are required")
		return
	}

	ctx := r.Context()

	tx, err := h.db.Pool.Begin(ctx)

	if err != nil {
		log.Printf("begin tx: %v", err)
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	defer tx.Rollback(ctx) // no-op after a successful commit

	job := model.Job{
		Name:   req.Name,
		Image:  req.Image,
		Status: model.StatusPending,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO jobs (name, image, status)
		 VALUES ($1, $2, $3)
		 RETURNING id, date`,
		job.Name, job.Image, job.Status,
	).Scan(&job.ID, &job.Date)

	if err != nil {
		log.Printf("insert job: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save job")
		return
	}

	msg, _ := json.Marshal(job)
	if err := h.queue.Publish(ctx, "jobs", msg); err != nil {
		log.Printf("publish job %s: %v", job.ID, err)
		writeError(w, http.StatusServiceUnavailable, "queue unavailable")
		return // deferred Rollback removes the row
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("commit job %s: %v", job.ID, err)
		writeError(w, http.StatusInternalServerError, "could not save job")
		return
	}

	writeJSON(w, http.StatusCreated, job)

}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
