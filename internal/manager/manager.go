package manager

import (
	"fmt"
	"sort"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Manager interface {
	GetTitle() string

	ListTasks() ([]task.Task, error)
	ExecuteTask(name string)
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

	taskNames := make([]string, len(tasks))
	for _, t := range tasks {
		taskNames = append(taskNames, t.Name)
	}

	matches := fuzzy.RankFind(arg, taskNames)
	sort.Sort(matches)

	logger.Debug(fmt.Sprintf("Fuzzy matches for '%s': %v", arg, matches))

	if len(matches) != 0 {
		manager.ExecuteTask(matches[0].Target)
		return nil
	}
	return fmt.Errorf("no task found for '%s'", arg)
}
