package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	logger     *zap.Logger
}

func NewServer(config *config.Config, router *Router, logger *zap.Logger) *Server {
	httpServer := &http.Server{
		Addr:              config.ListenAddr,
		Handler:           router.mux,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		IdleTimeout:       config.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		config:     config,
		logger:     logger,
	}
}

func (s *Server) Start() error {
	// Listen to the SIGINT and SIGKILL and stop the server
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// This will run when the server is stopped
	go func() {
		// Wait for the OS signals
		<-shutdownCh

		s.logger.Info("Server is shutting down")

		// Prepare a shutdown context
		shutdownCtx, shutdownRelease := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)

		defer shutdownRelease()

		// Shutdown the server with a timeout to let it complete the processing
		// of ongoing requests
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Server stopped with an error", zap.Error(err))
		} else {
			s.logger.Info("Server stopped")
		}
	}()

	s.logger.Info("Server is starting", zap.String("addr", s.config.ListenAddr))

	// This blocks until the server is stopped
	err := s.httpServer.ListenAndServe()

	// This error is expected when the server is being stopped
	if errors.Is(err, http.ErrServerClosed) {
		s.logger.Info("Server stopped")

		return nil
	}

	return err
}
