package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"prutya/todo/internal/config"
	"prutya/todo/internal/handlers/ts"
	handlerutils "prutya/todo/internal/handlers/utils"
	"prutya/todo/internal/logger"
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

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlerutils.RenderJson(w, r, map[string]any{"message": "Hello World!"}, 200, nil)
	})

	router.Get("/ts", ts.NewHandler())

	server := &http.Server{Addr: ":3000", Handler: router}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
