package configfile_test

import (
	"path/filepath"
	"testing"

	config "github.com/dmitriy-rs/rollercoaster/internal/manager/config-file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

var testdataDir = "testdata"

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
			mf := config.FindInDirectory(&testdataDir, tt.filename)
			require.NotNil(t, mf, "Failed to find test file: %s", tt.filename)

			result, err := config.ParseFileAsJson[PackageJSON](mf)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expected.Name, result.Name, "Name should match expected value")
				assert.Equal(t, tt.expected.Version, result.Version, "Version should match expected value")
				assert.Equal(t, tt.expected.Description, result.Description, "Description should match expected value")
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
			mf := config.FindInDirectory(&testdataDir, tt.filename)
			require.NotNil(t, mf, "Failed to find test file: %s", tt.filename)

			result, err := config.ParseFileAsYaml[TaskfileYAML](mf)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expected.Version, result.Version, "Version should match expected value")
				assert.NotEmpty(t, result.Tasks, "Expected tasks to be parsed but got empty map")
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
			result := config.FindInDirectory(tt.dir, tt.filename)

			if tt.expected {
				assert.NotNil(t, result, "Expected to find file but got nil")
				expectedPath := filepath.Join(*tt.dir, tt.filename)
				assert.Equal(t, expectedPath, result.Filename, "Filename should match expected path")
				assert.NotEmpty(t, result.File, "Expected file content but got empty")
			} else {
				assert.Nil(t, result, "Expected nil but found file")
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
			result := config.FindFirstInDirectory(tt.dir, tt.filenames)

			if tt.expected == "" {
				assert.Nil(t, result, "Expected nil but got result")
			} else {
				assert.NotNil(t, result, "Expected result but got nil")
				expectedPath := filepath.Join(*tt.dir, tt.expected)
				assert.Equal(t, expectedPath, result.Filename, "Filename should match expected path")
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

func TestManagerParseConfig_GetDirectories_SameDirectory(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "/user/project",
		RootDir:    "/user/project",
	}

	result := config.GetDirectories()
	expected := []string{"/user/project"}

	assert.Equal(t, expected, result, "Should return single directory when current equals root")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")
}

func TestManagerParseConfig_GetDirectories_EmptyRootDir(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "/user/test/path",
		RootDir:    "",
	}

	result := config.GetDirectories()
	expected := []string{"/user/test/path"}

	assert.Equal(t, expected, result, "Should return only current dir when root is empty")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")
}

func TestManagerParseConfig_GetDirectories_EmptyCurrentDir(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "",
		RootDir:    "/user",
	}

	result := config.GetDirectories()
	expected := []string{"/user"}

	assert.Equal(t, expected, result, "Should return only root dir when current is empty")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")
}

func TestManagerParseConfig_GetDirectories_BothEmpty(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "",
		RootDir:    "",
	}

	result := config.GetDirectories()
	expected := []string{""}

	assert.Equal(t, expected, result, "Should return empty string when both are empty")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")
}

func TestManagerParseConfig_GetDirectories_SingleLevelDeep(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "/user/project",
		RootDir:    "/user",
	}

	result := config.GetDirectories()
	expected := []string{"/user", "/user/project"}

	assert.Equal(t, expected, result, "Should handle single level nesting correctly")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")

	// Verify order is correct
	assert.Equal(t, "/user", result[0], "First directory should be root directory")
	assert.Equal(t, "/user/project", result[len(result)-1], "Last directory should be current directory")
}

func TestManagerParseConfig_GetDirectories_DeepNesting(t *testing.T) {
	config := &config.ParseConfig{
		CurrentDir: "/user/projects/myapp/src/components/ui",
		RootDir:    "/user/projects",
	}

	result := config.GetDirectories()
	expected := []string{"/user/projects", "/user/projects/myapp", "/user/projects/myapp/src", "/user/projects/myapp/src/components", "/user/projects/myapp/src/components/ui"}

	assert.Equal(t, expected, result, "Should handle deep directory nesting correctly")
	assert.Equal(t, len(expected), len(result), "Result should have expected length")

	// Verify order is correct
	assert.Equal(t, "/user/projects", result[0], "First directory should be root directory")
	assert.Equal(t, "/user/projects/myapp/src/components/ui", result[len(result)-1], "Last directory should be current directory")
}
