package parser

import (
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	taskmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/task"
)

func ParseManager(dir *string) (manager.Manager, error) {
	managers := []manager.Manager{}

	parseConfig := manager.ManagerParseConfig{
		CurrentDir: *dir,
		RootDir:    findClosestGitDir(dir),
	}

	jsWorkspace, err := jsmanager.ParseJsWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}

	directories := parseConfig.GetDirectories()
	for _, dir := range directories {
		manager, err := taskmanager.ParseTaskManager(&dir)
		if err != nil {
			logger.Warning(err.Error())
		} else if manager != nil {
			managers = append(managers, manager)
		}

		if jsWorkspace == nil {
			continue
		}

		jsManager, err := jsmanager.ParseJsManager(&dir, *jsWorkspace)
		if err != nil {
			logger.Warning(err.Error())
		} else if jsManager != nil {
			managers = append(managers, jsManager)
		}
	}

	if len(managers) > 0 {
		return managers[0], nil
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
