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

func FindClosestTaskFromList(managers []Manager, arg string) (*ManagerTask, error) {
	for _, manager := range managers {
		task, _ := FindClosestTask(manager, arg)
		if task != nil {
			return task, nil
		}
	}
	return nil, fmt.Errorf("no task found for '%s'", arg)
}

func FindClosestTask(manager Manager, arg string) (*ManagerTask, error) {
	tasks, err := manager.ListTasks()
	if err != nil {
		logger.Error("Failed to list tasks", err)
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Found tasks: %s", tasks))

	matches := fuzzy.FindFrom(arg, task.TaskSource(tasks))

	logger.Debug(fmt.Sprintf("Fuzzy matches for '%s': %v", arg, matches))

	if len(matches) != 0 {
		return &ManagerTask{
			Task:    tasks[matches[0].Index],
			Manager: &manager,
		}, nil
	}
	return nil, fmt.Errorf("no task found for '%s'", arg)
}

func FindAllClosestTasksFromList(managers []Manager, arg string) ([]ManagerTask, error) {
	tasks, err := GetManagerTasksFromList(managers)
	if err != nil {
		return nil, err
	}

	matches := fuzzy.FindFrom(arg, ManagerTaskSource(tasks))

	logger.Debug(fmt.Sprintf("Fuzzy matches for '%s': %v", arg, matches))

	result := make([]ManagerTask, len(matches))
	for i, match := range matches {
		result[i] = tasks[match.Index]
	}

	return result, nil
}

type ManagerTask struct {
	task.Task
	Manager *Manager
}

type ManagerTaskSource []ManagerTask

func (mts ManagerTaskSource) String(i int) string {
	return mts[i].Name
}

func (mts ManagerTaskSource) Len() int {
	return len(mts)
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
