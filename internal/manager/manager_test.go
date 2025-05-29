package manager_test

import (
	"fmt"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

// MockManager implements the Manager interface for testing
type MockManager struct {
	title     string
	tasks     []task.Task
	listError error
	executed  []ExecutedTask
}

type ExecutedTask struct {
	Task *task.Task
	Args []string
}

func NewMockManager(title string, tasks []task.Task) *MockManager {
	return &MockManager{
		title:    title,
		tasks:    tasks,
		executed: make([]ExecutedTask, 0),
	}
}

func (m *MockManager) GetTitle() manager.Title {
	return manager.Title{
		Name:        m.title,
		Description: "",
	}
}

func (m *MockManager) ListTasks() ([]task.Task, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tasks, nil
}

func (m *MockManager) ExecuteTask(task *task.Task, args ...string) {
	m.executed = append(m.executed, ExecutedTask{
		Task: task,
		Args: args,
	})
}

func (m *MockManager) SetListError(err error) {
	m.listError = err
}

func (m *MockManager) GetExecutedTasks() []ExecutedTask {
	return m.executed
}

func (m *MockManager) ClearExecutedTasks() {
	m.executed = make([]ExecutedTask, 0)
}

func TestFindClosestTask_ExactMatch(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "lint", Description: "Run linters"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	result, err := manager.FindClosestTask(mockManager, "build")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected task, got nil")
	}

	if result.Name != "build" {
		t.Errorf("Expected task name 'build', got '%s'", result.Name)
	}

	if result.Description != "Build the application" {
		t.Errorf("Expected description 'Build the application', got '%s'", result.Description)
	}
}

func TestFindClosestTask_FuzzyMatch(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "lint", Description: "Run linters"},
		{Name: "deploy", Description: "Deploy the application"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "partial match",
			input:    "bui",
			expected: "build",
		},
		{
			name:     "substring match",
			input:    "dep",
			expected: "deploy",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := manager.FindClosestTask(mockManager, tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Expected task, got nil")
			}

			if result.Name != tc.expected {
				t.Errorf("Expected task name '%s', got '%s'", tc.expected, result.Name)
			}
		})
	}
}

func TestFindClosestTask_NoMatch(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "lint", Description: "Run linters"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	result, err := manager.FindClosestTask(mockManager, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}

	if result != nil {
		t.Error("Expected nil result for non-existent task")
	}

	expectedError := "no task found for 'nonexistent'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFindClosestTask_EmptyTaskList(t *testing.T) {
	tasks := []task.Task{}

	mockManager := NewMockManager("Empty Manager", tasks)

	result, err := manager.FindClosestTask(mockManager, "anything")
	if err == nil {
		t.Fatal("Expected error for empty task list")
	}

	if result != nil {
		t.Error("Expected nil result for empty task list")
	}

	expectedError := "no task found for 'anything'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFindClosestTask_ListTasksError(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
	}

	mockManager := NewMockManager("Error Manager", tasks)
	expectedError := fmt.Errorf("failed to list tasks")
	mockManager.SetListError(expectedError)

	result, err := manager.FindClosestTask(mockManager, "build")
	if err == nil {
		t.Fatal("Expected error when ListTasks fails")
	}

	if result != nil {
		t.Error("Expected nil result when ListTasks fails")
	}

	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

func TestFindClosestTask_MultipleMatches(t *testing.T) {
	tasks := []task.Task{
		{Name: "test", Description: "Run tests"},
		{Name: "test-unit", Description: "Run unit tests"},
		{Name: "test-integration", Description: "Run integration tests"},
		{Name: "build", Description: "Build the application"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	// Should return the best match (exact or closest)
	result, err := manager.FindClosestTask(mockManager, "test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected task, got nil")
	}

	// Should match "test" exactly
	if result.Name != "test" {
		t.Errorf("Expected task name 'test', got '%s'", result.Name)
	}
}

func TestFindClosestTask_SimilarNames(t *testing.T) {
	tasks := []task.Task{
		{Name: "build-dev", Description: "Build for development"},
		{Name: "build-prod", Description: "Build for production"},
		{Name: "build-test", Description: "Build for testing"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dev match",
			input:    "build-dev",
			expected: "build-dev",
		},
		{
			name:     "prod match",
			input:    "build-prod",
			expected: "build-prod",
		},
		{
			name:     "partial dev",
			input:    "dev",
			expected: "build-dev",
		},
		{
			name:     "partial prod",
			input:    "prod",
			expected: "build-prod",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := manager.FindClosestTask(mockManager, tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Expected task, got nil")
			}

			if result.Name != tc.expected {
				t.Errorf("Expected task name '%s', got '%s'", tc.expected, result.Name)
			}
		})
	}
}

func TestFindClosestTask_SpecialCharacters(t *testing.T) {
	tasks := []task.Task{
		{Name: "test:unit", Description: "Run unit tests"},
		{Name: "test:integration", Description: "Run integration tests"},
		{Name: "build-app", Description: "Build application"},
		{Name: "deploy_prod", Description: "Deploy to production"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon separator",
			input:    "test:unit",
			expected: "test:unit",
		},
		{
			name:     "dash separator",
			input:    "build-app",
			expected: "build-app",
		},
		{
			name:     "underscore separator",
			input:    "deploy_prod",
			expected: "deploy_prod",
		},
		{
			name:     "partial with colon",
			input:    "unit",
			expected: "test:unit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := manager.FindClosestTask(mockManager, tc.input)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if result == nil {
				t.Fatal("Expected task, got nil")
			}

			if result.Name != tc.expected {
				t.Errorf("Expected task name '%s', got '%s'", tc.expected, result.Name)
			}
		})
	}
}

