package tasks_client

import (
	"context"

	"github.com/hibiken/asynq"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/tasks"
)

type Client interface {
	Ping() error
	Enqueue(ctx context.Context, task *tasks.Task) (*tasks.TaskInfo, error)
	Close() error
}

type client struct {
	asynqClient *asynq.Client
}

func NewClient(redisAddr string, redisPassword string) Client {
	// TODO: Support more configuration options
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &client{
		asynqClient: asynqClient,
	}
}

func (c *client) Ping() error {
	return c.asynqClient.Ping()
}

func (c *client) Enqueue(ctx context.Context, task *tasks.Task) (*tasks.TaskInfo, error) {
	logger := logger.MustFromContext(ctx)
	taskType := task.AsynqTask.Type()

	logger.DebugContext(ctx, "Enqueueing task", "task_type", taskType)

	asynqTaskInfo, err := c.asynqClient.EnqueueContext(ctx, task.AsynqTask)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Enqueued task", "task_id", asynqTaskInfo.ID, "task_type", asynqTaskInfo.Type)

	return tasks.NewTaskInfo(asynqTaskInfo), nil
}

func (c *client) Close() error {
	return c.asynqClient.Close()
}
