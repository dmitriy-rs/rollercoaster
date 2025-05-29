package mocks_test

import (
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/js/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMockJsWorkspace_BasicFunctionality(t *testing.T) {
	mock := mocks.NewMockJsWorkspace("test-workspace")

	// Test basic properties
	assert.Equal(t, "test-workspace", mock.Name(), "Name() should return correct name")
	assert.Equal(t, "test-workspacex", mock.ExecName(), "ExecName() should return correct exec name")

	// Test command tracking
	mock.ClearExecutedCommands()
	assert.Empty(t, mock.GetExecutedCommands(), "Should start with no executed commands")

	// Test command execution tracking
	cmd := mock.Cmd()
	assert.Equal(t, []string{"echo", "test-workspace"}, cmd.Args, "Cmd() should return correct args")

	executedCmds := mock.GetExecutedCommands()
	assert.Len(t, executedCmds, 1, "Should track executed command")
	assert.Equal(t, []string{"echo", "test-workspace"}, executedCmds[0], "Should track correct command")
}

func TestMockJsWorkspace_WithCustomExecName(t *testing.T) {
	mock := mocks.NewMockJsWorkspaceWithExecName("custom-workspace", "custom-exec")

	assert.Equal(t, "custom-workspace", mock.Name(), "Name() should return correct name")
	assert.Equal(t, "custom-exec", mock.ExecName(), "ExecName() should return custom exec name")
}

func TestMockJsWorkspace_AllCommands(t *testing.T) {
	mock := mocks.NewMockJsWorkspace("test")
	mock.ClearExecutedCommands()

	// Test all command methods
	commands := []struct {
		name     string
		cmd      func() []string
		expected []string
	}{
		{"Cmd", func() []string { return mock.Cmd().Args }, []string{"echo", "test"}},
		{"ExecuteCmd", func() []string { return mock.ExecuteCmd().Args }, []string{"echo", "test", "run"}},
		{"InstallCmd", func() []string { return mock.InstallCmd().Args }, []string{"echo", "test", "install"}},
		{"AddCmd", func() []string { return mock.AddCmd().Args }, []string{"echo", "test", "add"}},
		{"RemoveCmd", func() []string { return mock.RemoveCmd().Args }, []string{"echo", "test", "remove"}},
	}

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			result := cmd.cmd()
			assert.Equal(t, cmd.expected, result, "%s should return correct args", cmd.name)
		})
	}

	// Verify all commands were tracked
	executedCmds := mock.GetExecutedCommands()
	assert.Len(t, executedCmds, len(commands), "Should track all executed commands")
}

func TestMockJsWorkspace_InterfaceConversion(t *testing.T) {
	mock := mocks.NewMockJsWorkspace("test")

	// Test ToJsWorkspace conversion
	workspace := mock.ToJsWorkspace()
	assert.Equal(t, "test", workspace.Name(), "Converted workspace should work correctly")

	// Test ToJsWorkspacePtr conversion
	workspacePtr := mock.ToJsWorkspacePtr()
	assert.NotNil(t, workspacePtr, "Should return non-nil pointer")
	assert.Equal(t, "test", (*workspacePtr).Name(), "Converted workspace pointer should work correctly")
}
