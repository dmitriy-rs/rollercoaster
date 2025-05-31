package taskmanager_test

import (
	"os"
	"path/filepath"
	"testing"

	manager "github.com/dmitriy-rs/rollercoaster/internal/manager/task-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskManagerTestDir(t *testing.T, dirName string) string {
	testDir := filepath.Join(os.TempDir(), "task-manager-test", dirName)
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err, "Failed to create test directory")
	return testDir
}

func cleanupTaskManagerTestDir(testDir string) {
	_ = os.RemoveAll(filepath.Dir(testDir))
}

func createTaskFile(t *testing.T, dir, filename, content string) {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create task file %s", filename)
}

const validTaskfileContent = `version: '3'

tasks:
  build:
    desc: "Build the application"
    cmds:
      - echo "Building the application"

  test:
    desc: "Run tests"
    cmds:
      - echo "Running tests"

  lint:
    desc: "Run linters"
    cmds:
      - echo "Running linters"
`

const validDistTaskfileContent = `version: '3'

tasks:
  install:
    desc: "Install dependencies"
    cmds:
      - echo "Installing dependencies"

  clean:
    desc: "Clean build artifacts"
    cmds:
      - echo "Cleaning build artifacts"
`

const invalidTaskfileContent = `version: '2'

tasks:
  build:
    desc: "Build the application"
    cmds:
      - echo "Building the application"
`

const noVersionTaskfileContent = `tasks:
  build:
    desc: "Build the application"
    cmds:
      - echo "Building the application"
`

func TestParseTaskManager_LocalTaskfiles(t *testing.T) {
	localFilenames := []string{
		"Taskfile.yml",
		"taskfile.yml",
		"Taskfile.yaml",
		"taskfile.yaml",
	}

	for _, filename := range localFilenames {
		t.Run("local_"+filename, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "local_"+filename)
			defer cleanupTaskManagerTestDir(testDir)

			createTaskFile(t, testDir, filename, validTaskfileContent)

			tm, err := manager.ParseTaskManager(&testDir)
			require.NoError(t, err, "Should parse task manager without error")
			require.NotNil(t, tm, "Should return a TaskManager instance")

			tasks, err := tm.ListTasks()
			require.NoError(t, err, "Should list tasks without error")

			expectedTasks := []string{"build", "lint", "test"}
			assert.Len(t, tasks, len(expectedTasks), "Should return correct number of tasks")

			for i, expectedName := range expectedTasks {
				assert.Equal(t, expectedName, tasks[i].Name, "Task name should match at index %d", i)
			}

			title := tm.GetTitle()
			assert.NotEmpty(t, title.Name, "Title name should not be empty")
		})
	}
}

func TestParseTaskManager_DistTaskfiles(t *testing.T) {
	distFilenames := []string{
		"Taskfile.dist.yml",
		"taskfile.dist.yml",
		"Taskfile.dist.yaml",
		"taskfile.dist.yaml",
	}

	for _, filename := range distFilenames {
		t.Run("dist_"+filename, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "dist_"+filename)
			defer cleanupTaskManagerTestDir(testDir)

			createTaskFile(t, testDir, filename, validDistTaskfileContent)

			tm, err := manager.ParseTaskManager(&testDir)
			require.NoError(t, err, "Should parse task manager without error")
			require.NotNil(t, tm, "Should return a TaskManager instance")

			tasks, err := tm.ListTasks()
			require.NoError(t, err, "Should list tasks without error")

			expectedTasks := []string{"clean", "install"}
			assert.Len(t, tasks, len(expectedTasks), "Should return correct number of tasks")

			for i, expectedName := range expectedTasks {
				assert.Equal(t, expectedName, tasks[i].Name, "Task name should match at index %d", i)
			}
		})
	}
}

func TestParseTaskManager_BothDistAndLocal(t *testing.T) {
	testCases := []struct {
		name      string
		localFile string
		distFile  string
	}{
		{"yml_files", "Taskfile.yml", "Taskfile.dist.yml"},
		{"yaml_files", "Taskfile.yaml", "Taskfile.dist.yaml"},
		{"mixed_case", "taskfile.yml", "taskfile.dist.yaml"},
		{"all_lowercase", "taskfile.yaml", "taskfile.dist.yml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "both_"+tc.name)
			defer cleanupTaskManagerTestDir(testDir)

			// Create both dist and local files
			createTaskFile(t, testDir, tc.distFile, validDistTaskfileContent)
			createTaskFile(t, testDir, tc.localFile, validTaskfileContent)

			tm, err := manager.ParseTaskManager(&testDir)
			require.NoError(t, err, "Should parse task manager without error")
			require.NotNil(t, tm, "Should return a TaskManager instance")

			tasks, err := tm.ListTasks()
			require.NoError(t, err, "Should list tasks without error")

			// Should have tasks from both files (5 total: build, lint, test from local + clean, install from dist)
			expectedTasks := []string{"build", "clean", "install", "lint", "test"}
			assert.Len(t, tasks, len(expectedTasks), "Should return tasks from both files")

			taskNames := make([]string, len(tasks))
			for i, task := range tasks {
				taskNames[i] = task.Name
			}

			for _, expectedName := range expectedTasks {
				assert.Contains(t, taskNames, expectedName, "Should contain task %s", expectedName)
			}
		})
	}
}

