package manager_test

import (
	"fmt"
	"testing"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Should not return error for exact match")
	require.NotNil(t, result, "Should return a task for exact match")

	assert.Equal(t, "build", result.Name, "Task name should match exactly")
	assert.Equal(t, "Build the application", result.Description, "Task description should match")
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
			require.NoError(t, err, "Should not return error for fuzzy match")
			require.NotNil(t, result, "Should return a task for fuzzy match")

			assert.Equal(t, tc.expected, result.Name, "Should return the expected fuzzy match")
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
	assert.Error(t, err, "Should return error for non-existent task")
	assert.Nil(t, result, "Should return nil result for non-existent task")

	expectedError := "no task found for 'nonexistent'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
}

func TestFindClosestTask_EmptyTaskList(t *testing.T) {
	tasks := []task.Task{}

	mockManager := NewMockManager("Empty Manager", tasks)

	result, err := manager.FindClosestTask(mockManager, "anything")
	assert.Error(t, err, "Should return error for empty task list")
	assert.Nil(t, result, "Should return nil result for empty task list")

	expectedError := "no task found for 'anything'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
}

func TestFindClosestTask_ListTasksError(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
	}

	mockManager := NewMockManager("Error Manager", tasks)
	expectedError := fmt.Errorf("failed to list tasks")
	mockManager.SetListError(expectedError)

	result, err := manager.FindClosestTask(mockManager, "build")
	assert.Error(t, err, "Should return error when ListTasks fails")
	assert.Nil(t, result, "Should return nil result when ListTasks fails")
	assert.Equal(t, expectedError, err, "Should return the exact error from ListTasks")
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
	require.NoError(t, err, "Should not return error for multiple matches")
	require.NotNil(t, result, "Should return a task for multiple matches")

	// Should match "test" exactly
	assert.Equal(t, "test", result.Name, "Should return exact match when available")
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
			require.NoError(t, err, "Should not return error for similar names")
			require.NotNil(t, result, "Should return a task for similar names")

			assert.Equal(t, tc.expected, result.Name, "Should return the expected similar name match")
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
			require.NoError(t, err, "Should not return error for special characters")
			require.NotNil(t, result, "Should return a task for special characters")

			assert.Equal(t, tc.expected, result.Name, "Should handle special characters correctly")
		})
	}
}

func TestFindClosestTask_EmptySearchString(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	result, err := manager.FindClosestTask(mockManager, "")
	assert.Error(t, err, "Should return error for empty search string")
	assert.Nil(t, result, "Should return nil result for empty search string")

	expectedError := "no task found for ''"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
}

func TestFindClosestTask_SingleTask(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
	}

	mockManager := NewMockManager("Single Task Manager", tasks)

	// Test exact match
	result, err := manager.FindClosestTask(mockManager, "build")
	require.NoError(t, err, "Should not return error for exact match with single task")
	require.NotNil(t, result, "Should return task for exact match with single task")

	assert.Equal(t, "build", result.Name, "Should return the single task for exact match")

	// Test fuzzy match
	result, err = manager.FindClosestTask(mockManager, "bui")
	require.NoError(t, err, "Should not return error for fuzzy match with single task")
	require.NotNil(t, result, "Should return task for fuzzy match with single task")

	assert.Equal(t, "build", result.Name, "Should return the single task for fuzzy match")
}

// Test MockManager functionality
func TestMockManager_GetTitle(t *testing.T) {
	title := "Test Manager Title"
	mockManager := NewMockManager(title, []task.Task{})

	result := mockManager.GetTitle()
	assert.Equal(t, title, result.Name, "GetTitle should return the correct title")
}

func TestMockManager_ListTasks(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}

	mockManager := NewMockManager("Test Manager", tasks)

	result, err := mockManager.ListTasks()
	require.NoError(t, err, "ListTasks should not return error")
	assert.Len(t, result, len(tasks), "Should return correct number of tasks")

	for i, task := range tasks {
		assert.Equal(t, task.Name, result[i].Name, "Task name should match at index %d", i)
		assert.Equal(t, task.Description, result[i].Description, "Task description should match at index %d", i)
	}
}

func TestMockManager_ListTasksWithError(t *testing.T) {
	mockManager := NewMockManager("Error Manager", []task.Task{})
	expectedError := fmt.Errorf("test error")
	mockManager.SetListError(expectedError)

	result, err := mockManager.ListTasks()
	assert.Error(t, err, "Should return error when error is set")
	assert.Nil(t, result, "Should return nil result when error occurs")
	assert.Equal(t, expectedError, err, "Should return the exact error that was set")
}

