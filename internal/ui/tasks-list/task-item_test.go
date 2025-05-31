package ui

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/stretchr/testify/assert"
)

// Mock item for testing that implements list.Item but is not a taskItem
type mockListItem struct{}

func (m mockListItem) Title() string       { return "mock" }
func (m mockListItem) FilterValue() string { return "mock" }

func TestTaskItem(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := taskItem(tt.task)

			assert.Equal(t, tt.expectTitle, item.Title())
			assert.Equal(t, tt.expectFilter, item.FilterValue())
		})
	}
}

func TestItemDelegate(t *testing.T) {
	managerTitles := []manager.Title{
		{Name: "task", Description: "Taskfile runner"},
		{Name: "npm", Description: "Node.js package manager"},
	}
	taskCounts := []int{2, 2}
	managerStartIndices := []int{0, 2}

	t.Run("basic properties", func(t *testing.T) {
		delegate := itemDelegate{
			managerTitles:        managerTitles,
			taskCounts:           taskCounts,
			managerStartIndices:  managerStartIndices,
			showManagerIndicator: true,
		}

		assert.Equal(t, 1, delegate.Height())
		assert.Equal(t, 0, delegate.Spacing())
		assert.Nil(t, delegate.Update(nil, nil))
	})

	t.Run("render method", func(t *testing.T) {
		delegate := itemDelegate{
			managerTitles:        managerTitles,
			taskCounts:           taskCounts,
			managerStartIndices:  managerStartIndices,
			showManagerIndicator: false,
		}

		tasks := []list.Item{
			taskItem(task.Task{Name: "build", Description: "Build application"}),
			taskItem(task.Task{Name: "test", Description: "Run tests"}),
		}

		// Create a mock list model
		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "build")
		assert.Contains(t, output, "Build application")
	})

	t.Run("render with manager indicator", func(t *testing.T) {
		delegate := itemDelegate{
			managerTitles:        managerTitles,
			taskCounts:           taskCounts,
			managerStartIndices:  managerStartIndices,
			showManagerIndicator: true,
		}

		tasks := []list.Item{
			taskItem(task.Task{Name: "build", Description: "Build application"}),
			taskItem(task.Task{Name: "test", Description: "Run tests"}),
		}

		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "build")
		assert.Contains(t, output, "[task]") // Manager indicator
	})

	t.Run("render with long descriptions", func(t *testing.T) {
		delegate := itemDelegate{
			managerTitles:        managerTitles,
			taskCounts:           taskCounts,
			managerStartIndices:  managerStartIndices,
			showManagerIndicator: false,
		}

		longDescription := "This is a very long description that should be truncated to ensure proper formatting and alignment in the UI"
		tasks := []list.Item{
			taskItem(task.Task{Name: "build", Description: longDescription}),
		}

		listModel := list.New(tasks, delegate, 80, 10)

		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])

		output := buf.String()
		assert.Contains(t, output, "...")
	})

	t.Run("render non-taskItem", func(t *testing.T) {
		delegate := itemDelegate{}

		var buf bytes.Buffer
		delegate.Render(&buf, list.Model{}, 0, mockListItem{})

		// Should not write anything for non-taskItem
		assert.Empty(t, buf.String())
	})
}
