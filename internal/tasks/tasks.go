package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"prutya/go-api-template/internal/logger"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TypeDemo = "demo"

type DemoPayload struct {
	Message string `json:"message"`
}

func NewDemoTask(message string) (*asynq.Task, error) {
	payload, err := json.Marshal(DemoPayload{Message: message})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeDemo, payload), nil
}

func HandleDemoTask(ctx context.Context, task *asynq.Task) error {
	logger := logger.MustFromContext(ctx)

	var payload DemoPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	logger.Info("Processing task", zap.String("message", payload.Message))

	// Imitate some work
	time.Sleep(10 * time.Second)

	logger.Info("Task done")

	return nil
}
