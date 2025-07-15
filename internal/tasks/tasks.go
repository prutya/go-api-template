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
}

func NewTaskInfo(asynqTaskInfo *asynq.TaskInfo) *TaskInfo {
	return &TaskInfo{
		asynqTaskInfo: asynqTaskInfo,
	}
}
