package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// Server wraps http.Server with graceful shutdown helpers.
type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

// New creates an HTTP server bound to addr.
func New(addr string, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: constant.HTTPReadHeaderTimeout,
			ReadTimeout:       constant.HTTPReadTimeout,
			WriteTimeout:      constant.HTTPWriteTimeout,
			IdleTimeout:       constant.HTTPIdleTimeout,
		},
		log: log,
	}
}

// ListenAndServe starts the server and blocks until it stops.
func (s *Server) ListenAndServe() error {
	s.log.Info("http server listening", constant.LogAttrAddr, s.httpServer.Addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
