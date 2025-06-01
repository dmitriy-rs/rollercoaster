package jsmanager

import (
	"os/exec"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

type JsWorkspaceManager struct {
	Workspace *JsWorkspace
}

var (
	WorkspaceInstallTask      = "install"
	WorkspaceAddTask          = "add"
	WorkspaceRemoveTask       = "remove"
	WorkspaceExecuteTask      = "x"
)

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
