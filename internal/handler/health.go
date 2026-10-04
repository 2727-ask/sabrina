package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Checker is anything that can report its own health.
type Checker func(ctx context.Context) error

type Health struct {
	checks map[string]Checker
}

func NewHealth(checks map[string]Checker) *Health {
	return &Health{checks: checks}
}

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]string, len(h.checks))
		healthy = true
	)

	for name, check := range h.checks {
		wg.Add(1)
		go func(name string, check Checker) {
			defer wg.Done()
			err := check(ctx)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[name] = err.Error()
				healthy = false
				return
			}
			results[name] = "ok"
		}(name, check)
	}
	wg.Wait()

	resp := healthResponse{Status: "ok", Checks: results}
	code := http.StatusOK
	if !healthy {
		resp.Status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}