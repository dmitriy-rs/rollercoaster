package manager_test

import (
	"path/filepath"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
)

// Test structures for JSON and YAML parsing
type PackageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
}

type TaskfileYAML struct {
	Version string                 `yaml:"version"`
	Tasks   map[string]interface{} `yaml:"tasks"`
}

var testdataDir = "testdata/manager-file"

func TestParseFileAsJson(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		expected    PackageJSON
	}{
		{
			name:        "valid package.json",
			filename:    "package.json",
			expectError: false,
			expected: PackageJSON{
				Name:        "example-project",
				Version:     "1.2.3",
				Description: "An example Node.js project for testing",
			},
		},
		{
			name:        "minimal package.json",
			filename:    "minimal-package.json",
			expectError: false,
			expected: PackageJSON{
				Name:    "minimal-project",
				Version: "0.1.0",
			},
		},
		{
			name:        "complex package.json",
			filename:    "complex-package.json",
			expectError: false,
			expected: PackageJSON{
				Name:        "@scope/complex-project",
				Version:     "2.5.1-beta.3",
				Description: "A complex Node.js project with advanced configuration",
			},
		},
		{
			name:        "invalid JSON",
			filename:    "invalid.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := manager.FindInDirectory(&testdataDir, tt.filename)
			if mf == nil {
				t.Fatalf("Failed to find test file: %s", tt.filename)
			}

			result, err := manager.ParseFileAsJson[PackageJSON](mf)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result.Name != tt.expected.Name {
					t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
				}
				if result.Version != tt.expected.Version {
					t.Errorf("Expected version %s, got %s", tt.expected.Version, result.Version)
				}
				if result.Description != tt.expected.Description {
					t.Errorf("Expected description %s, got %s", tt.expected.Description, result.Description)
				}
			}
		})
	}
}

func TestParseFileAsYaml(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		expected    TaskfileYAML
	}{
		{
			name:        "valid Taskfile.yml",
			filename:    "Taskfile.yml",
			expectError: false,
			expected: TaskfileYAML{
				Version: "3",
			},
		},
		{
			name:        "minimal Taskfile.yml",
			filename:    "minimal-Taskfile.yml",
			expectError: false,
			expected: TaskfileYAML{
				Version: "3",
			},
		},
		{
			name:        "invalid YAML",
			filename:    "invalid.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := manager.FindInDirectory(&testdataDir, tt.filename)
			if mf == nil {
				t.Fatalf("Failed to find test file: %s", tt.filename)
			}

			result, err := manager.ParseFileAsYaml[TaskfileYAML](mf)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result.Version != tt.expected.Version {
					t.Errorf("Expected version %s, got %s", tt.expected.Version, result.Version)
				}
				// Verify that tasks were parsed for valid files
				if len(result.Tasks) == 0 {
					t.Error("Expected tasks to be parsed but got empty map")
				}
			}
		})
	}
}

func TestFindInDirectory(t *testing.T) {
	tests := []struct {
		name     string
		dir      *string
		filename string
		expected bool
	}{
		{
			name:     "find existing package.json",
			dir:      &testdataDir,
			filename: "package.json",
			expected: true,
		},
		{
			name:     "find existing Taskfile.yml",
			dir:      &testdataDir,
			filename: "Taskfile.yml",
			expected: true,
		},
		{
			name:     "find minimal-package.json",
			dir:      &testdataDir,
			filename: "minimal-package.json",
			expected: true,
		},
		{
			name:     "file does not exist",
			dir:      &testdataDir,
			filename: "nonexistent.json",
			expected: false,
		},
		{
			name:     "empty filename",
			dir:      &testdataDir,
			filename: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.FindInDirectory(tt.dir, tt.filename)

			if tt.expected && result == nil {
				t.Error("Expected to find file but got nil")
			}
			if !tt.expected && result != nil {
				t.Error("Expected nil but found file")
			}

			if result != nil {
				expectedPath := filepath.Join(*tt.dir, tt.filename)
				if result.Filename != expectedPath {
					t.Errorf("Expected filename %s, got %s", expectedPath, result.Filename)
				}
				if len(result.File) == 0 {
					t.Error("Expected file content but got empty")
				}
			}
		})
	}
}

func TestFindFirstInDirectory(t *testing.T) {
	tests := []struct {
		name      string
		dir       *string
		filenames []string
		expected  string // expected filename that should be found
	}{
		{
			name:      "find first existing file - package.json first",
			dir:       &testdataDir,
			filenames: []string{"package.json", "Taskfile.yml"},
			expected:  "package.json",
		},
		{
			name:      "find first existing file - Taskfile.yml first",
			dir:       &testdataDir,
			filenames: []string{"Taskfile.yml", "package.json"},
			expected:  "Taskfile.yml",
		},
		{
			name:      "find first existing file - skip nonexistent",
			dir:       &testdataDir,
			filenames: []string{"nonexistent.json", "minimal-package.json", "Taskfile.yml"},
			expected:  "minimal-package.json",
		},
		{
			name:      "no files exist",
			dir:       &testdataDir,
			filenames: []string{"nonexistent1.json", "nonexistent2.yml"},
			expected:  "",
		},
		{
			name:      "empty filenames list",
			dir:       &testdataDir,
			filenames: []string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.FindFirstInDirectory(tt.dir, tt.filenames)

			if tt.expected == "" && result != nil {
				t.Error("Expected nil but got result")
			}
			if tt.expected != "" && result == nil {
				t.Error("Expected result but got nil")
			}

			if result != nil {
				expectedPath := filepath.Join(*tt.dir, tt.expected)
				if result.Filename != expectedPath {
					t.Errorf("Expected filename %s, got %s", expectedPath, result.Filename)
				}
			}
		})
	}
}

func TestFindInDirectoryWithNilDir(t *testing.T) {
	// This test should trigger os.Exit(1) due to nil directory
	// We can't easily test os.Exit in unit tests, so we'll just verify the function handles nil gracefully
	// In a real scenario, you might want to refactor the code to return errors instead of calling os.Exit

	// For now, we'll skip this test as it would cause the test suite to exit
	t.Skip("Skipping test that would cause os.Exit(1)")
}