func TestParseTaskManager_LocalOverridesDist(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "override_test")
	defer cleanupTaskManagerTestDir(testDir)

	// Create dist file with one version of a task
	distContent := `version: '3'

tasks:
  build:
    desc: "Build from dist"
    cmds:
      - echo "dist build"
`

	// Create local file with different version of same task
	localContent := `version: '3'

tasks:
  build:
    desc: "Build from local"
    cmds:
      - echo "local build"
`

	createTaskFile(t, testDir, "Taskfile.dist.yml", distContent)
	createTaskFile(t, testDir, "Taskfile.yml", localContent)

	tm, err := manager.ParseTaskManager(&testDir)
	require.NoError(t, err, "Should parse task manager without error")

	tasks, err := tm.ListTasks()
	require.NoError(t, err, "Should list tasks without error")

	require.Len(t, tasks, 1, "Should have exactly one task")

	// Local should override dist
	assert.Equal(t, "Build from local", tasks[0].Description, "Local task should override dist task")
}

func TestParseTaskManager_NoTaskfiles(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_taskfiles")
	defer cleanupTaskManagerTestDir(testDir)

	tm, err := manager.ParseTaskManager(&testDir)
	assert.NoError(t, err, "Should not return error when no taskfiles found")
	assert.Nil(t, tm, "Should return nil TaskManager when no taskfiles found")
}

func TestParseTaskManager_InvalidVersion(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "invalid_version")
	defer cleanupTaskManagerTestDir(testDir)

	createTaskFile(t, testDir, "Taskfile.yml", invalidTaskfileContent)

	tm, err := manager.ParseTaskManager(&testDir)
	assert.Error(t, err, "Should return error for unsupported version")
	assert.Nil(t, tm, "Should return nil TaskManager for invalid version")
}

func TestParseTaskManager_NoVersion(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_version")
	defer cleanupTaskManagerTestDir(testDir)

	createTaskFile(t, testDir, "Taskfile.yml", noVersionTaskfileContent)

	tm, err := manager.ParseTaskManager(&testDir)
	assert.Error(t, err, "Should return error when version is not specified")
	assert.Nil(t, tm, "Should return nil TaskManager when version is missing")
}

func TestParseTaskManager_EmptyTasks(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "empty_tasks")
	defer cleanupTaskManagerTestDir(testDir)

	emptyTasksContent := `version: '3'

tasks: {}
`

	createTaskFile(t, testDir, "Taskfile.yml", emptyTasksContent)

	tm, err := manager.ParseTaskManager(&testDir)
	require.NoError(t, err, "Should parse taskfile with empty tasks")
	require.NotNil(t, tm, "Should return TaskManager for empty tasks")

	tasks, err := tm.ListTasks()
	assert.NoError(t, err, "Should list empty tasks without error")
	assert.Empty(t, tasks, "Should return empty task list")
}

func TestParseTaskManager_NoTasksSection(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_tasks_section")
	defer cleanupTaskManagerTestDir(testDir)

	noTasksContent := `version: '3'
`

	createTaskFile(t, testDir, "Taskfile.yml", noTasksContent)

	tm, err := manager.ParseTaskManager(&testDir)
	require.NoError(t, err, "Should parse taskfile without tasks section")
	require.NotNil(t, tm, "Should return TaskManager without tasks section")

	tasks, err := tm.ListTasks()
	assert.NoError(t, err, "Should handle missing tasks section")
	assert.Empty(t, tasks, "Should return empty task list when no tasks section")
}

func TestTaskManager_GetTitle(t *testing.T) {
	testCases := []struct {
		name          string
		files         map[string]string
		expectedFiles []string
	}{
		{
			name: "single_local_file",
			files: map[string]string{
				"Taskfile.yml": validTaskfileContent,
			},
			expectedFiles: []string{"Taskfile.yml"},
		},
		{
			name: "single_dist_file",
			files: map[string]string{
				"Taskfile.dist.yml": validDistTaskfileContent,
			},
			expectedFiles: []string{"Taskfile.dist.yml"},
		},
		{
			name: "both_files",
			files: map[string]string{
				"Taskfile.dist.yml": validDistTaskfileContent,
				"Taskfile.yml":      validTaskfileContent,
			},
			expectedFiles: []string{"Taskfile.dist.yml", "Taskfile.yml"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "title_"+tc.name)
			defer cleanupTaskManagerTestDir(testDir)

			for filename, content := range tc.files {
				createTaskFile(t, testDir, filename, content)
			}

			tm, err := manager.ParseTaskManager(&testDir)
			require.NoError(t, err, "Should parse task manager without error")
			require.NotNil(t, tm, "Should return TaskManager")

			title := tm.GetTitle()
			assert.Equal(t, "task", title.Name, "Title name should be 'task'")
			assert.Contains(t, title.Description, "parsed from", "Title should contain base description")

			// Check that all expected files are mentioned in the title
			for _, expectedFile := range tc.expectedFiles {
				assert.Contains(t, title.Description, expectedFile, "Title should mention file %s", expectedFile)
			}
		})
	}
}

