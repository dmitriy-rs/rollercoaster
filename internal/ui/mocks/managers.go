package mocks

import (
	"fmt"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

// MockManager simulates any manager (task, workspace, etc.)
type MockManager struct {
	title        manager.Title
	tasks        []task.Task
	listError    error
	executeError error
	executed     []ExecutedTask
}

type ExecutedTask struct {
	Task *task.Task
	Args []string
}

// NewTaskManagerMock creates a new task manager mock
func NewTaskManagerMock(name, description string, tasks []task.Task) *MockManager {
	return &MockManager{
		title: manager.Title{
			Name:        name,
			Description: description,
		},
		tasks:    tasks,
		executed: make([]ExecutedTask, 0),
	}
}

// NewWorkspaceManagerMock creates a new workspace manager mock
func NewWorkspaceManagerMock(name, description string, tasks []task.Task) *MockManager {
	return &MockManager{
		title: manager.Title{
			Name:        name,
			Description: description,
		},
		tasks:    tasks,
		executed: make([]ExecutedTask, 0),
	}
}

// MockManager implementations
func (m *MockManager) GetTitle() manager.Title {
	return m.title
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

func (m *MockManager) SetExecuteError(err error) {
	m.executeError = err
}

func (m *MockManager) GetExecutedTasks() []ExecutedTask {
	return m.executed
}

func (m *MockManager) ClearExecutedTasks() {
	m.executed = make([]ExecutedTask, 0)
}

func (m *MockManager) AddTask(t task.Task) {
	m.tasks = append(m.tasks, t)
}

// Helper functions for creating common test data

// CreateSampleTaskManagerTasks returns typical task manager tasks
func CreateSampleTaskManagerTasks() []task.Task {
	return []task.Task{
		{
			Name:        "build",
			Description: "Build the application",
			Aliases:     []string{"b"},
		},
		{
			Name:        "test",
			Description: "Run all tests",
			Aliases:     []string{"t"},
		},
		{
			Name:        "lint",
			Description: "Run code linting",
			Aliases:     []string{"l"},
		},
		{
			Name:        "clean",
			Description: "Clean build artifacts",
		},
		{
			Name:        "deploy",
			Description: "Deploy to production",
		},
	}
}

// CreateSampleWorkspaceTasks returns typical workspace manager tasks
func CreateSampleWorkspaceTasks() []task.Task {
	return []task.Task{
		{
			Name:        "install",
			Description: "Install dependencies",
			Aliases:     []string{"i"},
		},
		{
			Name:        "dev",
			Description: "Start development server",
			Aliases:     []string{"start"},
		},
		{
			Name:        "build",
			Description: "Build for production",
		},
		{
			Name:        "test",
			Description: "Run test suite",
		},
		{
			Name:        "add",
			Description: "Add a new dependency",
		},
		{
			Name:        "remove",
			Description: "Remove a dependency",
		},
	}
}

// CreateLongDescriptionTasks returns tasks with long descriptions for testing truncation
func CreateLongDescriptionTasks() []task.Task {
	return []task.Task{
		{
			Name:        "very-long-task-name-that-exceeds-normal-length",
			Description: "This is a very long description that should be truncated in the UI to ensure proper formatting and readability",
			Aliases:     []string{"vl"},
		},
		{
			Name:        "short",
			Description: "Short description",
		},
	}
}

// CreateTasksWithSpecialCharacters returns tasks with special characters for testing
func CreateTasksWithSpecialCharacters() []task.Task {
	return []task.Task{
		{
			Name:        "test:unit",
			Description: "Run unit tests",
		},
		{
			Name:        "test:integration",
			Description: "Run integration tests",
		},
		{
			Name:        "build-prod",
			Description: "Build for production",
		},
		{
			Name:        "deploy_staging",
			Description: "Deploy to staging environment",
		},
	}
}

// CreateManagersForMultiManagerTest creates managers for testing multi-manager scenarios
func CreateManagersForMultiManagerTest() ([]manager.Manager, map[string]*MockManager, map[string]*MockManager) {
	taskMocks := make(map[string]*MockManager)
	workspaceMocks := make(map[string]*MockManager)
	var managers []manager.Manager

	// Create a task manager
	taskManager := NewTaskManagerMock("task", "Taskfile runner", CreateSampleTaskManagerTasks())
	taskMocks["task"] = taskManager
	managers = append(managers, taskManager)

	// Create a workspace manager
	workspaceManager := NewWorkspaceManagerMock("npm", "Node.js package manager", CreateSampleWorkspaceTasks())
	workspaceMocks["npm"] = workspaceManager
	managers = append(managers, workspaceManager)

	return managers, taskMocks, workspaceMocks
}

// CreateErrorManager creates a manager that returns errors for testing error scenarios
func CreateErrorManager() *MockManager {
	manager := NewTaskManagerMock("error-manager", "Manager that returns errors", []task.Task{})
	manager.SetListError(fmt.Errorf("failed to list tasks"))
	return manager
}
