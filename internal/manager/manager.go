package manager

import (
	"fmt"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	fuzzy "github.com/sahilm/fuzzy"
)

type Manager interface {
	GetTitle() Title

	ListTasks() ([]task.Task, error)
	ExecuteTask(task *task.Task, args ...string)
}

type Title struct {
	Name        string
	Description string
}

func FindClosestTaskFromList(managers []Manager, arg string) (Manager, *task.Task, error) {
	for _, manager := range managers {
		task, _ := FindClosestTask(manager, arg)
		if task != nil {
			return manager, task, nil
		}
	}
	return nil, nil, fmt.Errorf("no task found for '%s'", arg)
}

func FindClosestTask(manager Manager, arg string) (*task.Task, error) {
	tasks, err := manager.ListTasks()
	if err != nil {
		logger.Error("Failed to list tasks", err)
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Found tasks: %s", tasks))

	taskNames := make([]string, len(tasks))
	for i, t := range tasks {
		taskNames[i] = t.Name
	}

	matches := fuzzy.Find(arg, taskNames)

	logger.Debug(fmt.Sprintf("Fuzzy matches for '%s': %v", arg, matches))

	if len(matches) != 0 {
		var matchedTask *task.Task
		for _, task := range tasks {
			if task.Name == matches[0].Str {
				matchedTask = &task
				break
			}
		}
		return matchedTask, nil
	}
	return nil, fmt.Errorf("no task found for '%s'", arg)
}
