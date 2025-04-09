package tasks

import "github.com/hibiken/asynq"

type Task struct {
	AsynqTask *asynq.Task
}

func NewTask(t *asynq.Task) *Task {
	return &Task{
		AsynqTask: t,
	}
}

type TaskInfo struct {
	asynqTaskInfo *asynq.TaskInfo

	ID string

	// There are more fields in asynq.TaskInfo struct that can be made public
	// if necessary
}

func NewTaskInfo(asynqTaskInfo *asynq.TaskInfo) *TaskInfo {
	return &TaskInfo{
		asynqTaskInfo: asynqTaskInfo,
		ID:            asynqTaskInfo.ID,
	}
}
