package mocks

import (
	"os/exec"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
)

// MockJsWorkspace implements the JsWorkspace interface for testing
type MockJsWorkspace struct {
	name         string
	execName     string
	cmdArgs      []string
	runCmdArgs   []string
	installArgs  []string
	addArgs      []string
	removeArgs   []string
	executedCmds [][]string
}

// NewMockJsWorkspace creates a new mock workspace with default command configurations
func NewMockJsWorkspace(name string) *MockJsWorkspace {
	return &MockJsWorkspace{
		name:         name,
		execName:     name + "x",
		cmdArgs:      []string{name},
		runCmdArgs:   []string{name, "run"},
		installArgs:  []string{name, "install"},
		addArgs:      []string{name, "add"},
		removeArgs:   []string{name, "remove"},
		executedCmds: make([][]string, 0),
	}
}

// NewMockJsWorkspaceWithExecName creates a new mock workspace with custom exec name
func NewMockJsWorkspaceWithExecName(name, execName string) *MockJsWorkspace {
	mock := NewMockJsWorkspace(name)
	mock.execName = execName
	return mock
}

func (m *MockJsWorkspace) Name() string {
	return m.name
}

func (m *MockJsWorkspace) ExecName() string {
	return m.execName
}

func (m *MockJsWorkspace) Cmd() *exec.Cmd {
	cmd := exec.Command("echo", m.cmdArgs...)
	m.executedCmds = append(m.executedCmds, cmd.Args)
	return cmd
}

func (m *MockJsWorkspace) ExecuteCmd() *exec.Cmd {
	cmd := exec.Command("echo", m.runCmdArgs...)
	m.executedCmds = append(m.executedCmds, cmd.Args)
	return cmd
}

func (m *MockJsWorkspace) InstallCmd() *exec.Cmd {
	cmd := exec.Command("echo", m.installArgs...)
	m.executedCmds = append(m.executedCmds, cmd.Args)
	return cmd
}

func (m *MockJsWorkspace) AddCmd() *exec.Cmd {
	cmd := exec.Command("echo", m.addArgs...)
	m.executedCmds = append(m.executedCmds, cmd.Args)
	return cmd
}

func (m *MockJsWorkspace) RemoveCmd() *exec.Cmd {
	cmd := exec.Command("echo", m.removeArgs...)
	m.executedCmds = append(m.executedCmds, cmd.Args)
	return cmd
}

// GetExecutedCommands returns all commands that have been executed
func (m *MockJsWorkspace) GetExecutedCommands() [][]string {
	return m.executedCmds
}

// ClearExecutedCommands clears the list of executed commands
func (m *MockJsWorkspace) ClearExecutedCommands() {
	m.executedCmds = make([][]string, 0)
}

// ToJsWorkspace converts the mock to a JsWorkspace interface
func (m *MockJsWorkspace) ToJsWorkspace() jsmanager.JsWorkspace {
	return jsmanager.JsWorkspace(m)
}

// ToJsWorkspacePtr converts the mock to a JsWorkspace interface pointer
func (m *MockJsWorkspace) ToJsWorkspacePtr() *jsmanager.JsWorkspace {
	workspace := jsmanager.JsWorkspace(m)
	return &workspace
}
