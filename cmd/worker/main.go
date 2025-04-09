package main

import (
	"go.uber.org/zap"

	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/tasks_server"
)

func main() {
	app := app.Initialize()

	// Tasks server
	tasksServer := tasks_server.NewServer(
		app.Context,
		app.Config.TasksRedisAddr,
		app.Config.TasksRedisPassword,
		app.UserService,
	)

	if err := tasksServer.Run(); err != nil {
		app.Logger.Fatal("Tasks server error", zap.Error(err))
	}

	app.Logger.Info("Bye!")
}
