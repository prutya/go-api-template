package main

import (
	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/server"
)

func main() {
	app := app.NewApp()
	cfg, ctx, logger := app.Essentials.Config, app.Essentials.Context, app.Essentials.Logger

	server := server.NewServer(
		cfg,
		server.NewRouter(
			cfg,
			logger,
			app.AuthenticationService,
			app.UserService,
			app.TransactionalEmailService,
			app.CaptchaService,
		),
		logger,
	)

	if err := server.Start(); err != nil {
		logger.FatalContext(ctx, "Server error", "error", err)
	}

	logger.InfoContext(ctx, "Bye!")
}
