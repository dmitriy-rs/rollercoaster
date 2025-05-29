package parser

import (
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/task"
)

func ParseManager(dir *string) (manager.Manager, error) {
	parseConfig := manager.ManagerParseConfig{
		CurrentDir: *dir,
		RootDir:    findClosestGitDir(dir),
	}

	manager, err := taskmanager.ParseTaskManager(dir)
	if err != nil {
		return nil, err
	}
	if manager != nil {
		return manager, nil
	}

		logger.Warning("Could not find a task manager in the current directory or its parents")
		return nil, nil
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
