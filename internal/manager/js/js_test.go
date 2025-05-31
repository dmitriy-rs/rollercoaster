package jsmanager_test

import (
	"path/filepath"
	"testing"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/js/mocks"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJsManager(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		workspace   *jsmanager.JsWorkspace
		wantNil     bool
		wantError   bool
		wantScripts int
	}{
		{
			name:        "empty directory (no package.json)",
			testdataDir: "empty",
			workspace:   mocks.NewMockJsWorkspace("npm").ToJsWorkspacePtr(),
			wantNil:     true,
			wantError:   false,
			wantScripts: 0,
		},
		{
			name:        "package.json without scripts",
			testdataDir: "no-scripts",
			workspace:   mocks.NewMockJsWorkspace("npm").ToJsWorkspacePtr(),
			wantNil:     false,
			wantError:   false,
			wantScripts: 0,
		},
		{
			name:        "package.json with scripts",
			testdataDir: "with-scripts",
			workspace:   mocks.NewMockJsWorkspace("npm").ToJsWorkspacePtr(),
			wantNil:     false,
			wantError:   false,
			wantScripts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			manager, err := jsmanager.ParseJsManager(&testDir, tt.workspace)

			if tt.wantError {
				assert.Error(t, err, "ParseJsManager() should return error for %s", tt.testdataDir)
				assert.Nil(t, manager, "ParseJsManager() should return nil manager when error occurs")
				return
			}

			assert.NoError(t, err, "ParseJsManager() should not return error for %s", tt.testdataDir)

			if tt.wantNil {
				assert.Nil(t, manager, "ParseJsManager() should return nil for %s", tt.testdataDir)
				return
			}

			require.NotNil(t, manager, "ParseJsManager() should return manager for %s", tt.testdataDir)

			// Test ListTasks
			tasks, err := manager.ListTasks()
			require.NoError(t, err, "ListTasks() should not return error")
			assert.Len(t, tasks, tt.wantScripts, "Should have correct number of scripts")

			// Test GetTitle
			title := manager.GetTitle()
			assert.Equal(t, (*tt.workspace).Name(), title.Name, "Title name should match workspace name")
			assert.Contains(t, title.Description, "package.json", "Title description should mention package.json")
		})
	}
}

func TestJsManager_ListTasks(t *testing.T) {
	testDir := filepath.Join("testdata", "with-scripts")
	mockWorkspace := mocks.NewMockJsWorkspace("npm")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager, err := jsmanager.ParseJsManager(&testDir, workspace)
	require.NoError(t, err, "ParseJsManager() should not return error")
	require.NotNil(t, manager, "ParseJsManager() should return manager")

	tasks, err := manager.ListTasks()
	require.NoError(t, err, "ListTasks() should not return error")
	require.Len(t, tasks, 3, "Should have 3 tasks")

	// Check that we have the expected tasks
	taskNames := make(map[string]string)
	for _, task := range tasks {
		taskNames[task.Name] = task.Description
	}

	assert.Equal(t, "echo 'Starting application'", taskNames["start"], "start task should have correct description")
	assert.Equal(t, "echo 'Running tests'", taskNames["test"], "test task should have correct description")
	assert.Equal(t, "echo 'Building application'", taskNames["build"], "build task should have correct description")
}

func TestJsManager_ExecuteTask(t *testing.T) {
	testDir := filepath.Join("testdata", "with-scripts")
	mockWorkspace := mocks.NewMockJsWorkspace("npm")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager, err := jsmanager.ParseJsManager(&testDir, workspace)
	require.NoError(t, err, "ParseJsManager() should not return error")
	require.NotNil(t, manager, "ParseJsManager() should return manager")

	testTask := &task.Task{
		Name:        "test",
		Description: "echo 'Running tests'",
	}

	// Clear any previous commands
	mockWorkspace.ClearExecutedCommands()

	// Execute the task
	manager.ExecuteTask(testTask, "arg1", "arg2")

	// Check that the command was executed
	executedCmds := mockWorkspace.GetExecutedCommands()
	require.Len(t, executedCmds, 1, "Should have executed exactly one command")

	// The command should be echo with the workspace name
	expectedCmd := []string{"echo", "npm"}
	assert.Equal(t, expectedCmd, executedCmds[0], "Should execute the correct command")
}

