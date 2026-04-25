// internal/server/server.go

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panditvishnuu/userservice/internal/config"
	"github.com/panditvishnuu/userservice/internal/handler"
	"github.com/panditvishnuu/userservice/internal/httpx"
	"github.com/panditvishnuu/userservice/internal/middleware"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

func New(cfg *config.Config, userHandler *handler.UserHandler) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux, userHandler)

	// Middleware is applied outside-in. RequestID runs first so every
	// subsequent middleware and handler has access to the request ID.
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      h,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		cfg: cfg,
	}
}

func registerRoutes(mux *http.ServeMux, u *handler.UserHandler) {
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/v1/users/register", u.Register)
	mux.HandleFunc("POST /api/v1/users/login", u.Login)
	mux.HandleFunc("GET /api/v1/users/me", u.GetMe)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Run starts the server and blocks until a shutdown signal is received.
// It handles the full lifecycle: start, signal wait, graceful drain, exit.
func (s *Server) Run() error {
	// Channel to receive OS signals. Buffered by 1 — if the signal arrives
	// before we call Notify, it won't be dropped.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start serving in a goroutine so we can wait for the signal below.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Block until signal or server error.
	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	// Give in-flight requests up to ShutdownTimeout seconds to complete.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(s.cfg.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("server stopped cleanly")
	return nil
}
