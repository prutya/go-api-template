package main

import (
	"prutya/go-api-template/internal/app"
	"prutya/go-api-template/internal/tasks_scheduler"
)

func main() {
	appEssentials := app.NewAppEssentials()
	ctx, cfg, logger := appEssentials.Context, appEssentials.Config, appEssentials.Logger

	// Tasks server
	scheduler, err := tasks_scheduler.NewScheduler(
		ctx,
		cfg.TasksRedisAddr,
		cfg.TasksRedisPassword,
	)
	if err != nil {
		logger.FatalContext(ctx, "Failed to create scheduler", "error", err)
	}

	if err := scheduler.Run(); err != nil {
		logger.FatalContext(ctx, "Scheduler start error", "error", err)
	}

	logger.InfoContext(ctx, "Bye!")
}
