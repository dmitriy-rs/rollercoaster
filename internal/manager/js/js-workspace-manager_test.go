package jsmanager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsmanager "github.com/dmitriy-rs/rollercoaster/internal/manager/js"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/js/mocks"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

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
	assert.Contains(t, title.Description, "package", "Title description should be correct")
}
