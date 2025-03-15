package task_check

import (
	"encoding/json"
	"net/http"

	"github.com/hibiken/asynq"

	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/tasks"
)

type Request struct {
	Message string `json:"message" validate:"required,gte=1,lte=255"`
}

type Response struct {
	TaskID string `json:"task_id"`
}

// TODO: Test
func NewHandler(tasksClient *asynq.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request
		reqBody := Request{}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&reqBody); err != nil {
			utils.RenderInvalidJsonError(w, r)
			return
		}

		// Validate the request
		if err := utils.Validate.Struct(reqBody); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Create a new demo task
		demoTask, err := tasks.NewDemoTask(reqBody.Message)
		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		demoTaskInfo, err := tasksClient.EnqueueContext(r.Context(), demoTask)
		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		responseBody := &Response{TaskID: demoTaskInfo.ID}

		utils.RenderJson(w, r, responseBody, http.StatusOK, nil)
	}
}
