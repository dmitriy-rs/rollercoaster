package task_test

import (
	"reflect"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

func TestSortTasks(t *testing.T) {
	tests := []struct {
		name     string
		input    []task.Task
		expected []task.Task
	}{
		{
			name:     "empty slice",
			input:    []task.Task{},
			expected: []task.Task{},
		},
		{
			name: "single task",
			input: []task.Task{
				{Name: "Task A", Description: "First task"},
			},
			expected: []task.Task{
				{Name: "Task A", Description: "First task"},
			},
		},
		{
			name: "already sorted tasks",
			input: []task.Task{
				{Name: "A Task", Description: "First task"},
				{Name: "B Task", Description: "Second task"},
				{Name: "C Task", Description: "Third task"},
			},
			expected: []task.Task{
				{Name: "A Task", Description: "First task"},
				{Name: "B Task", Description: "Second task"},
				{Name: "C Task", Description: "Third task"},
			},
		},
		{
			name: "reverse sorted tasks",
			input: []task.Task{
				{Name: "Z Task", Description: "Last task"},
				{Name: "B Task", Description: "Second task"},
				{Name: "A Task", Description: "First task"},
			},
			expected: []task.Task{
				{Name: "A Task", Description: "First task"},
				{Name: "B Task", Description: "Second task"},
				{Name: "Z Task", Description: "Last task"},
			},
		},
		{
			name: "mixed case sorting",
			input: []task.Task{
				{Name: "zebra", Description: "Lowercase z"},
				{Name: "Apple", Description: "Uppercase A"},
				{Name: "banana", Description: "Lowercase b"},
			},
			expected: []task.Task{
				{Name: "Apple", Description: "Uppercase A"},
				{Name: "banana", Description: "Lowercase b"},
				{Name: "zebra", Description: "Lowercase z"},
			},
		},
		{
			name: "tasks with same name prefix",
			input: []task.Task{
				{Name: "Task 10", Description: "Tenth task"},
				{Name: "Task 2", Description: "Second task"},
				{Name: "Task 1", Description: "First task"},
			},
			expected: []task.Task{
				{Name: "Task 1", Description: "First task"},
				{Name: "Task 10", Description: "Tenth task"},
				{Name: "Task 2", Description: "Second task"},
			},
		},
		{
			name: "tasks with special characters",
			input: []task.Task{
				{Name: "Task-C", Description: "Task with dash"},
				{Name: "Task A", Description: "Task with space"},
				{Name: "Task_B", Description: "Task with underscore"},
			},
			expected: []task.Task{
				{Name: "Task A", Description: "Task with space"},
				{Name: "Task-C", Description: "Task with dash"},
				{Name: "Task_B", Description: "Task with underscore"},
			},
		},
		{
			name: "duplicate task names",
			input: []task.Task{
				{Name: "Duplicate", Description: "Second occurrence"},
				{Name: "Duplicate", Description: "First occurrence"},
				{Name: "Another", Description: "Different task"},
			},
			expected: []task.Task{
				{Name: "Another", Description: "Different task"},
				{Name: "Duplicate", Description: "Second occurrence"},
				{Name: "Duplicate", Description: "First occurrence"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of input to avoid modifying the test data
			input := make([]task.Task, len(tt.input))
			copy(input, tt.input)

			task.SortTasks(input)

			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("SortTasks() = %v, want %v", input, tt.expected)
			}
		})
	}
}

func TestSortTasksStability(t *testing.T) {
	// Test that the sort is stable (maintains relative order of equal elements)
	tasks := []task.Task{
		{Name: "Same", Description: "First"},
		{Name: "Different", Description: "Middle"},
		{Name: "Same", Description: "Second"},
	}

	expected := []task.Task{
		{Name: "Different", Description: "Middle"},
		{Name: "Same", Description: "First"},
		{Name: "Same", Description: "Second"},
	}

	task.SortTasks(tasks)

	if !reflect.DeepEqual(tasks, expected) {
		t.Errorf("SortTasks() stability test failed. Got %v, want %v", tasks, expected)
	}
}

func TestSortTasksModifiesOriginalSlice(t *testing.T) {
	// Test that SortTasks modifies the original slice in place
	original := []task.Task{
		{Name: "C", Description: "Third"},
		{Name: "A", Description: "First"},
		{Name: "B", Description: "Second"},
	}

	// Keep a reference to the original slice
	sliceToSort := original

	task.SortTasks(sliceToSort)

	// Verify the original slice was modified
	expected := []task.Task{
		{Name: "A", Description: "First"},
		{Name: "B", Description: "Second"},
		{Name: "C", Description: "Third"},
	}

	if !reflect.DeepEqual(original, expected) {
		t.Errorf("SortTasks() should modify original slice. Got %v, want %v", original, expected)
	}
}
