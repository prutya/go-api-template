package main

import (
	"go.uber.org/zap"

	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/server"
)

func main() {
	app := app.Initialize()

	// Server
	router := server.NewRouter(app.Config, app.Logger, app.AuthenticationService, app.UserService)
	server := server.NewServer(app.Config, router, app.Logger)

	if err := server.Start(); err != nil {
		app.Logger.Fatal("Server error", zap.Error(err))
	}

	app.Logger.Info("Bye!")
}