func TestMockManager_ExecuteTask(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}
	args := []string{"arg1", "arg2"}

	mockManager.ExecuteTask(testTask, args...)

	executed := mockManager.GetExecutedTasks()
	require.Len(t, executed, 1, "Should have executed exactly one task")

	assert.Equal(t, testTask, executed[0].Task, "Executed task should match input task")
	assert.Len(t, executed[0].Args, len(args), "Should have correct number of arguments")

	for i, arg := range args {
		assert.Equal(t, arg, executed[0].Args[i], "Argument should match at index %d", i)
	}
}

func TestMockManager_ExecuteTaskNoArgs(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}

	mockManager.ExecuteTask(testTask)

	executed := mockManager.GetExecutedTasks()
	require.Len(t, executed, 1, "Should have executed exactly one task")

	assert.Equal(t, testTask, executed[0].Task, "Executed task should match input task")
	assert.Empty(t, executed[0].Args, "Should have no arguments when none provided")
}

func TestMockManager_ClearExecutedTasks(t *testing.T) {
	mockManager := NewMockManager("Test Manager", []task.Task{})

	testTask := &task.Task{Name: "test", Description: "Test task"}
	mockManager.ExecuteTask(testTask)

	executed := mockManager.GetExecutedTasks()
	require.Len(t, executed, 1, "Should have one executed task before clear")

	mockManager.ClearExecutedTasks()

	executed = mockManager.GetExecutedTasks()
	assert.Empty(t, executed, "Should have no executed tasks after clear")
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
	require.NoError(t, err, "Should not return error for exact case match")
	require.NotNil(t, result, "Should return task for exact case match")

	assert.Equal(t, "Build", result.Name, "Should respect case sensitivity")
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
	assert.Error(t, err, "Should return error for very poor match")
	assert.Nil(t, result, "Should return nil result for very poor match")

	expectedError := "no task found for 'xyz123'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
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
		require.NoError(t, err, "Should not return error for task '%s'", expectedTask.Name)
		require.NotNil(t, result, "Should return task for '%s'", expectedTask.Name)

		assert.Equal(t, expectedTask.Name, result.Name, "Task name should match for '%s'", expectedTask.Name)
		assert.Equal(t, expectedTask.Description, result.Description, "Task description should match for '%s'", expectedTask.Name)
	}
}

func TestFindClosestTaskFromList_SingleManager(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
		{Name: "deploy", Description: "Deploy the application"},
	}

	manager1 := NewMockManager("Manager1", tasks)
	managers := []manager.Manager{manager1}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "build")
	require.NoError(t, err, "Should not return error when task is found")
	require.NotNil(t, resultManager, "Should return the manager that contains the task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Manager1", resultManager.GetTitle().Name, "Should return the correct manager")
	assert.Equal(t, "build", resultTask.Name, "Should return the correct task")
	assert.Equal(t, "Build the application", resultTask.Description, "Task description should match")
}

func TestFindClosestTaskFromList_TwoManagers_TaskInFirst(t *testing.T) {
	tasks1 := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}
	tasks2 := []task.Task{
		{Name: "deploy", Description: "Deploy the application"},
		{Name: "lint", Description: "Run linters"},
	}

	manager1 := NewMockManager("Manager1", tasks1)
	manager2 := NewMockManager("Manager2", tasks2)
	managers := []manager.Manager{manager1, manager2}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "build")
	require.NoError(t, err, "Should not return error when task is found")
	require.NotNil(t, resultManager, "Should return the manager that contains the task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Manager1", resultManager.GetTitle().Name, "Should return the first manager with matching task")
	assert.Equal(t, "build", resultTask.Name, "Should return the correct task")
}

func TestFindClosestTaskFromList_TwoManagers_TaskInSecond(t *testing.T) {
	tasks1 := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}
	tasks2 := []task.Task{
		{Name: "deploy", Description: "Deploy the application"},
		{Name: "lint", Description: "Run linters"},
	}

	manager1 := NewMockManager("Manager1", tasks1)
	manager2 := NewMockManager("Manager2", tasks2)
	managers := []manager.Manager{manager1, manager2}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "deploy")
	require.NoError(t, err, "Should not return error when task is found")
	require.NotNil(t, resultManager, "Should return the manager that contains the task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Manager2", resultManager.GetTitle().Name, "Should return the second manager with matching task")
	assert.Equal(t, "deploy", resultTask.Name, "Should return the correct task")
}