func TestTaskManager_ListTasks_Sorting(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "sorting_test")
	defer cleanupTaskManagerTestDir(testDir)

	unsortedTasksContent := `version: '3'

tasks:
  zebra:
    desc: "Last task alphabetically"
    cmds:
      - echo "zebra"

  alpha:
    desc: "First task alphabetically"
    cmds:
      - echo "alpha"

  beta:
    desc: "Middle task"
    cmds:
      - echo "beta"
`

	createTaskFile(t, testDir, "Taskfile.yml", unsortedTasksContent)

	tm, err := manager.ParseTaskManager(&testDir)
	require.NoError(t, err, "Should parse task manager without error")

	tasks, err := tm.ListTasks()
	require.NoError(t, err, "Should list tasks without error")

	expectedOrder := []string{"alpha", "beta", "zebra"}
	require.Len(t, tasks, len(expectedOrder), "Should return correct number of tasks")

	for i, expectedName := range expectedOrder {
		assert.Equal(t, expectedName, tasks[i].Name, "Task should be in correct order at index %d", i)
	}
}

func TestParseTaskManager_WithTestdata(t *testing.T) {
	tests := []struct {
		name          string
		testdataDir   string
		expectError   bool
		expectedTasks []string
	}{
		{
			name:          "local_only_testdata",
			testdataDir:   "testdata/local-only",
			expectError:   false,
			expectedTasks: []string{"build", "lint", "test"},
		},
		{
			name:          "dist_only_testdata",
			testdataDir:   "testdata/dist-only",
			expectError:   false,
			expectedTasks: []string{"clean", "install"},
		},
		{
			name:          "both_files_testdata",
			testdataDir:   "testdata/both-files",
			expectError:   false,
			expectedTasks: []string{"build", "clean", "install", "lint", "test"},
		},
		{
			name:        "invalid_version_testdata",
			testdataDir: "testdata/invalid-version",
			expectError: true,
		},
		{
			name:        "no_version_testdata",
			testdataDir: "testdata/no-version",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := manager.ParseTaskManager(&tt.testdataDir)

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid testdata")
				assert.Nil(t, tm, "Should return nil TaskManager for error case")
				return
			}

			require.NoError(t, err, "Should parse testdata without error")
			require.NotNil(t, tm, "Should return TaskManager for valid testdata")

			tasks, err := tm.ListTasks()
			require.NoError(t, err, "Should list tasks without error")

			assert.Len(t, tasks, len(tt.expectedTasks), "Should return correct number of tasks")

			for i, expectedName := range tt.expectedTasks {
				assert.Equal(t, expectedName, tasks[i].Name, "Task name should match at index %d", i)
			}

			// Verify title is not empty
			title := tm.GetTitle()
			assert.NotEmpty(t, title.Name, "Title name should not be empty")
		})
	}
}

func TestParseTaskManager_AllFilenameVariations(t *testing.T) {
	allFilenames := []struct {
		name     string
		filename string
		isLocal  bool
	}{
		{"Taskfile.yml", "Taskfile.yml", true},
		{"taskfile.yml", "taskfile.yml", true},
		{"Taskfile.yaml", "Taskfile.yaml", true},
		{"taskfile.yaml", "taskfile.yaml", true},
		{"Taskfile.dist.yml", "Taskfile.dist.yml", false},
		{"taskfile.dist.yml", "taskfile.dist.yml", false},
		{"Taskfile.dist.yaml", "Taskfile.dist.yaml", false},
		{"taskfile.dist.yaml", "taskfile.dist.yaml", false},
	}

	for _, file := range allFilenames {
		t.Run(file.name, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "filename_"+file.name)
			defer cleanupTaskManagerTestDir(testDir)

			var content string
			if file.isLocal {
				content = validTaskfileContent
			} else {
				content = validDistTaskfileContent
			}

			createTaskFile(t, testDir, file.filename, content)

			tm, err := manager.ParseTaskManager(&testDir)
			require.NoError(t, err, "Should parse %s without error", file.filename)
			require.NotNil(t, tm, "Should return TaskManager for %s", file.filename)

			tasks, err := tm.ListTasks()
			require.NoError(t, err, "Should list tasks for %s", file.filename)
			assert.NotEmpty(t, tasks, "Should return non-empty task list for %s", file.filename)

			title := tm.GetTitle()
			assert.NotEmpty(t, title.Name, "Title should not be empty for %s", file.filename)
			assert.Contains(t, title.Description, file.filename, "Title should mention the filename")
		})
	}
}
