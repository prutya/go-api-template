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
	"prutya/go-api-template/internal/handlers/sessions"
	"prutya/go-api-template/internal/handlers/timeout_check"
	"prutya/go-api-template/internal/handlers/users"
	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

type Router struct {
	mux *chi.Mux
}

func NewRouter(
	config *config.Config,
	logger *zap.Logger,
	authenticationService authentication_service.AuthenticationService,
) *Router {
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

	// Route-specific middleware
	authenticationMiddleware := utils.NewAuthenticationMiddleware(authenticationService)

	// API routes
	mux.Post("/echo", echo.NewHandler())
	mux.Get("/health", health.NewHandler())
	mux.Get("/panic-check", panic_check.NewHandler())
	mux.Mount("/sessions", newSessionsMux(config, authenticationMiddleware, authenticationService))
	mux.Get("/timeout-check", timeout_check.NewHandler())
	mux.Mount("/users", newUsersMux(config, authenticationMiddleware))

	return &Router{mux: mux}
}

func generateRequestID(r *http.Request) (string, error) {
	val, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return val.String(), nil
}

func newSessionsMux(
	config *config.Config,
	authenticationMiddleware func(next http.Handler) http.Handler,
	authenticationService authentication_service.AuthenticationService,
) *chi.Mux {
	m := chi.NewRouter()

	m.Post("/", sessions.NewCreateHandler(config, authenticationService))
	m.Mount("/current", newSessionsCurrentMux(config, authenticationMiddleware, authenticationService))

	return m
}

func newSessionsCurrentMux(
	config *config.Config,
	authenticationMiddleware func(next http.Handler) http.Handler,
	authenticationService authentication_service.AuthenticationService,
) *chi.Mux {
	m := chi.NewRouter()

	m.Use(authenticationMiddleware)

	m.Delete("/", sessions.NewDeleteCurrentHandler(config, authenticationService))

	return m
}

func newUsersMux(config *config.Config, authenticationMiddleware func(next http.Handler) http.Handler) *chi.Mux {
	m := chi.NewRouter()

	m.Mount("/current", newUsersCurrentMux(config, authenticationMiddleware))

	return m
}

func newUsersCurrentMux(config *config.Config, authenticationMiddleware func(next http.Handler) http.Handler) *chi.Mux {
	m := chi.NewRouter()

	m.Use(authenticationMiddleware)

	m.Get("/", users.NewShowCurrentHandler(config))

	return m
}