func TestFindClosestTaskFromList_ThreeManagers_FuzzyMatch(t *testing.T) {
	tasks1 := []task.Task{
		{Name: "build-frontend", Description: "Build frontend"},
		{Name: "test-frontend", Description: "Test frontend"},
	}
	tasks2 := []task.Task{
		{Name: "build-backend", Description: "Build backend"},
		{Name: "test-backend", Description: "Test backend"},
	}
	tasks3 := []task.Task{
		{Name: "deploy-staging", Description: "Deploy to staging"},
		{Name: "deploy-production", Description: "Deploy to production"},
	}

	manager1 := NewMockManager("Frontend", tasks1)
	manager2 := NewMockManager("Backend", tasks2)
	manager3 := NewMockManager("Deployment", tasks3)
	managers := []manager.Manager{manager1, manager2, manager3}

	// Test fuzzy matching - "dep" should match "deploy-staging" in the third manager
	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "dep")
	require.NoError(t, err, "Should not return error when fuzzy match is found")
	require.NotNil(t, resultManager, "Should return the manager that contains the matching task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Deployment", resultManager.GetTitle().Name, "Should return the third manager with fuzzy matching task")
	assert.True(t, resultTask.Name == "deploy-staging" || resultTask.Name == "deploy-production",
		"Should return one of the deploy tasks")
}

func TestFindClosestTaskFromList_NoTaskFound(t *testing.T) {
	tasks1 := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run tests"},
	}
	tasks2 := []task.Task{
		{Name: "deploy", Description: "Deploy the application"},
		{Name: "lint", Description: "Run linters"},
	}
	tasks3 := []task.Task{
		{Name: "format", Description: "Format code"},
		{Name: "clean", Description: "Clean build artifacts"},
	}

	manager1 := NewMockManager("Manager1", tasks1)
	manager2 := NewMockManager("Manager2", tasks2)
	manager3 := NewMockManager("Manager3", tasks3)
	managers := []manager.Manager{manager1, manager2, manager3}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "nonexistent")
	assert.Error(t, err, "Should return error when no task is found in any manager")
	assert.Nil(t, resultManager, "Should return nil manager when no task is found")
	assert.Nil(t, resultTask, "Should return nil task when no task is found")

	expectedError := "no task found for 'nonexistent'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
}

func TestFindClosestTaskFromList_EmptyManagersList(t *testing.T) {
	managers := []manager.Manager{}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "anything")
	assert.Error(t, err, "Should return error when managers list is empty")
	assert.Nil(t, resultManager, "Should return nil manager when managers list is empty")
	assert.Nil(t, resultTask, "Should return nil task when managers list is empty")

	expectedError := "no task found for 'anything'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected format")
}

func TestFindClosestTaskFromList_ManagerWithEmptyTasks(t *testing.T) {
	tasks1 := []task.Task{} // Empty task list
	tasks2 := []task.Task{
		{Name: "deploy", Description: "Deploy the application"},
		{Name: "lint", Description: "Run linters"},
	}

	manager1 := NewMockManager("EmptyManager", tasks1)
	manager2 := NewMockManager("Manager2", tasks2)
	managers := []manager.Manager{manager1, manager2}

	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "deploy")
	require.NoError(t, err, "Should not return error when task is found in second manager")
	require.NotNil(t, resultManager, "Should return the manager that contains the task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Manager2", resultManager.GetTitle().Name, "Should return the second manager with the task")
	assert.Equal(t, "deploy", resultTask.Name, "Should return the correct task")
}

func TestFindClosestTaskFromList_ManagerWithListError(t *testing.T) {
	tasks1 := []task.Task{
		{Name: "build", Description: "Build the application"},
	}
	tasks2 := []task.Task{
		{Name: "deploy", Description: "Deploy the application"},
	}

	manager1 := NewMockManager("ErrorManager", tasks1)
	manager1.SetListError(fmt.Errorf("failed to list tasks"))
	manager2 := NewMockManager("Manager2", tasks2)
	managers := []manager.Manager{manager1, manager2}

	// Even if first manager has an error, should find task in second manager
	resultManager, resultTask, err := manager.FindClosestTaskFromList(managers, "deploy")
	require.NoError(t, err, "Should not return error when task is found in working manager")
	require.NotNil(t, resultManager, "Should return the manager that contains the task")
	require.NotNil(t, resultTask, "Should return the found task")

	assert.Equal(t, "Manager2", resultManager.GetTitle().Name, "Should return the working manager")
	assert.Equal(t, "deploy", resultTask.Name, "Should return the correct task")
}
