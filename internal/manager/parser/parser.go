package parser

import (
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/task"
)

func ParseManager(dir *string) (manager.Manager, error) {
	manager, err := taskmanager.ParseTaskManager(dir)
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

	return taskmanager.ParseTaskManager(&gitDir)
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
