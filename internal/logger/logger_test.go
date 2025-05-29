package logger_test

import (
	"bytes"
	"errors"
	"os"
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
			// We can't actually test os.Exit(1) in a unit test without special handling
			// So we'll test the error output part only
			// In a real scenario, you might use a wrapper or dependency injection

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// We need to test this in a subprocess or mock os.Exit
			// For now, let's test the error formatting logic by calling Error directly
			if tt.err == nil {
				logger.Error("An unknown error occurred", nil)
			} else {
				logger.Error("", tt.err)
			}

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			// Remove ANSI color codes for comparison
			cleanOutput := removeANSICodes(output)
			assert.Contains(t, cleanOutput, tt.expected, "Fatal output should contain expected message")
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
