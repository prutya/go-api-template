package tasks_server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"

	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/transactional_email_service"
	"prutya/go-api-template/internal/tasks"
)

type sendPasswordResetEmailTaskHandler struct {
	authenticationService authentication_service.AuthenticationService
}

func newSendPasswordResetEmailTaskHandler(authenticationService authentication_service.AuthenticationService) *sendPasswordResetEmailTaskHandler {
	return &sendPasswordResetEmailTaskHandler{
		authenticationService: authenticationService,
	}
}

func (h *sendPasswordResetEmailTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendPasswordResetEmailPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	if err := h.authenticationService.SendPasswordResetEmail(ctx, payload.UserID); err != nil {
		if errors.Is(err, authentication_service.ErrUserNotFound) {
			return asynq.RevokeTask
		}

		if errors.Is(err, authentication_service.ErrPasswordResetRateLimited) {
			return asynq.RevokeTask
		}

		if errors.Is(err, transactional_email_service.ErrGlobalLimitReached) {
			return asynq.SkipRetry
		}

		return err
	}

	return nil
}
