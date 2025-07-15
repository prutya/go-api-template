package tasks_scheduler

import (
	"context"

	"github.com/hibiken/asynq"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/tasks"
)

type Scheduler interface {
	Run() error
}

type scheduler struct {
	asynqScheduler *asynq.Scheduler
}

func NewScheduler(
	baseCtx context.Context,
	redisAddr string,
	redisPassword string,
) (Scheduler, error) {
	logger := logger.MustFromContext(baseCtx)

	asynqScheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		&asynq.SchedulerOpts{
			Logger: tasks.NewSlogLoggerAdapter(logger),
		},
	)

	// Cleanup email send attempts ever day at 02:00
	if _, err := asynqScheduler.Register(
		"0 2 * * *",
		asynq.NewTask(tasks.TypeCleanupEmailSendAttempts, nil),
	); err != nil {
		return nil, err
	}

	return &scheduler{
		asynqScheduler: asynqScheduler,
	}, nil
}

func (s *scheduler) Run() error {
	return s.asynqScheduler.Run()
}
