package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeUserHello = "user_hello"

type UserHelloPayload struct {
	UserID string
}

func NewUserHelloTask(userID string) (*Task, error) {
	payload, err := json.Marshal(UserHelloPayload{
		UserID: userID,
	})

	if err != nil {
		return nil, err
	}

	return NewTask(asynq.NewTask(TypeUserHello, payload)), nil
}
