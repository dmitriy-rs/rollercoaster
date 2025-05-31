package jsmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
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
	_, err := os.Stat(filepath.Join(*dir, packageJsonFilename))
	if err != nil {
		return nil, nil
	}

	workspaces := []JsWorkspace{}

	pnpmWorkspace, err := ParsePnpmWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if pnpmWorkspace != nil {
		workspaces = append(workspaces, pnpmWorkspace)
	}

	yarnWorkspace, err := ParseYarnWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if yarnWorkspace != nil {
		workspaces = append(workspaces, yarnWorkspace)
	}

	npmWorkspace, err := ParseNpmWorkspace(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	if npmWorkspace != nil {
		workspaces = append(workspaces, npmWorkspace)
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

var (
	WorkspaceInstallTask      = "install"
	WorkspaceInstallAliasTask = "i"
	WorkspaceAddTask          = "add"
	WorkspaceRemoveTask       = "remove"
	WorkspaceExecuteTask      = "x"
)
