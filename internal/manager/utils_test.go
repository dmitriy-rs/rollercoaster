package manager_test

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/stretchr/testify/assert"
)

func TestTaskExecute_SuccessfulCommand(t *testing.T) {
	// Create a command that will succeed
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo", "Hello World")
	} else {
		cmd = exec.Command("echo", "Hello World")
	}

	// Execute the task - should complete without error
	assert.NotPanics(t, func() {
		manager.CommandExecute(cmd)
	}, "CommandExecute should not panic on successful command")

	// The output will go to os.Stdout as intended by the function
}

func TestTaskExecute_CommandWithError(t *testing.T) {
	// Create a command that will fail
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "exit", "1")
	} else {
		cmd = exec.Command("sh", "-c", "exit 1")
	}

	// Execute the task - should not panic or exit
	assert.NotPanics(t, func() {
		manager.CommandExecute(cmd)
	}, "CommandExecute should handle errors gracefully without panicking")
}

func TestTaskExecute_WithAdditionalArgs(t *testing.T) {
	// Create a base command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo")
	} else {
		cmd = exec.Command("echo")
	}

	// Store original args to verify they were modified
	originalArgsLen := len(cmd.Args)

	// Execute with additional arguments
	manager.CommandExecute(cmd, "test", "argument")

	// Verify the arguments were added (they should be added before execution)
	assert.Greater(t, len(cmd.Args), originalArgsLen, "Additional arguments should be added to command")

	// Check that our additional args were added
	argsFound := false
	for i, arg := range cmd.Args {
		if arg == "test" && i+1 < len(cmd.Args) && cmd.Args[i+1] == "argument" {
			argsFound = true
			break
		}
	}

	assert.True(t, argsFound, "Additional arguments 'test' and 'argument' should be found in command args")
}

func TestTaskExecute_NoAdditionalArgs(t *testing.T) {
	// Create a command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo", "test")
	} else {
		cmd = exec.Command("echo", "test")
	}

	// Store original args length
	originalArgsLen := len(cmd.Args)

	// Execute without additional arguments
	manager.CommandExecute(cmd)

	// Verify no additional args were added
	assert.Equal(t, originalArgsLen, len(cmd.Args), "No additional arguments should be added when none provided")
}

func TestTaskExecute_StdoutStderrAssignment(t *testing.T) {
	// Create a command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo", "test")
	} else {
		cmd = exec.Command("echo", "test")
	}

	// Initially stdout and stderr should be nil
	assert.Nil(t, cmd.Stdout, "cmd.Stdout should be nil initially")
	assert.Nil(t, cmd.Stderr, "cmd.Stderr should be nil initially")

	// Execute the task
	manager.CommandExecute(cmd)

	// Verify that stdout and stderr are set to os.Stdout and os.Stderr
	assert.Equal(t, os.Stdout, cmd.Stdout, "cmd.Stdout should be set to os.Stdout")
	assert.Equal(t, os.Stderr, cmd.Stderr, "cmd.Stderr should be set to os.Stderr")
}

func TestTaskExecute_CommandNotFound(t *testing.T) {
	// Create a command that doesn't exist
	cmd := exec.Command("nonexistentcommand12345")

	// Execute the task - should handle the error gracefully
	assert.NotPanics(t, func() {
		manager.CommandExecute(cmd)
	}, "CommandExecute should handle non-existent commands gracefully without panicking")
}

func TestTaskExecute_EmptyCommand(t *testing.T) {
	// Test with an empty command (this should fail gracefully)
	cmd := exec.Command("")

	// Execute the task - should handle the error gracefully
	assert.NotPanics(t, func() {
		manager.CommandExecute(cmd)
	}, "CommandExecute should handle empty commands gracefully without panicking")
}
