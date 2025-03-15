package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/echo"
	"prutya/go-api-template/internal/handlers/health"
	"prutya/go-api-template/internal/handlers/panic_check"
	"prutya/go-api-template/internal/handlers/timeout_check"
	"prutya/go-api-template/internal/handlers/utils"
)

type Router struct {
	mux *chi.Mux
}

func NewRouter(config *config.Config, logger *zap.Logger) *Router {
	mux := chi.NewRouter()

	// Middleware
	mux.Use(utils.NewRequestIDMiddleware(generateRequestID))
	mux.Use(middleware.RealIP)
	mux.Use(utils.NewLoggerMiddleware(logger))
	mux.Use(utils.NewRecoverMiddleware())
	mux.Use(utils.NewTimeoutMiddleware(config.RequestTimeout))
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.CorsAllowedOrigins,
		AllowedMethods:   config.CorsAllowedMethods,
		AllowedHeaders:   config.CorsAllowedHeaders,
		ExposedHeaders:   config.CorsExposedHeaders,
		AllowCredentials: config.CorsAllowCredentials,
		MaxAge:           int(config.CorsMaxAge.Seconds()),
	}))

	// Handle 404 Not Found and render a custom response
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		utils.RenderError(w, r, utils.ErrNotFound)
	})

	// Handle 405 Method Not Allowed and render a custom response
	mux.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		utils.RenderError(w, r, utils.ErrMethodNotAllowed)
	})

	// API routes
	mux.Get("/health", health.NewHandler())
	mux.Post("/echo", echo.NewHandler())
	mux.Get("/panic-check", panic_check.NewHandler())
	mux.Get("/timeout-check", timeout_check.NewHandler())

	return &Router{mux: mux}
}

func generateRequestID(r *http.Request) (string, error) {
	val, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return val.String(), nil
}
