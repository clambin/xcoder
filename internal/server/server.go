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

func New(addr, rootDir, profile string, removeConverted bool, logger *slog.Logger) (*Server, error) {
	r := requests.Requests{}
	scan, err := scanner.New(rootDir, profile, &r, logger.With(slog.String("component", "scanner")))
	if err != nil {
		return nil, err
	}
	s := Server{
		Scanner:   scan,
		Convertor: convertor.New(&r, removeConverted, logger.With("component", "processor")),
		logger:    logger,
	}
	s.HTTPServer = s.makeHTTPServer(addr)

	return &s, nil
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

func (s Server) makeHTTPServer(addr string) *http.Server {
	m := chi.NewMux()
	m.Get("/health", s.Health)
	m.Route("/convertor", s.Convertor.Router)
	return &http.Server{
		Addr:    addr,
		Handler: m,
	}
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
