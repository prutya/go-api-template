package tasks

import (
	"prutya/go-api-template/internal/config"

	"github.com/hibiken/asynq"
)

func NewClient(config *config.Config) *asynq.Client {
	// TODO: More configuration options
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.TasksRedisHost,
		Password: config.TasksRedisPassword,
	})

	return client
}
