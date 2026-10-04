package handler

import (
	"context"
	"log"
	"net/http"
	"time"
	"github.com/2727-ask/jobstar/internal/db"
	"github.com/2727-ask/jobstar/internal/queue"
)

// Handler holds the dependencies that routes need.
type Handler struct {
	db    *db.DB
	queue *queue.Publisher
}

func NewRouter(database *db.DB, q *queue.Publisher) http.Handler {
	h := &Handler{db: database, queue: q}
	mux := http.NewServeMux()

	// Probes
	health := NewHealth(map[string]Checker{
		"postgres": database.Health,
		"rabbitmq": func(ctx context.Context) error { return q.Health() },
	})
	mux.Handle("GET /healthz", health)
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API
	mux.HandleFunc("POST /schedule", h.schedule)

	return chain(mux, recoverer, logger)
}

// chain wraps h so that the first middleware runs outermost.
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}