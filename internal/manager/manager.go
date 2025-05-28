package manager

import (
	"fmt"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

type Manager interface {
	GetTitle() string

	ListTasks() ([]task.Task, error)
	ExecuteTask(name task.Task)
}

func FindManager(dir *string) (Manager, error) {
	manager, err := ParseTaskManager(dir)
	if err != nil {
		return nil, err
	}
	if manager != nil {
		return manager, nil
	}
	return nil, nil
}

func ExecuteClosestTask(manager Manager, arg string) error {
	tasks, err := manager.ListTasks()
	if err != nil {
		logger.Error("Failed to list tasks", err)
		return err
	}

	logger.Debug(fmt.Sprintf("Found tasks: %s", tasks))

	if len(tasks) != 0 {
		manager.ExecuteTask(tasks[0])
	}
	return nil
}
