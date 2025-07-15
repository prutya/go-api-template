package main

import (
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
		app.AuthenticationService,
		app.TransactionalEmailService,
	)

	if err := tasksServer.Run(); err != nil {
		logger.FatalContext(ctx, "Worker start error", "error", err)
	}

	logger.InfoContext(ctx, "Bye!")
}
