package taskmanager_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	manager "github.com/dmitriy-rs/rollercoaster/internal/manager/task"
)

func setupTaskManagerTestDir(t *testing.T, dirName string) string {
	testDir := filepath.Join(os.TempDir(), "task-manager-test", dirName)
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return testDir
}

func cleanupTaskManagerTestDir(testDir string) {
	_ = os.RemoveAll(filepath.Dir(testDir))
}

func createTaskFile(t *testing.T, dir, filename, content string) {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create task file %s: %v", filename, err)
	}
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
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
			}

			tasks, err := tm.ListTasks()
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}

			expectedTasks := []string{"build", "lint", "test"}
			if len(tasks) != len(expectedTasks) {
				t.Fatalf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
			}

			for i, expectedName := range expectedTasks {
				if tasks[i].Name != expectedName {
					t.Errorf("Expected task name %s, got %s", expectedName, tasks[i].Name)
				}
			}

			title := tm.GetTitle()
			if title.Name == "" {
				t.Error("Expected non-empty title")
			}
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
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
			}

			tasks, err := tm.ListTasks()
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}

			expectedTasks := []string{"clean", "install"}
			if len(tasks) != len(expectedTasks) {
				t.Fatalf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
			}

			for i, expectedName := range expectedTasks {
				if tasks[i].Name != expectedName {
					t.Errorf("Expected task name %s, got %s", expectedName, tasks[i].Name)
				}
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
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
			}

			tasks, err := tm.ListTasks()
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}

			// Should have tasks from both files: build, clean, install, lint, test
			expectedTasks := []string{"build", "clean", "install", "lint", "test"}
			if len(tasks) != len(expectedTasks) {
				t.Fatalf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
			}

			for i, expectedName := range expectedTasks {
				if tasks[i].Name != expectedName {
					t.Errorf("Expected task name %s, got %s", expectedName, tasks[i].Name)
				}
			}

			// Check that title mentions both files
			title := tm.GetTitle()
			if title.Name == "" {
				t.Error("Expected non-empty title")
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
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	tasks, err := tm.ListTasks()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	// Local should override dist
	if tasks[0].Description != "Build from local" {
		t.Errorf("Expected local task description, got: %s", tasks[0].Description)
	}
}

func TestParseTaskManager_NoTaskfiles(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_taskfiles")
	defer cleanupTaskManagerTestDir(testDir)

	tm, err := manager.ParseTaskManager(&testDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if tm != nil {
		t.Error("Expected nil TaskManager when no taskfiles found")
	}
}

func TestParseTaskManager_InvalidVersion(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "invalid_version")
	defer cleanupTaskManagerTestDir(testDir)

	createTaskFile(t, testDir, "Taskfile.yml", invalidTaskfileContent)

	tm, err := manager.ParseTaskManager(&testDir)
	if err == nil {
		t.Fatal("Expected error for unsupported version")
	}

	if tm != nil {
		t.Error("Expected nil TaskManager for invalid version")
	}
}

func TestParseTaskManager_NoVersion(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_version")
	defer cleanupTaskManagerTestDir(testDir)

	createTaskFile(t, testDir, "Taskfile.yml", noVersionTaskfileContent)

	tm, err := manager.ParseTaskManager(&testDir)
	if err == nil {
		t.Fatal("Expected error for missing version")
	}

	if tm != nil {
		t.Error("Expected nil TaskManager for missing version")
	}
}

func TestParseTaskManager_EmptyTasks(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "empty_tasks")
	defer cleanupTaskManagerTestDir(testDir)

	emptyTasksContent := `version: '3'

tasks: {}
`

	createTaskFile(t, testDir, "Taskfile.yml", emptyTasksContent)

	tm, err := manager.ParseTaskManager(&testDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if tm == nil {
		t.Fatal("Expected TaskManager, got nil")
	}

	tasks, err := tm.ListTasks()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseTaskManager_NoTasksSection(t *testing.T) {
	testDir := setupTaskManagerTestDir(t, "no_tasks_section")
	defer cleanupTaskManagerTestDir(testDir)

	noTasksContent := `version: '3'
`

	createTaskFile(t, testDir, "Taskfile.yml", noTasksContent)

	tm, err := manager.ParseTaskManager(&testDir)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if tm == nil {
		t.Fatal("Expected TaskManager, got nil")
	}

	tasks, err := tm.ListTasks()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestTaskManager_GetTitle(t *testing.T) {
	tests := []struct {
		name      string
		filenames []string
	}{
		{
			name:      "no_files",
			filenames: []string{},
		},
		{
			name:      "single_file",
			filenames: []string{"Taskfile.yml"},
		},
		{
			name:      "multiple_files",
			filenames: []string{"Taskfile.dist.yml", "Taskfile.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "title_test_"+tt.name)
			defer cleanupTaskManagerTestDir(testDir)

			// Create files if specified
			for _, filename := range tt.filenames {
				createTaskFile(t, testDir, filename, validTaskfileContent)
			}

			tm, err := manager.ParseTaskManager(&testDir)
			if len(tt.filenames) == 0 {
				if tm != nil {
					t.Error("Expected nil TaskManager for no files")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
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
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	tasks, err := tm.ListTasks()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	expectedOrder := []string{"alpha", "beta", "zebra"}
	if len(tasks) != len(expectedOrder) {
		t.Fatalf("Expected %d tasks, got %d", len(expectedOrder), len(tasks))
	}

	for i, expectedName := range expectedOrder {
		if tasks[i].Name != expectedName {
			t.Errorf("Expected task %d to be %s, got %s", i, expectedName, tasks[i].Name)
		}
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
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tm != nil {
					t.Error("Expected nil TaskManager for error case")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
			}

			tasks, err := tm.ListTasks()
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}

			if len(tasks) != len(tt.expectedTasks) {
				t.Fatalf("Expected %d tasks, got %d", len(tt.expectedTasks), len(tasks))
			}

			for i, expectedName := range tt.expectedTasks {
				if tasks[i].Name != expectedName {
					t.Errorf("Expected task %d to be %s, got %s", i, expectedName, tasks[i].Name)
				}
			}

			// Verify title is not empty
			title := tm.GetTitle()
			if title.Name == "" {
				t.Error("Expected non-empty title")
			}
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

	for _, tf := range allFilenames {
		t.Run(tf.name, func(t *testing.T) {
			testDir := setupTaskManagerTestDir(t, "filename_variation_"+tf.name)
			defer cleanupTaskManagerTestDir(testDir)

			content := validTaskfileContent
			if !tf.isLocal {
				content = validDistTaskfileContent
			}

			createTaskFile(t, testDir, tf.filename, content)

			tm, err := manager.ParseTaskManager(&testDir)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tm == nil {
				t.Fatal("Expected TaskManager, got nil")
			}

			tasks, err := tm.ListTasks()
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}

			if len(tasks) == 0 {
				t.Error("Expected at least one task")
			}

			// Verify the title contains the filename
			title := tm.GetTitle()
			if !strings.Contains(title.Description, tf.filename) {
				t.Errorf("Expected title to contain %s, got: %s", tf.filename, title.Description)
			}
		})
	}
}
