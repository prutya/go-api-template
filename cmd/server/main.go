package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
	dbpkg "prutya/go-api-template/internal/db"
	"prutya/go-api-template/internal/handlers/echo"
	"prutya/go-api-template/internal/handlers/health"
	handlerutils "prutya/go-api-template/internal/handlers/utils"
	loggerpkg "prutya/go-api-template/internal/logger"
)

func main() {
	// Initialize the config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize the logger
	logger, err := loggerpkg.New(cfg.LogLevel, cfg.LogTimeFormat)
	if err != nil {
		panic(err)
	}
	logger.Info("Logger OK")

	// Initialize the main app context
	ctx := context.WithValue(context.Background(), loggerpkg.LoggerContextKey{}, logger)

	// Initialize the database connection
	db := dbpkg.New(cfg.DatabaseUrl)

	// Smoke-test the database connection
	if err := db.PingContext(ctx); err == nil {
		logger.Info("Database OK")
	} else {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}

	// Initialize the router
	router := chi.NewRouter()

	// Initialize the Request ID middleware
	router.Use(handlerutils.NewRequestIDMiddleware(func(r *http.Request) (string, error) {
		val, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}

		return val.String(), nil
	}))

	// Initialize the Real IP middleware
	router.Use(chimiddleware.RealIP)

	// Initialize the Logger middleware
	router.Use(handlerutils.NewLoggerMiddleware(logger))

	// Initialize the Recover middleware (to handle panics)
	router.Use(handlerutils.NewRecoverMiddleware())

	// Initialize the Timeout middleware
	router.Use(chimiddleware.Timeout(cfg.RequestTimeout))

	// Initialize the CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CorsAllowedOrigins,
		AllowedMethods:   cfg.CorsAllowedMethods,
		AllowedHeaders:   cfg.CorsAllowedHeaders,
		ExposedHeaders:   cfg.CorsExposedHeaders,
		AllowCredentials: cfg.CorsAllowCredentials,
		MaxAge:           int(cfg.CorsMaxAge.Seconds()),
	}))

	// Initialize the routes
	router.Get("/health", health.NewHandler())
	router.Post("/echo", echo.NewHandler())

	// Prepare the server
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	// Prepare channels for shutdown signals
	shutdownSignals := map[string]chan os.Signal{
		"http_shutdown": make(chan os.Signal, 1),
	}

	// Subscribe to shutdown signals from the OS
	for k := range shutdownSignals {
		signal.Notify(shutdownSignals[k], syscall.SIGINT, syscall.SIGTERM)
	}

	// Make sure to wait for every internal process to complete
	var cleanupWg sync.WaitGroup
	cleanupWg.Add(len(shutdownSignals))

	// HTTP shutdown goroutine
	go func() {
		defer cleanupWg.Done()

		// Wait for the OS signals
		<-shutdownSignals["http_shutdown"]

		// Prepare a shutdown context
		shutdownCtx, shutdownRelease := context.WithTimeout(
			context.Background(),
			cfg.ShutdownTimeout,
		)

		defer shutdownRelease()

		// Shutdown the server with a timeout to let it complete the processing
		// of existing requests
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("HTTP shutdown error", zap.Error(err))
		}

		logger.Info("HTTP shutdown complete")
	}()

	logger.Info("Listening", zap.String("addr", cfg.ListenAddr))

	// Run the server
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("HTTP server error", zap.Error(err))
	}

	logger.Info("HTTP shutdown started")

	// Wait for cleanups to complete
	cleanupWg.Wait()

	logger.Info("Bye!")
}
