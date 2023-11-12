package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Health struct {
	Components []HealthChecker
	Addr       string
}

type HealthChecker interface {
	Health(ctx context.Context) any
}

func (h *Health) Run(ctx context.Context) error {
	r := http.NewServeMux()
	r.Handle("/health", h)

	s := http.Server{Addr: h.Addr, Handler: r}

	ch := make(chan error)
	go func() { ch <- s.ListenAndServe() }()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)

	err := <-ch
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := make([]any, len(h.Components))
	for i := range h.Components {
		response[i] = h.Components[i].Health(r.Context())
	}
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
