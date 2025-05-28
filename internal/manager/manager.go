package manager

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Manager interface {
	GetTitle() string

	ListTasks() ([]task.Task, error)
	ExecuteTask(task *task.Task, args ...string)
}

func FindManager(dir *string) (Manager, error) {
	manager, err := ParseTaskManager(dir)
	if err != nil {
		return nil, err
	}
	if manager != nil {
		return manager, nil
	}

	gitDir := findClosestGitDir(dir)
	if gitDir == "" {
		logger.Warning("Could not find a task manager in the current directory or its parents")
		return nil, nil
	}

	return ParseTaskManager(&gitDir)
}

func findClosestGitDir(dir *string) string {
	if dir == nil || *dir == "" {
		return ""
	}
	currentDir := *dir
	for {
		gitPath := path.Join(currentDir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return currentDir
		}
		parentDir := path.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}
	return ""
}

func FindClosestTask(manager Manager, arg string) (*task.Task, error) {
	tasks, err := manager.ListTasks()
	if err != nil {
		logger.Error("Failed to list tasks", err)
		return nil, err
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
		var matchedTask *task.Task
		for _, task := range tasks {
			if task.Name == matches[0].Target {
				matchedTask = &task
				break
			}
		}
		return matchedTask, nil
	}
	return nil, fmt.Errorf("no task found for '%s'", arg)
}
