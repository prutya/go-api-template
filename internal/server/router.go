package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid/v5"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/handlers/account"
	"prutya/go-api-template/internal/handlers/account/sessions"
	"prutya/go-api-template/internal/handlers/users"
	"prutya/go-api-template/internal/handlers/utils"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/captcha_service"
	"prutya/go-api-template/internal/services/transactional_email_service"
	"prutya/go-api-template/internal/services/user_service"
)

type Router struct {
	mux *chi.Mux
}

func NewRouter(
	config *config.Config,
	logger *loggerpkg.Logger,
	authenticationService authentication_service.AuthenticationService,
	userService user_service.UserService,
	transactionalEmailService transactional_email_service.TransactionalEmailService,
	captchaService captcha_service.CaptchaService,
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

	captchaCheckMiddleware := utils.NewCaptchaCheckMiddleware(captchaService)
	authenticationMiddleware := utils.NewAuthenticationMiddleware(authenticationService)

	// NOTE: Use this in the routes that require email verification
	// emailVerificationCheckMiddleware := utils.NewEmailVerificationCheckMiddleware(authenticationService)

	// API routes

	// /account

	mux.Route("/account", func(r chi.Router) {
		r.Post("/refresh-session", account.NewRefreshSessionHandler(config, authenticationService))
		r.Post("/verify-email", account.NewVerifyEmailHandler(config, authenticationService))
		r.Post("/reset-password", account.NewResetPasswordHandler(config, authenticationService))

		r.Group(func(r chi.Router) {
			r.Use(captchaCheckMiddleware)

			r.Post("/login", account.NewLoginHandler(config, authenticationService))
			r.Post("/register", account.NewRegisterHandler(authenticationService))
			r.Post("/request-email-verification", account.NewRequestEmailVerificationHandler(authenticationService))
			r.Post("/request-password-reset", account.NewRequestPasswordResetHandler(authenticationService))
			r.Post("/verify-password-reset-otp", account.NewVerifyPasswordResetOTPHandler(config, authenticationService))
		})

		r.Group(func(r chi.Router) {
			r.Use(authenticationMiddleware)

			r.Post("/logout", account.NewLogoutHandler(config, authenticationService))
			r.Post("/change-password", account.NewChangePasswordHandler(config, authenticationService))
			r.Post("/delete-account", account.NewDeleteAccountHandler(config, authenticationService))

			r.Route("/sessions", func(r chi.Router) {
				r.Get("/", sessions.NewSessionsListHandler(authenticationService))
				r.Delete("/{sessionID}", sessions.NewSessionsTerminateHandler(config, authenticationService))
			})
		})
	})

	// /users

	mux.Route("/users", func(r chi.Router) {
		r.Use(authenticationMiddleware)

		r.Get("/current", users.NewCurrentHandler(userService))
	})

	return &Router{mux: mux}
}

func generateRequestID(r *http.Request) (string, error) {
	val, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	return val.String(), nil
}
