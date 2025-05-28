package manager_test

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
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
	manager.TaskExecute(cmd)

	// The test passes if the function returns without panicking
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
	manager.TaskExecute(cmd)

	// The function should return gracefully even on error
	// This test passes if no panic occurs
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
	manager.TaskExecute(cmd, "test", "argument")

	// Verify the arguments were added (they should be added before execution)
	if len(cmd.Args) <= originalArgsLen {
		t.Errorf("Expected more than %d arguments after adding args, got %d", originalArgsLen, len(cmd.Args))
	}

	// Check that our additional args were added
	argsFound := false
	for i, arg := range cmd.Args {
		if arg == "test" && i+1 < len(cmd.Args) && cmd.Args[i+1] == "argument" {
			argsFound = true
			break
		}
	}

	if !argsFound {
		t.Error("Additional arguments 'test' and 'argument' were not found in command args")
	}
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
	manager.TaskExecute(cmd)

	// Verify no additional args were added
	if len(cmd.Args) != originalArgsLen {
		t.Errorf("Expected %d arguments, got %d", originalArgsLen, len(cmd.Args))
	}
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
	if cmd.Stdout != nil {
		t.Error("Expected cmd.Stdout to be nil initially")
	}
	if cmd.Stderr != nil {
		t.Error("Expected cmd.Stderr to be nil initially")
	}

	// Execute the task
	manager.TaskExecute(cmd)

	// Verify that stdout and stderr are set to os.Stdout and os.Stderr
	if cmd.Stdout != os.Stdout {
		t.Error("Expected cmd.Stdout to be set to os.Stdout")
	}
	if cmd.Stderr != os.Stderr {
		t.Error("Expected cmd.Stderr to be set to os.Stderr")
	}
}

func TestTaskExecute_CommandNotFound(t *testing.T) {
	// Create a command that doesn't exist
	cmd := exec.Command("nonexistentcommand12345")

	// Execute the task - should handle the error gracefully
	manager.TaskExecute(cmd)

	// The function should return without panicking
	// This test passes if no panic occurs
}

func TestTaskExecute_EmptyCommand(t *testing.T) {
	// Test with an empty command (this should fail gracefully)
	cmd := exec.Command("")

	// Execute the task - should handle the error gracefully
	manager.TaskExecute(cmd)

	// The function should return without panicking
	// This test passes if no panic occurs
}
