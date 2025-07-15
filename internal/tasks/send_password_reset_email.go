package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeSendPasswordResetEmail = "send_password_reset_email"

type SendPasswordResetEmailPayload struct {
	UserID string
}

func NewSendPasswordResetEmailTask(userID string) (*Task, error) {
	payload, err := json.Marshal(SendPasswordResetEmailPayload{
		UserID: userID,
	})

	if err != nil {
		return nil, err
	}

	return NewTask(asynq.NewTask(TypeSendPasswordResetEmail, payload)), nil
}
