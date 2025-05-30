package parser

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/dmitriy-rs/rollercoaster/internal/config"
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	taskmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/task-manager"
)

func ParseManager(dir *string) ([]manager.Manager, error) {
	managers := []manager.Manager{}

	parseConfig := config.ParseConfig{
		CurrentDir: *dir,
		RootDir:    findClosestGitDir(dir),
	}

	jsWorkspace, err := jsmanager.ParseJsWorkspace(&parseConfig.RootDir)
	if err != nil {
		logger.Warning(err.Error())
	}

	directories := parseConfig.GetDirectories()
	for _, dir := range directories {
		if jsWorkspace != nil {
			jsManager, err := jsmanager.ParseJsManager(&dir, jsWorkspace)
			if err != nil {
				logger.Warning(err.Error())
			} else if jsManager != nil {
				managers = append(managers, jsManager)
			}
		}

		manager, err := taskmanager.ParseTaskManager(&dir)
		if err != nil {
			logger.Warning(err.Error())
		} else if manager != nil {
			managers = append(managers, manager)
		}
	}

	if jsWorkspace != nil {
		managers = append(managers, &jsmanager.JsWorkspaceManager{
			Workspace: jsWorkspace,
		})
	}

	if len(managers) > 0 {
		slices.Reverse(managers)
		return managers, nil
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
		gitPath := filepath.Join(currentDir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}
	return *dir
}
