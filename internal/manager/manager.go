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

type ManagerTask struct {
	task.Task
	Manager *Manager
}

func GetManagerTasksFromList(managers []Manager) ([]ManagerTask, error) {
	allTasks := []ManagerTask{}
	for _, manager := range managers {
		tasks, err := getManagerTasks(manager)
		if err != nil {
			logger.Warning("Failed to get tasks for manager: " + manager.GetTitle().Name)
			continue
		}
		allTasks = append(allTasks, tasks...)
	}
	return allTasks, nil
}

func getManagerTasks(manager Manager) ([]ManagerTask, error) {
	tasks, err := manager.ListTasks()
	if err != nil {
		return nil, err
	}
	task.SortTasks(tasks)

	taskWithManager := make([]ManagerTask, len(tasks))
	for i, t := range tasks {
		taskWithManager[i] = ManagerTask{
			Task:    t,
			Manager: &manager,
		}
	}
	return taskWithManager, nil
}

func FindAllClosestTasksFromList(managers []Manager, arg string) (Manager, *task.Task, error) {
	for _, manager := range managers {
		task, _ := FindClosestTask(manager, arg)
		if task != nil {
			return manager, task, nil
		}
	}
	return nil, nil, fmt.Errorf("no task found for '%s'", arg)
}
