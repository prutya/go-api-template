package main

import (
	"os"

	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/server"
)

func main() {
	app := app.Initialize()
	cfg, ctx, logger := app.Essentials.Config, app.Essentials.Context, app.Essentials.Logger

	// Server
	router := server.NewRouter(cfg, logger, app.AuthenticationService, app.UserService)
	server := server.NewServer(cfg, router, logger)

	if err := server.Start(); err != nil {
		logger.ErrorContext(ctx, "Server error", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Bye!")
}