func TestJsManager_GetTitle(t *testing.T) {
	tests := []struct {
		name          string
		workspaceName string
		expectedName  string
	}{
		{
			name:          "npm workspace",
			workspaceName: "npm",
			expectedName:  "npm",
		},
		{
			name:          "yarn workspace",
			workspaceName: "yarn",
			expectedName:  "yarn",
		},
		{
			name:          "pnpm workspace",
			workspaceName: "pnpm",
			expectedName:  "pnpm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", "with-scripts")
			mockWorkspace := mocks.NewMockJsWorkspace(tt.workspaceName)
			workspace := mockWorkspace.ToJsWorkspacePtr()

			manager, err := jsmanager.ParseJsManager(&testDir, workspace)
			require.NoError(t, err, "ParseJsManager() should not return error")
			require.NotNil(t, manager, "ParseJsManager() should return manager")

			title := manager.GetTitle()
			assert.Equal(t, tt.expectedName, title.Name, "Title name should match workspace name")
			assert.Contains(t, title.Description, "package.json", "Title description should be correct")
		})
	}
}

func TestJsManager_ExecuteTaskWithMultipleArgs(t *testing.T) {
	testDir := filepath.Join("testdata", "with-scripts")
	mockWorkspace := mocks.NewMockJsWorkspace("yarn")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager, err := jsmanager.ParseJsManager(&testDir, workspace)
	require.NoError(t, err, "ParseJsManager() should not return error")
	require.NotNil(t, manager, "ParseJsManager() should return manager")

	testTask := &task.Task{
		Name:        "build",
		Description: "echo 'Building application'",
	}

	// Clear any previous commands
	mockWorkspace.ClearExecutedCommands()

	// Execute the task with multiple arguments
	manager.ExecuteTask(testTask, "--verbose", "--production", "extra-arg")

	// Check that the command was executed
	executedCmds := mockWorkspace.GetExecutedCommands()
	require.Len(t, executedCmds, 1, "Should have executed exactly one command")

	// The command should be echo with the workspace name
	expectedCmd := []string{"echo", "yarn"}
	assert.Equal(t, expectedCmd, executedCmds[0], "Should execute the correct command")
}

func TestJsManager_ExecuteTaskNoArgs(t *testing.T) {
	testDir := filepath.Join("testdata", "with-scripts")
	mockWorkspace := mocks.NewMockJsWorkspace("pnpm")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager, err := jsmanager.ParseJsManager(&testDir, workspace)
	require.NoError(t, err, "ParseJsManager() should not return error")
	require.NotNil(t, manager, "ParseJsManager() should return manager")

	testTask := &task.Task{
		Name:        "start",
		Description: "echo 'Starting application'",
	}

	// Clear any previous commands
	mockWorkspace.ClearExecutedCommands()

	// Execute the task without arguments
	manager.ExecuteTask(testTask)

	// Check that the command was executed
	executedCmds := mockWorkspace.GetExecutedCommands()
	require.Len(t, executedCmds, 1, "Should have executed exactly one command")

	// The command should be echo with the workspace name
	expectedCmd := []string{"echo", "pnpm"}
	assert.Equal(t, expectedCmd, executedCmds[0], "Should execute the correct command")
}

func TestMockJsWorkspace_AllMethods(t *testing.T) {
	workspace := mocks.NewMockJsWorkspace("test-workspace")

	// Test Name
	assert.Equal(t, "test-workspace", workspace.Name(), "Name() should return correct name")

	// Test all command methods
	workspace.ClearExecutedCommands()

	// Test Cmd
	cmd := workspace.Cmd()
	assert.Equal(t, []string{"echo", "test-workspace"}, cmd.Args, "Cmd() should return correct args")

	// Test ExecuteCmd
	runCmd := workspace.ExecuteCmd()
	assert.Equal(t, []string{"echo", "test-workspace", "run"}, runCmd.Args, "ExecuteCmd() should return correct args")

	// Test InstallCmd
	installCmd := workspace.InstallCmd()
	assert.Equal(t, []string{"echo", "test-workspace", "install"}, installCmd.Args, "InstallCmd() should return correct args")

	// Test AddCmd
	addCmd := workspace.AddCmd()
	assert.Equal(t, []string{"echo", "test-workspace", "add"}, addCmd.Args, "AddCmd() should return correct args")

	// Test RemoveCmd
	removeCmd := workspace.RemoveCmd()
	assert.Equal(t, []string{"echo", "test-workspace", "remove"}, removeCmd.Args, "RemoveCmd() should return correct args")

	// Check that all commands were tracked
	executedCmds := workspace.GetExecutedCommands()
	assert.Len(t, executedCmds, 5, "Should have tracked all 5 command executions")
}
