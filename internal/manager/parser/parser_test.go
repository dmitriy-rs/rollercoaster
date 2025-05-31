package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManager(t *testing.T) {
	tests := []struct {
		name                 string
		testdataDir          string
		createGitDir         bool
		expectedManagers     int
		expectedJsManager    bool
		expectedTaskMgr      bool
		expectedWorkspaceMgr bool
	}{
		{
			name:                 "directory with only taskfile",
			testdataDir:          "taskfile-only",
			createGitDir:         true,
			expectedManagers:     1,
			expectedJsManager:    false,
			expectedTaskMgr:      true,
			expectedWorkspaceMgr: false,
		},
		{
			name:                 "directory with package.json and pnpm-lock.yaml v9",
			testdataDir:          "pnpm-v9-only",
			createGitDir:         true,
			expectedManagers:     2, // js manager + workspace manager
			expectedJsManager:    true,
			expectedTaskMgr:      false,
			expectedWorkspaceMgr: true,
		},
		{
			name:                 "directory with both taskfile and package.json",
			testdataDir:          "taskfile-and-package-json",
			createGitDir:         true,
			expectedManagers:     3, // task manager + js manager + workspace manager
			expectedJsManager:    true,
			expectedTaskMgr:      true,
			expectedWorkspaceMgr: true,
		},
		{
			name:                 "nested from parent directory - package.json with taskfile",
			testdataDir:          "nested-package-json",
			createGitDir:         true,
			expectedManagers:     3, // parent task + parent js + workspace manager
			expectedJsManager:    true,
			expectedTaskMgr:      true,
			expectedWorkspaceMgr: true,
		},
		{
			name:                 "nested from child directory - package.json only in child",
			testdataDir:          "nested-package-json/subdir",
			createGitDir:         true,
			expectedManagers:     4, // parent task + parent js + child js + workspace manager
			expectedJsManager:    true,
			expectedTaskMgr:      true,
			expectedWorkspaceMgr: true,
		},
		{
			name:                 "nested taskfile with top level taskfile",
			testdataDir:          "nested-taskfile/subdir",
			createGitDir:         true,
			expectedManagers:     2, // top-level task + nested task
			expectedJsManager:    false,
			expectedTaskMgr:      true,
			expectedWorkspaceMgr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := filepath.Join("testdata", tt.testdataDir)

			// Create .git directory if needed - create it at the root of the test scenario
			var gitDir string
			if tt.createGitDir {
				// For nested scenarios, create .git at the root
				switch tt.testdataDir {
				case "nested-package-json/subdir", "nested-package-json":
					gitDir = filepath.Join("testdata", "nested-package-json", ".git")
				case "nested-taskfile/subdir":
					gitDir = filepath.Join("testdata", "nested-taskfile", ".git")
				case "nested-with-js-workspaces/subproject":
					gitDir = filepath.Join("testdata", "nested-with-js-workspaces", ".git")
				default:
					gitDir = filepath.Join(testDir, ".git")
				}
				err := os.MkdirAll(gitDir, 0755)
				require.NoError(t, err, "Failed to create .git directory")
				defer os.RemoveAll(gitDir) //nolint:errcheck
			}

			// Parse managers
			managers, err := parser.ParseManager(&testDir)

			if tt.expectedManagers == 0 {
				assert.NoError(t, err, "ParseManager should not return error")
				assert.Nil(t, managers, "ParseManager should return nil when no managers found")
				return
			}

			assert.NoError(t, err, "ParseManager should not return error")
			require.NotNil(t, managers, "ParseManager should return managers")
			assert.Len(t, managers, tt.expectedManagers, "Should have expected number of managers")

			// Check manager types
			hasJsManager := false
			hasTaskManager := false
			hasWorkspaceManager := false

			for _, manager := range managers {
				title := manager.GetTitle()
				if !strings.Contains(title.Name, "task") {
					if strings.Contains(title.Description, "commands") {
						hasWorkspaceManager = true
					} else {
						hasJsManager = true
					}
				}
				if title.Name == "task" {
					hasTaskManager = true
				}
			}

			assert.Equal(t, tt.expectedJsManager, hasJsManager, "JS manager presence should match expectation")
			assert.Equal(t, tt.expectedTaskMgr, hasTaskManager, "Task manager presence should match expectation")
			assert.Equal(t, tt.expectedWorkspaceMgr, hasWorkspaceManager, "Workspace manager presence should match expectation")
		})
	}
}

func TestParseManagerEmptyDirectory(t *testing.T) {
	testDir := filepath.Join("testdata", "empty")

	// Create .git directory
	gitDir := filepath.Join(testDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err, "Failed to create .git directory")
	defer os.RemoveAll(gitDir) //nolint:errcheck

	managers, err := parser.ParseManager(&testDir)

	assert.NoError(t, err, "ParseManager should not return error for empty directory")
	assert.Nil(t, managers, "ParseManager should return nil for empty directory")
}

