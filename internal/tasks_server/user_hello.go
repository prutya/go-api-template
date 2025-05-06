package tasks_server

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks"
)

type userHelloTaskHandler struct {
	userService user_service.UserService
}

func newUserHelloTaskHandler(userService user_service.UserService) *userHelloTaskHandler {
	return &userHelloTaskHandler{
		userService: userService,
	}
}

func (h *userHelloTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	logger := logger.MustFromContext(ctx)

	var payload tasks.UserHelloPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	user, err := h.userService.GetUserByID(ctx, payload.UserID)

	if err != nil {
		return err
	}

	// Simulate sending a hello message to the user
	// In a real application, you would replace this with actual logic
	logger.InfoContext(ctx, "Simulating hello message to user", "user_id", user.ID)

	return nil
}
