package main

import (
	"os"
	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/tasks_server"
)

func main() {
	app := app.Initialize()
	ctx, cfg, logger := app.Essentials.Context, app.Essentials.Config, app.Essentials.Logger

	// Tasks server
	tasksServer := tasks_server.NewServer(
		ctx,
		cfg.TasksRedisAddr,
		cfg.TasksRedisPassword,
		app.UserService,
	)

	if err := tasksServer.Run(); err != nil {
		logger.ErrorContext(ctx, "Tasks server error", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Bye!")
}
