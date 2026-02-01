package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-gin/internal/infrastructure/logger"
)

type Server struct {
	httpServer *http.Server
	log        *logger.Logger
}

func New(addr string, handler http.Handler, log *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

func (s *Server) Run() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		s.log.Info("starting server", "addr", s.httpServer.Addr)
		serverErr <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdown:
		s.log.Info("shutdown signal received", "signal", sig.String())
		return s.Shutdown()
	}
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.log.Info("initiating graceful shutdown")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("graceful shutdown failed", "error", err)
		return s.httpServer.Close()
	}

	s.log.Info("server stopped gracefully")
	return nil
}
