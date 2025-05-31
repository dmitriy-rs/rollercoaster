package ui

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/stretchr/testify/assert"
)

// Mock item for testing that implements list.Item but is not a managerTaskItem
type mockListItem struct{}

func (m mockListItem) Title() string       { return "mock" }
func (m mockListItem) FilterValue() string { return "mock" }

// Mock manager for testing
type mockManager struct {
	title manager.Title
}

func (m *mockManager) GetTitle() manager.Title                     { return m.title }
func (m *mockManager) ListTasks() ([]task.Task, error)             { return nil, nil }
func (m *mockManager) ExecuteTask(task *task.Task, args ...string) {}

func TestManagerTaskItem(t *testing.T) {
	tests := []struct {
		name         string
		task         task.Task
		expectTitle  string
		expectFilter string
	}{
		{
			name: "task with aliases",
			task: task.Task{
				Name:    "build",
				Aliases: []string{"b", "compile"},
			},
			expectTitle:  "b",
			expectFilter: "build",
		},
		{
			name: "task without aliases",
			task: task.Task{
				Name:    "test",
				Aliases: []string{},
			},
			expectTitle:  "test",
			expectFilter: "test",
		},
		{
			name: "task with nil aliases",
			task: task.Task{
				Name:    "deploy",
				Aliases: nil,
			},
			expectTitle:  "deploy",
			expectFilter: "deploy",
		},
	}

	var mgr manager.Manager = &mockManager{title: manager.Title{Name: "test", Description: "Test manager"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managerTask := manager.ManagerTask{
				Task:    tt.task,
				Manager: &mgr,
			}
			item := managerTaskItem{ManagerTask: managerTask}

			assert.Equal(t, tt.expectTitle, item.Title())
			assert.Equal(t, tt.expectFilter, item.FilterValue())
		})
	}
}

func TestItemDelegate(t *testing.T) {
	var mgr manager.Manager = &mockManager{title: manager.Title{Name: "task", Description: "Taskfile runner"}}

	t.Run("basic properties", func(t *testing.T) {
		delegate := itemDelegate{
			showManagerIndicator: true,
		}

		assert.Equal(t, 1, delegate.Height())
		assert.Equal(t, 0, delegate.Spacing())
		assert.Nil(t, delegate.Update(nil, nil))
	})

	t.Run("render method without manager indicator", func(t *testing.T) {
		delegate := itemDelegate{
			showManagerIndicator: false,
		}

		managerTask := manager.ManagerTask{
			Task:    task.Task{Name: "build", Description: "Build application"},
			Manager: &mgr,
		}

		tasks := []list.Item{
			managerTaskItem{ManagerTask: managerTask},
		}

		// Create a mock list model
		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "build")
		assert.Contains(t, output, "Build application")
		assert.NotContains(t, output, "[task]") // No manager indicator
	})

	t.Run("render with manager indicator", func(t *testing.T) {
		delegate := itemDelegate{
			showManagerIndicator: true,
		}

		managerTask := manager.ManagerTask{
			Task:    task.Task{Name: "build", Description: "Build application"},
			Manager: &mgr,
		}

		tasks := []list.Item{
			managerTaskItem{ManagerTask: managerTask},
		}

		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "build")
		assert.Contains(t, output, "[task]") // Manager indicator should be present
	})

	t.Run("render with long descriptions", func(t *testing.T) {
		delegate := itemDelegate{
			showManagerIndicator: false,
		}

		longDescription := "This is a very long description that should be truncated to ensure proper formatting and alignment in the UI"
		managerTask := manager.ManagerTask{
			Task:    task.Task{Name: "build", Description: longDescription},
			Manager: &mgr,
		}

		tasks := []list.Item{
			managerTaskItem{ManagerTask: managerTask},
		}

		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "...")
	})

	t.Run("render non-managerTaskItem", func(t *testing.T) {
		delegate := itemDelegate{}

		var buf bytes.Buffer
		delegate.Render(&buf, list.Model{}, 0, mockListItem{})

		// Should not write anything for non-managerTaskItem
		assert.Empty(t, buf.String())
	})

	t.Run("render with long task name", func(t *testing.T) {
		delegate := itemDelegate{
			showManagerIndicator: false,
		}

		managerTask := manager.ManagerTask{
			Task:    task.Task{Name: "very-long-task-name-that-exceeds-normal-length", Description: "Test description"},
			Manager: &mgr,
		}

		tasks := []list.Item{
			managerTaskItem{ManagerTask: managerTask},
		}

		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "...") // Long name should be truncated
	})
}
