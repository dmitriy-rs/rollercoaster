package parser

import (
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	configfile "github.com/dmitriy-rs/rollercoaster/internal/manager/config-file"
	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	taskmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/task-manager"
)

type ParseManagerConfig struct {
	DefaultJSManager string
}

// managerResult holds the result of parsing a directory
type managerResult struct {
	managers []manager.Manager
	dirIndex int // To maintain order
}

func ParseManager(dir *string, config *ParseManagerConfig) ([]manager.Manager, error) {
	parseConfig := createParseConfig(dir)
	jsWorkspace := parseJSWorkspace(&parseConfig.RootDir, config.DefaultJSManager)
	directories := parseConfig.GetDirectories()

	if len(directories) == 0 {
		return handleNoDirectories(jsWorkspace), nil
	}

	managers := parseDirectoriesConcurrently(directories, jsWorkspace)
	return assembleManagersList(managers, jsWorkspace), nil
}

// createParseConfig creates the configuration for parsing
func createParseConfig(dir *string) configfile.ParseConfig {
	return configfile.ParseConfig{
		CurrentDir: *dir,
		RootDir:    findClosestGitDir(dir),
	}
}

// parseJSWorkspace attempts to parse the JS workspace
func parseJSWorkspace(rootDir *string, defaultJSManager string) *jsmanager.JsWorkspace {
	jsWorkspace, err := jsmanager.ParseJsWorkspace(rootDir, defaultJSManager)
	if err != nil {
		logger.Warning(err.Error())
	}
	return jsWorkspace
}

// handleNoDirectories returns appropriate managers when no directories are found
func handleNoDirectories(jsWorkspace *jsmanager.JsWorkspace) []manager.Manager {
	if jsWorkspace != nil {
		return []manager.Manager{&jsmanager.JsWorkspaceManager{Workspace: jsWorkspace}}
	}
	return nil
}

// parseDirectoriesConcurrently parses all directories in parallel
func parseDirectoriesConcurrently(directories []string, jsWorkspace *jsmanager.JsWorkspace) []manager.Manager {
	numDirs := len(directories)
	results := make(chan managerResult, numDirs)
	var wg sync.WaitGroup

	// Launch concurrent parsers for each directory
	for i, directory := range directories {
		wg.Add(1)
		go parseDirectoryAsync(i, directory, jsWorkspace, results, &wg)
	}

	// Close results channel when all parsing completes
	go func() {
		wg.Wait()
		close(results)
	}()

	return collectOrderedResults(results, numDirs)
}

// parseDirectoryAsync parses a single directory for managers
func parseDirectoryAsync(dirIndex int, dir string, jsWorkspace *jsmanager.JsWorkspace, results chan<- managerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	managers := parseDirectoryManagers(dir, jsWorkspace)

	results <- managerResult{
		managers: managers,
		dirIndex: dirIndex,
	}
}

// parseDirectoryManagers parses both JS and task managers for a directory
func parseDirectoryManagers(dir string, jsWorkspace *jsmanager.JsWorkspace) []manager.Manager {
	var managers []manager.Manager

	// Parse JS manager if workspace exists
	if jsWorkspace != nil {
		if jsManager := parseJSManager(dir, jsWorkspace); jsManager != nil {
			managers = append(managers, jsManager)
		}
	}

	// Parse task manager
	if taskMgr := parseTaskManager(dir); taskMgr != nil {
		managers = append(managers, taskMgr)
	}

	return managers
}

// parseJSManager attempts to parse a JS manager for the directory
func parseJSManager(dir string, jsWorkspace *jsmanager.JsWorkspace) manager.Manager {
	jsManager, err := jsmanager.ParseJsManager(&dir, jsWorkspace)
	if err != nil {
		logger.Warning(err.Error())
		return nil
	}
	if jsManager == nil {
		return nil
	}
	return jsManager
}

// parseTaskManager attempts to parse a task manager for the directory
func parseTaskManager(dir string) manager.Manager {
	taskMgr, err := taskmanager.ParseTaskManager(&dir)
	if err != nil {
		logger.Warning(err.Error())
		return nil
	}
	if taskMgr == nil {
		return nil
	}
	return taskMgr
}

// collectOrderedResults collects results from concurrent parsing and maintains order
func collectOrderedResults(results <-chan managerResult, numDirs int) []manager.Manager {
	orderedResults := make([][]manager.Manager, numDirs)
	for result := range results {
		orderedResults[result.dirIndex] = result.managers
	}

	// Flatten results while maintaining order
	var managers []manager.Manager
	for _, dirManagers := range orderedResults {
		managers = append(managers, dirManagers...)
	}

	return managers
}

// assembleManagersList creates the final managers list with proper ordering
func assembleManagersList(managers []manager.Manager, jsWorkspace *jsmanager.JsWorkspace) []manager.Manager {
	if len(managers) == 0 && jsWorkspace == nil {
		logger.Warning("Could not find a task manager in the current directory or its parents")
		return nil
	}

	slices.Reverse(managers)

	if jsWorkspace != nil {
		managers = append(managers, &jsmanager.JsWorkspaceManager{
			Workspace: jsWorkspace,
		})
	}

	return managers
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