func TestFindClosestTask_EmptySearchString(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	// Empty string returns fuzzy matches but with empty targets that don't match task names
	// This reveals a bug in the original function: it returns (nil, nil) instead of (nil, error)
	result, err := manager.FindClosestTask(mockManager, "")

	if err != nil {
		t.Fatalf("Expected no error (this is the current behavior), got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for empty search string, got: %v", result)
	}

	// Note: This test documents the current behavior which is arguably a bug.
	// The function should probably return an error when no valid match is found,
	// even if fuzzy search returns matches.
}

func TestFindClosestTask_SingleTask(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
	}

	mockManager := NewMockManager("Single Task Manager", tasks)

	// Test exact match
	result, err := manager.FindClosestTask(mockManager, "build")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected task, got nil")
	}

	if result.Name != "build" {
		t.Errorf("Expected task name 'build', got '%s'", result.Name)
	}

	// Test fuzzy match
	result, err = manager.FindClosestTask(mockManager, "bui")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected task, got nil")
	}

	if result.Name != "build" {
		t.Errorf("Expected task name 'build', got '%s'", result.Name)
	}
}

// Test MockManager functionality
func TestMockManager_GetTitle(t *testing.T) {
	title := "Test Manager Title"
	mockManager := NewMockManager(title, []task.Task{})

	result := mockManager.GetTitle()
	if result.Name != title {
		t.Errorf("Expected title '%s', got '%s'", title, result.Name)
	}
}

func TestMockManager_ListTasks(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	result, err := mockManager.ListTasks()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(result))
	}

	for i, task := range tasks {
		if result[i].Name != task.Name {
			t.Errorf("Expected task name '%s', got '%s'", task.Name, result[i].Name)
		}
		if result[i].Description != task.Description {
			t.Errorf("Expected task description '%s', got '%s'", task.Description, result[i].Description)
		}
	}
}

func TestMockManager_ListTasksWithError(t *testing.T) {
	mockManager := NewMockManager("Error Manager", []task.Task{})
	expectedError := fmt.Errorf("test error")
	mockManager.SetListError(expectedError)

	result, err := mockManager.ListTasks()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

func TestMockManager_ExecuteTask(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}
	args := []string{"arg1", "arg2"}

	mockManager.ExecuteTask(testTask, args...)

	executed := mockManager.GetExecutedTasks()
	if len(executed) != 1 {
		t.Fatalf("Expected 1 executed task, got %d", len(executed))
	}

	if executed[0].Task != testTask {
		t.Error("Expected executed task to match the input task")
	}

	if len(executed[0].Args) != len(args) {
		t.Errorf("Expected %d args, got %d", len(args), len(executed[0].Args))
	}

	for i, arg := range args {
		if executed[0].Args[i] != arg {
			t.Errorf("Expected arg '%s', got '%s'", arg, executed[0].Args[i])
		}
	}
}

func TestMockManager_ExecuteTaskNoArgs(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}

	mockManager.ExecuteTask(testTask)

	executed := mockManager.GetExecutedTasks()
	if len(executed) != 1 {
		t.Fatalf("Expected 1 executed task, got %d", len(executed))
	}

	if executed[0].Task != testTask {
		t.Error("Expected executed task to match the input task")
	}

	if len(executed[0].Args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(executed[0].Args))
	}
}

func TestMockManager_ClearExecutedTasks(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}
	mockManager.ExecuteTask(testTask)

	executed := mockManager.GetExecutedTasks()
	if len(executed) != 1 {
		t.Fatalf("Expected 1 executed task before clear, got %d", len(executed))
	}

	mockManager.ClearExecutedTasks()

	executed = mockManager.GetExecutedTasks()
	if len(executed) != 0 {
		t.Errorf("Expected 0 executed tasks after clear, got %d", len(executed))
	}
}

func TestFindClosestTask_CaseSensitivity(t *testing.T) {
	tasks := []task.Task{
		{Name: "Build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "LINT", Description: "Run linters"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	// Test exact case match
	result, err := manager.FindClosestTask(mockManager, "Build")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected task, got nil")
	}

	if result.Name != "Build" {
		t.Errorf("Expected task name 'Build', got '%s'", result.Name)
	}
}

func TestFindClosestTask_VeryPoorMatch(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "lint", Description: "Run linters"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	// Test with a string that has very poor fuzzy match
	result, err := manager.FindClosestTask(mockManager, "xyz123")
	if err == nil {
		t.Fatal("Expected error for very poor match")
	}

	if result != nil {
		t.Error("Expected nil result for very poor match")
	}

	expectedError := "no task found for 'xyz123'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFindClosestTask_TaskNameSliceCreation(t *testing.T) {
	// Test that the function correctly creates the task names slice
	tasks := []task.Task{
		{Name: "alpha", Description: "First task"},
		{Name: "beta", Description: "Second task"},
		{Name: "gamma", Description: "Third task"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	// Test that we can find each task
	for _, expectedTask := range tasks {
		result, err := manager.FindClosestTask(mockManager, expectedTask.Name)
		if err != nil {
			t.Fatalf("Expected no error for task '%s', got: %v", expectedTask.Name, err)
		}

		if result == nil {
			t.Fatalf("Expected task for '%s', got nil", expectedTask.Name)
		}

		if result.Name != expectedTask.Name {
			t.Errorf("Expected task name '%s', got '%s'", expectedTask.Name, result.Name)
		}

		if result.Description != expectedTask.Description {
			t.Errorf("Expected task description '%s', got '%s'", expectedTask.Description, result.Description)
		}
	}
}
