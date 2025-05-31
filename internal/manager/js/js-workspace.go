package jsmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
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

func ParseJsWorkspace(dir *string) (*JsWorkspace, error) {
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

	if len(workspaces) == 0 {
		return nil, nil
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

var (
	WorkspaceInstallTask      = "install"
	WorkspaceInstallAliasTask = "i"
	WorkspaceAddTask          = "add"
	WorkspaceRemoveTask       = "remove"
	WorkspaceExecuteTask      = "x"
)

type JsWorkspaceManager struct {
	Workspace *JsWorkspace
}

func (m *JsWorkspaceManager) ListTasks() ([]task.Task, error) {
	tasks := []task.Task{
		{
			Name:        WorkspaceAddTask,
			Description: "Install a dependency",
		},
		{
			Name:        WorkspaceRemoveTask,
			Description: "Remove a dependency",
		},
		{
			Name:        WorkspaceInstallTask,
			Description: "Install dependencies",
		},
		{
			Name:        WorkspaceExecuteTask,
			Description: "Execute a command",
			Aliases:     []string{(*m.Workspace).ExecName()},
		},
	}
	return tasks, nil
}

func (m *JsWorkspaceManager) ExecuteTask(task *task.Task, args ...string) {
	var cmd *exec.Cmd

	switch task.Name {
	case WorkspaceInstallTask:
		cmd = (*m.Workspace).InstallCmd()
	case WorkspaceAddTask:
		cmd = (*m.Workspace).AddCmd()
	case WorkspaceRemoveTask:
		cmd = (*m.Workspace).RemoveCmd()
	case WorkspaceExecuteTask:
		cmd = (*m.Workspace).ExecuteCmd()
	default:
		cmd = (*m.Workspace).Cmd()
	}

	manager.CommandExecute(cmd, append([]string{task.Name}, args...)...)
}

func (m *JsWorkspaceManager) GetTitle() manager.Title {
	return manager.Title{
		Name:        (*m.Workspace).Name(),
		Description: "package commands",
	}
}
