package main

import (
	"errors"
	"net/http"
	"prutya/todo/internal/handlers/middleware"
	"prutya/todo/internal/logger"

	"prutya/todo/internal/config"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	// Initialize the config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize the logger
	logger, err := logger.New(cfg.LogLevel, cfg.LogTimeFormat)
	if err != nil {
		panic(err)
	}
	logger.Debug("Logger OK")

	// Initialize the router
	router := chi.NewRouter()

	// Initialize the Request ID middleware
	router.Use(middleware.NewRequestID(func(r *http.Request) (string, error) {
		val, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}

		return val.String(), nil
	}))

	// TODO: Logger
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(cfg.RequestTimeout))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// requestId := r.Context().Value(middleware.RequestIdContextKey)
		// fmt.Println(requestId)
		w.Write([]byte("Hello World!"))
	})

	server := &http.Server{Addr: ":3000", Handler: router}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
