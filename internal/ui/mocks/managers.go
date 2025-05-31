package mocks

import (
	"fmt"

	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

// TaskManagerMock simulates a task manager (like Taskfile/Makefile)
type TaskManagerMock struct {
	title        manager.Title
	tasks        []task.Task
	listError    error
	executeError error
	executed     []ExecutedTask
}

// WorkspaceManagerMock simulates a workspace manager (like npm/yarn/pnpm)
type WorkspaceManagerMock struct {
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
func NewTaskManagerMock(name, description string, tasks []task.Task) *TaskManagerMock {
	return &TaskManagerMock{
		title: manager.Title{
			Name:        name,
			Description: description,
		},
		tasks:    tasks,
		executed: make([]ExecutedTask, 0),
	}
}

// NewWorkspaceManagerMock creates a new workspace manager mock
func NewWorkspaceManagerMock(name, description string, tasks []task.Task) *WorkspaceManagerMock {
	return &WorkspaceManagerMock{
		title: manager.Title{
			Name:        name,
			Description: description,
		},
		tasks:    tasks,
		executed: make([]ExecutedTask, 0),
	}
}

// TaskManagerMock implementations
func (m *TaskManagerMock) GetTitle() manager.Title {
	return m.title
}

func (m *TaskManagerMock) ListTasks() ([]task.Task, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tasks, nil
}

func (m *TaskManagerMock) ExecuteTask(task *task.Task, args ...string) {
	m.executed = append(m.executed, ExecutedTask{
		Task: task,
		Args: args,
	})
}

func (m *TaskManagerMock) SetListError(err error) {
	m.listError = err
}

func (m *TaskManagerMock) SetExecuteError(err error) {
	m.executeError = err
}

func (m *TaskManagerMock) GetExecutedTasks() []ExecutedTask {
	return m.executed
}

func (m *TaskManagerMock) ClearExecutedTasks() {
	m.executed = make([]ExecutedTask, 0)
}

func (m *TaskManagerMock) AddTask(t task.Task) {
	m.tasks = append(m.tasks, t)
}

// WorkspaceManagerMock implementations
func (m *WorkspaceManagerMock) GetTitle() manager.Title {
	return m.title
}

func (m *WorkspaceManagerMock) ListTasks() ([]task.Task, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.tasks, nil
}

func (m *WorkspaceManagerMock) ExecuteTask(task *task.Task, args ...string) {
	m.executed = append(m.executed, ExecutedTask{
		Task: task,
		Args: args,
	})
}

func (m *WorkspaceManagerMock) SetListError(err error) {
	m.listError = err
}

func (m *WorkspaceManagerMock) SetExecuteError(err error) {
	m.executeError = err
}

func (m *WorkspaceManagerMock) GetExecutedTasks() []ExecutedTask {
	return m.executed
}

func (m *WorkspaceManagerMock) ClearExecutedTasks() {
	m.executed = make([]ExecutedTask, 0)
}

func (m *WorkspaceManagerMock) AddTask(t task.Task) {
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
func CreateManagersForMultiManagerTest() ([]manager.Manager, map[string]*TaskManagerMock, map[string]*WorkspaceManagerMock) {
	taskMocks := make(map[string]*TaskManagerMock)
	workspaceMocks := make(map[string]*WorkspaceManagerMock)
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
func CreateErrorManager() *TaskManagerMock {
	manager := NewTaskManagerMock("error-manager", "Manager that returns errors", []task.Task{})
	manager.SetListError(fmt.Errorf("failed to list tasks"))
	return manager
}
