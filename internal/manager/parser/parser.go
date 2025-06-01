package parser

import (
	"path/filepath"
	"slices"
	"sync"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
	configfile "github.com/dmitriy-rs/rollercoaster/internal/manager/config-file"
	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	taskmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/task-manager"
	"golang.org/x/sync/errgroup"
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

// parseDirectoryManagers parses both JS and task managers for a directory concurrently
func parseDirectoryManagers(dir string, jsWorkspace *jsmanager.JsWorkspace) []manager.Manager {
	// Pre-allocate slice with capacity 2 (typically JS + Task manager)
	managers := make([]manager.Manager, 0, 2)

	// Use errgroup for concurrent parsing with proper error handling
	var eg errgroup.Group
	var jsManager, taskMgr manager.Manager
	var jsErr, taskErr error

	// Parse JS manager concurrently if workspace exists
	if jsWorkspace != nil {
		eg.Go(func() error {
			jsManager, jsErr = parseJSManagerWithError(dir, jsWorkspace)
			return nil // Don't fail the group, just capture the error
		})
	}

	// Parse task manager concurrently
	eg.Go(func() error {
		taskMgr, taskErr = parseTaskManagerWithError(dir)
		return nil // Don't fail the group, just capture the error
	})

	// Wait for both operations to complete
	_ = eg.Wait()

	// Log any errors that occurred
	if jsErr != nil {
		logger.Warning(jsErr.Error())
	}
	if taskErr != nil {
		logger.Warning(taskErr.Error())
	}

	// Collect successful results
	if jsManager != nil {
		managers = append(managers, jsManager)
	}
	if taskMgr != nil {
		managers = append(managers, taskMgr)
	}

	return managers
}

// parseJSManagerWithError attempts to parse a JS manager and returns both result and error
func parseJSManagerWithError(dir string, jsWorkspace *jsmanager.JsWorkspace) (manager.Manager, error) {
	jsManager, err := jsmanager.ParseJsManager(&dir, jsWorkspace)
	if err != nil {
		return nil, err
	}
	if jsManager == nil {
		return nil, nil
	}
	return jsManager, nil
}

// parseTaskManagerWithError attempts to parse a task manager and returns both result and error
func parseTaskManagerWithError(dir string) (manager.Manager, error) {
	taskMgr, err := taskmanager.ParseTaskManager(&dir)
	if err != nil {
		return nil, err
	}
	if taskMgr == nil {
		return nil, nil
	}
	return taskMgr, nil
}

// collectOrderedResults collects results from concurrent parsing and maintains order
func collectOrderedResults(results <-chan managerResult, numDirs int) []manager.Manager {
	orderedResults := make([][]manager.Manager, numDirs)
	for result := range results {
		orderedResults[result.dirIndex] = result.managers
	}

	// Estimate total capacity needed (typically 2 managers per directory)
	estimatedCapacity := numDirs * 2
	managers := make([]manager.Manager, 0, estimatedCapacity)

	// Flatten results while maintaining order
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
		info, err := cache.DefaultFSCache.Stat(gitPath)
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
