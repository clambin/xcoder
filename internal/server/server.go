package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/clambin/videoConvertor/internal/server/convertor"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/internal/server/scanner"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type Server struct {
	Scanner    *scanner.Scanner
	Convertor  *convertor.Convertor
	HTTPServer *http.Server
	logger     *slog.Logger
}

type Config struct {
	Addr            string
	ScannerConfig   scanner.Config
	RemoveConverted bool
}

func New(cfg Config, logger *slog.Logger) (*Server, error) {
	r := requests.Requests{}
	s, err := scanner.New(cfg.ScannerConfig, &r, logger.With(slog.String("component", "scanner")))
	if err != nil {
		return nil, err
	}
	server := Server{
		Scanner:    s,
		Convertor:  convertor.New(cfg.RemoveConverted, &r, logger.With("component", "processor")),
		HTTPServer: &http.Server{Addr: cfg.Addr},
		logger:     logger,
	}
	m := chi.NewMux()
	m.Get("/health", server.Health)
	m.Route("/convertor", server.Convertor.Router)
	server.HTTPServer.Handler = m

	return &server, nil
}

func (s Server) Run(ctx context.Context) error {
	go func() {
		if err := s.HTTPServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to start HTTP server", slog.Any("err", err))
			panic(err)
		}
	}()

	errConvertor := make(chan error)
	go func() {
		errConvertor <- s.Convertor.Run(ctx, 2)
	}()

	errScanner := s.Scanner.Run(ctx, 4)

	errs := errors.Join(<-errConvertor, errScanner)
	if errs != nil {
		s.logger.Error("failed to start server", slog.Any("err", errs))
	}
	_ = s.HTTPServer.Shutdown(context.Background())
	return errs
}

type HealthChecker interface {
	Health(ctx context.Context) any
}

func (s Server) Health(w http.ResponseWriter, r *http.Request) {
	healthItems := make(map[string]any)
	if f, ok := any(s.Convertor).(HealthChecker); ok {
		healthItems["Convertor"] = f.Health(r.Context())
	}
	if f, ok := any(s.Scanner).(HealthChecker); ok {
		healthItems["Scanner"] = f.Health(r.Context())
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(healthItems); err != nil {
		s.logger.Error("failed to write health", slog.Any("err", err))
	}
}
