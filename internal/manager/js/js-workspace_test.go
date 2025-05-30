package jsmanager_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/js/mocks"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

func TestParseJsWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		testdataDir string
		wantName    string
		wantNil     bool
		wantError   bool
	}{
		{
			name:        "npm with lock",
			testdataDir: "npm-with-lock",
			wantName:    "npm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "yarn v1 with lock",
			testdataDir: "yarn-v1-with-lock",
			wantName:    "yarn",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "yarn v2+ with lock (unsupported)",
			testdataDir: "yarn-v2-with-lock",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "pnpm lock v6",
			testdataDir: "pnpm-lock-v6",
			wantName:    "pnpm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm lock v9",
			testdataDir: "pnpm-lock-v9",
			wantName:    "pnpm",
			wantNil:     false,
			wantError:   false,
		},
		{
			name:        "pnpm with random lock (unsupported)",
			testdataDir: "pnpm-random-lock",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "empty directory",
			testdataDir: "empty",
			wantName:    "",
			wantNil:     true,
			wantError:   false,
		},
		{
			name:        "multiple locks (error)",
			testdataDir: "multiple-locks",
			wantName:    "",
			wantNil:     true,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join("testdata", tt.testdataDir)
			workspace, err := jsmanager.ParseJsWorkspace(&testDir)

			if tt.wantError {
				assert.Error(t, err, "ParseJsWorkspace() should return an error")
				return
			}

			require.NoError(t, err, "ParseJsWorkspace() should not return an error")

			if tt.wantNil {
				assert.Nil(t, workspace, "ParseJsWorkspace() should return nil workspace")
				return
			}

			require.NotNil(t, workspace, "ParseJsWorkspace() should return a workspace")
			assert.Contains(t, (*workspace).Name(), tt.wantName, "workspace name should match expected")
		})
	}
}

func TestJsWorkspaceManager_ListTasks(t *testing.T) {
	mockWorkspace := mocks.NewMockJsWorkspaceWithExecName("test-workspace", "testx")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager := &jsmanager.JsWorkspaceManager{
		Workspace: workspace,
	}

	tasks, err := manager.ListTasks()
	require.NoError(t, err, "ListTasks() should not return an error")
	require.Len(t, tasks, 4, "Should return exactly 4 tasks")

	expectedTasks := []struct {
		name        string
		description string
		aliases     []string
	}{
		{
			name:        "add",
			description: "Install a dependency",
			aliases:     nil,
		},
		{
			name:        "remove",
			description: "Remove a dependency",
			aliases:     nil,
		},
		{
			name:        "install",
			description: "Install dependencies",
			aliases:     []string{"i"},
		},
		// {
		// 	name:        "x",
		// 	description: "testx Execute a command",
		// 	aliases:     nil,
		// },
	}

	for i, expected := range expectedTasks {
		assert.Equal(t, expected.name, tasks[i].Name, "Task name should match at index %d", i)
		assert.Equal(t, expected.description, tasks[i].Description, "Task description should match at index %d", i)
		assert.Equal(t, expected.aliases, tasks[i].Aliases, "Task aliases should match at index %d", i)
	}
}

func TestJsWorkspaceManager_ExecuteTask(t *testing.T) {
	mockWorkspace := mocks.NewMockJsWorkspaceWithExecName("test-workspace", "testx")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager := &jsmanager.JsWorkspaceManager{
		Workspace: workspace,
	}

	tests := []struct {
		name        string
		taskName    string
		expectedCmd []string
	}{
		{
			name:        "install task",
			taskName:    "install",
			expectedCmd: []string{"echo", "test-workspace", "install"},
		},
		{
			name:        "add task",
			taskName:    "add",
			expectedCmd: []string{"echo", "test-workspace", "add"},
		},
		{
			name:        "remove task",
			taskName:    "remove",
			expectedCmd: []string{"echo", "test-workspace", "remove"},
		},
		// {
		// 	name:        "execute task",
		// 	taskName:    "x",
		// 	expectedCmd: []string{"echo", "test-workspace", "run"},
		// },
		{
			name:        "default task",
			taskName:    "unknown",
			expectedCmd: []string{"echo", "test-workspace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWorkspace.ClearExecutedCommands()

			testTask := &task.Task{
				Name:        tt.taskName,
				Description: "Test task",
			}

			manager.ExecuteTask(testTask, "arg1", "arg2")

			executedCmds := mockWorkspace.GetExecutedCommands()
			require.Len(t, executedCmds, 1, "Should have executed exactly one command")
			assert.Equal(t, tt.expectedCmd, executedCmds[0], "Should execute the correct command")
		})
	}
}

func TestJsWorkspaceManager_GetTitle(t *testing.T) {
	mockWorkspace := mocks.NewMockJsWorkspaceWithExecName("test-workspace", "testx")
	workspace := mockWorkspace.ToJsWorkspacePtr()

	manager := &jsmanager.JsWorkspaceManager{
		Workspace: workspace,
	}

	title := manager.GetTitle()

	assert.Equal(t, "test-workspace", title.Name, "Title name should match workspace name")
	assert.Equal(t, "package manager commands", title.Description, "Title description should be correct")
}
