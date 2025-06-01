package jsmanager

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

type JsWorkspace interface {
	Name() string
	ExecName() string
	Cmd() *exec.Cmd

	ExecuteCmd() *exec.Cmd
	InstallCmd() *exec.Cmd
	AddCmd() *exec.Cmd
	RemoveCmd() *exec.Cmd
}

func ParseJsWorkspace(dir *string, defaultJSManager string) (*JsWorkspace, error) {
	// Check if package.json exists first
	if !cache.DefaultFSCache.FileExists(filepath.Join(*dir, packageJsonFilename)) {
		return nil, nil
	}

	// Batch check for all lock files in one operation
	lockFiles := []string{pnpmLockFilename, "yarn.lock", "package-lock.json"}
	lockFileExists := cache.DefaultFSCache.FindFilesInDirectory(*dir, lockFiles)

	// Count existing lock files for early validation
	lockFileCount := 0
	pnpmLockExists := lockFileExists[pnpmLockFilename]
	yarnLockExists := lockFileExists["yarn.lock"]
	npmLockExists := lockFileExists["package-lock.json"]

	if pnpmLockExists {
		lockFileCount++
	}
	if yarnLockExists {
		lockFileCount++
	}
	if npmLockExists {
		lockFileCount++
	}

	// Early exit if multiple lock files detected
	if lockFileCount > 1 {
		return nil, fmt.Errorf("multiple package manager lock files detected")
	}

	// Pre-allocate slice with capacity 1 (typically only one workspace type)
	workspaces := make([]JsWorkspace, 0, 1)

	if pnpmLockExists {
		pnpmWorkspace, err := ParsePnpmWorkspace(dir)
		if err != nil {
			logger.Warning(err.Error())
		}
		if pnpmWorkspace != nil {
			workspaces = append(workspaces, pnpmWorkspace)
		}
	}

	if yarnLockExists {
		yarnWorkspace, err := ParseYarnWorkspace(dir)
		if err != nil {
			logger.Warning(err.Error())
		}
		if yarnWorkspace != nil {
			workspaces = append(workspaces, yarnWorkspace)
		}
	}

	if npmLockExists {
		npmWorkspace, err := ParseNpmWorkspace(dir)
		if err != nil {
			logger.Warning(err.Error())
		}
		if npmWorkspace != nil {
			workspaces = append(workspaces, npmWorkspace)
		}
	}

	// Only use default if package.json exists but no workspace was detected
	if len(workspaces) == 0 {
		defaultJSManager := parseDefaultJSManager(defaultJSManager)
		if defaultJSManager == nil {
			return nil, nil
		}
		return &defaultJSManager, nil
	}

	if len(workspaces) > 1 {
		formattedWorkspaces := strings.Builder{}
		for _, workspace := range workspaces {
			formattedWorkspaces.WriteString(workspace.Name())
			formattedWorkspaces.WriteString(", ")
		}
		return nil, fmt.Errorf("multiple workspace package managers found: %s", formattedWorkspaces.String())
	}

	return &workspaces[0], nil
}

func parseDefaultJSManager(defaultJSManager string) JsWorkspace {
	switch defaultJSManager {
	case "npm":
		workspace := GetDefaultNpmWorkspace()
		return &workspace
	case "yarn":
		workspace := GetDefaultYarnWorkspace()
		return &workspace
	case "pnpm":
		workspace := GetDefaultPnpmWorkspace()
		return &workspace
	default:
		return nil
	}
}

