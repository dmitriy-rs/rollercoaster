package logger_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		err      error
		expected string
	}{
		{
			name:     "error with message and error",
			message:  "test message",
			err:      errors.New("test error"),
			expected: "ERROR  test message test error",
		},
		{
			name:     "error with message only",
			message:  "test message",
			err:      nil,
			expected: "ERROR  test message",
		},
		{
			name:     "error with error only",
			message:  "",
			err:      errors.New("test error"),
			expected: "ERROR  test error",
		},
		{
			name:     "error with empty message and nil error",
			message:  "",
			err:      nil,
			expected: "ERROR  %!s(<nil>)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger.Error(tt.message, tt.err)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			// Remove ANSI color codes for comparison
			cleanOutput := removeANSICodes(output)
			assert.Contains(t, cleanOutput, tt.expected, "Error output should contain expected message")
		})
	}
}

func TestFatal(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "fatal with error",
			err:      errors.New("fatal error"),
			expected: "ERROR  fatal error",
		},
		{
			name:     "fatal with nil error",
			err:      nil,
			expected: "ERROR  An unknown error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv("BE_CRASHER") == "1" {
				logger.Fatal(tt.err)
				return
			}

			// Test Fatal function by running it in a subprocess
			cmd := exec.Command(os.Args[0], "-test.run=TestFatal/"+tt.name)
			cmd.Env = append(os.Environ(), "BE_CRASHER=1")

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Fatal should exit with code 1
			if e, ok := err.(*exec.ExitError); ok && !e.Success() {
				// Check that it exited with status 1
				assert.Equal(t, 1, e.ExitCode(), "Fatal should exit with code 1")

				// Check stderr output
				output := strings.TrimSpace(stderr.String())
				cleanOutput := removeANSICodes(output)
				assert.Contains(t, cleanOutput, tt.expected, "Fatal output should contain expected message")
			} else {
				t.Fatalf("process ran with err %v, want exit status 1", err)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "info message",
			message:  "test info message",
			expected: "INFO  test info message",
		},
		{
			name:     "empty info message",
			message:  "",
			expected: "INFO",
		},
		{
			name:     "info with special characters",
			message:  "test with special chars: !@#$%^&*()",
			expected: "INFO  test with special chars: !@#$%^&*()",
		},
		{
			name:     "info with newlines",
			message:  "test\nwith\nnewlines",
			expected: "INFO  test\nwith\nnewlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger.Info(tt.message)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			// Remove ANSI color codes for comparison
			cleanOutput := removeANSICodes(output)
			assert.Contains(t, cleanOutput, tt.expected, "Info output should contain expected message")
		})
	}
}

func TestWarning(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "warning message",
			message:  "test warning message",
			expected: "WARN  test warning message",
		},
		{
			name:     "empty warning message",
			message:  "",
			expected: "WARN",
		},
		{
			name:     "warning with special characters",
			message:  "warning with special chars: !@#$%^&*()",
			expected: "WARN  warning with special chars: !@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger.Warning(tt.message)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			// Remove ANSI color codes for comparison
			cleanOutput := removeANSICodes(output)
			assert.Contains(t, cleanOutput, tt.expected, "Warning output should contain expected message")
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name     string
		messages []any
		expected string
	}{
		{
			name:     "debug with string",
			messages: []any{"test debug message"},
			expected: "DEBG  [test debug message]",
		},
		{
			name:     "debug with multiple values",
			messages: []any{"debug", 123, true},
			expected: "DEBG  [debug %!s(int=123) %!s(bool=true)]",
		},
		{
			name:     "debug with no messages",
			messages: []any{},
			expected: "DEBG  []",
		},
		{
			name:     "debug with nil value",
			messages: []any{nil},
			expected: "DEBG  [<nil>]",
		},
		{
			name:     "debug with mixed types",
			messages: []any{"string", 42, 3.14, true, nil},
			expected: "DEBG  [string %!s(int=42) %!s(float64=3.14) %!s(bool=true) <nil>]",
		},
		{
			name:     "debug with struct",
			messages: []any{struct{ Name string }{"test"}},
			expected: "DEBG  [{test}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger.Debug(tt.messages...)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			// Remove ANSI color codes for comparison
			cleanOutput := removeANSICodes(output)
			assert.Contains(t, cleanOutput, tt.expected, "Debug output should contain expected message")
		})
	}
}

// removeANSICodes removes ANSI escape sequences from a string for testing
func removeANSICodes(s string) string {
	// Simple regex to remove ANSI escape sequences
	// This is a basic implementation for testing purposes
	result := ""
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // skip the '['
			continue
		}

		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}

		result += string(s[i])
	}

	return result
}