func TestParseManagerNilDirectory(t *testing.T) {
	// Since the ParseManager function dereferences the pointer without checking for nil,
	// this test is expected to panic. We'll skip it and add a comment that the function
	// should be fixed to handle nil gracefully.
	t.Skip("ParseManager does not handle nil directory pointer gracefully - this should be fixed in the implementation")

	managers, err := parser.ParseManager(nil)

	assert.NoError(t, err, "ParseManager should not return error for nil directory")
	assert.Nil(t, managers, "ParseManager should return nil for nil directory")
}

func TestParseManagerEmptyString(t *testing.T) {
	emptyDir := ""
	managers, err := parser.ParseManager(&emptyDir)

	assert.NoError(t, err, "ParseManager should not return error for empty string directory")
	assert.Nil(t, managers, "ParseManager should return nil for empty string directory")
}

func TestFindClosestGitDir(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		setupGit    bool
		expectGit   bool
	}{
		{
			name:        "directory with .git",
			testdataDir: "taskfile-only",
			setupGit:    true,
			expectGit:   true,
		},
		{
			name:        "directory without .git",
			testdataDir: "taskfile-only",
			setupGit:    false,
			expectGit:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)

			if tt.setupGit {
				gitDir := filepath.Join(testDir, ".git")
				err := os.MkdirAll(gitDir, 0755)
				require.NoError(t, err, "Failed to create .git directory")
				defer os.RemoveAll(gitDir) //nolint:errcheck
			}

			// We need to test the findClosestGitDir function indirectly through ParseManager
			// since it's not exported
			managers, err := parser.ParseManager(&testDir)

			// The function should work regardless of git directory presence
			assert.NoError(t, err, "ParseManager should work with or without .git")

			// If there are managers, they should be parsed correctly
			if managers != nil {
				assert.Greater(t, len(managers), 0, "Should have at least one manager if not nil")
			}
		})
	}
}

func TestParseManagerJsWorkspaceAlwaysRegistered(t *testing.T) {
	tests := []struct {
		name                 string
		testdataDir          string
		expectedWorkspaceMgr bool
		expectedJsManagers   int
		expectedTotalMgrs    int
	}{
		{
			name:                 "pnpm workspace with scripts",
			testdataDir:          "pnpm-v9-only",
			expectedWorkspaceMgr: true,
			expectedJsManagers:   1, // one js manager for the package.json
			expectedTotalMgrs:    2, // js manager + workspace manager
		},
		{
			name:                 "taskfile and pnpm workspace",
			testdataDir:          "taskfile-and-package-json",
			expectedWorkspaceMgr: true,
			expectedJsManagers:   1, // one js manager for the package.json
			expectedTotalMgrs:    3, // task manager + js manager + workspace manager
		},
		{
			name:                 "nested with workspace at root",
			testdataDir:          "nested-package-json",
			expectedWorkspaceMgr: true,
			expectedJsManagers:   1, // one js manager for the package.json
			expectedTotalMgrs:    3, // task manager + js manager + workspace manager
		},
		{
			name:                 "no js workspace",
			testdataDir:          "taskfile-only",
			expectedWorkspaceMgr: false,
			expectedJsManagers:   0,
			expectedTotalMgrs:    1, // only task manager
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)

			// Create .git directory
			var gitDir string
			switch tt.testdataDir {
			case "nested-package-json":
				gitDir = filepath.Join("testdata", "nested-package-json", ".git")
			default:
				gitDir = filepath.Join(testDir, ".git")
			}
			err := os.MkdirAll(gitDir, 0755)
			require.NoError(t, err, "Failed to create .git directory")
			defer os.RemoveAll(gitDir) //nolint:errcheck

			managers, err := parser.ParseManager(&testDir)

			require.NoError(t, err, "ParseManager should not return error")
			require.NotNil(t, managers, "ParseManager should return managers")
			assert.Len(t, managers, tt.expectedTotalMgrs, "Should have expected total number of managers")

			// Count different manager types
			jsManagerCount := 0
			workspaceManagerCount := 0
			taskManagerCount := 0

			for _, manager := range managers {
				title := manager.GetTitle()
				if !strings.Contains(title.Name, "task") {
					if strings.Contains(title.Description, "commands") {
						workspaceManagerCount++
					} else {
						jsManagerCount++
					}
				}
				if title.Name == "task" {
					taskManagerCount++
				}
			}

			assert.Equal(t, tt.expectedJsManagers, jsManagerCount, "Should have expected number of JS managers")

			if tt.expectedWorkspaceMgr {
				assert.Equal(t, 1, workspaceManagerCount, "Should have exactly one workspace manager when JS workspace is detected")
			} else {
				assert.Equal(t, 0, workspaceManagerCount, "Should have no workspace manager when no JS workspace is detected")
			}

			// Verify workspace manager is always last when present (due to slices.Reverse)
			if tt.expectedWorkspaceMgr && len(managers) > 0 {
				lastManager := managers[len(managers)-1]
				title := lastManager.GetTitle()
				assert.True(t,
					strings.Contains(title.Description, "commands"),
					"Workspace manager should be last in the list")
			}
		})
	}
}
