package main

import (
	"os"

	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/tasks_scheduler"
)

func main() {
	appEssentials := app.InitializeEssentials()
	ctx, cfg, logger := appEssentials.Context, appEssentials.Config, appEssentials.Logger

	// Tasks server
	scheduler, err := tasks_scheduler.NewScheduler(
		ctx,
		cfg.TasksRedisAddr,
		cfg.TasksRedisPassword,
	)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create scheduler", "error", err)
		os.Exit(1)
	}

	if err := scheduler.Run(); err != nil {
		logger.ErrorContext(ctx, "Scheduler start error", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Bye!")
}
