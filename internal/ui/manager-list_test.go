package ui

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/dmitriy-rs/rollercoaster/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock item for testing that implements list.Item but is not a taskItem
type mockListItem struct{}

func (m mockListItem) Title() string       { return "mock" }
func (m mockListItem) FilterValue() string { return "mock" }

func TestShouldShowManagerIndicator(t *testing.T) {
	tests := []struct {
		name     string
		titles   []manager.Title
		expected bool
	}{
		{
			name:     "empty slice",
			titles:   []manager.Title{},
			expected: false,
		},
		{
			name: "single manager",
			titles: []manager.Title{
				{Name: "task", Description: "Taskfile runner"},
			},
			expected: false,
		},
		{
			name: "two managers same name",
			titles: []manager.Title{
				{Name: "task", Description: "Taskfile runner"},
				{Name: "task", Description: "Another taskfile"},
			},
			expected: false,
		},
		{
			name: "two managers different names",
			titles: []manager.Title{
				{Name: "task", Description: "Taskfile runner"},
				{Name: "npm", Description: "Node.js package manager"},
			},
			expected: true,
		},
		{
			name: "three managers mixed names",
			titles: []manager.Title{
				{Name: "task", Description: "Taskfile runner"},
				{Name: "npm", Description: "Node.js package manager"},
				{Name: "task", Description: "Another taskfile"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShowManagerIndicator(tt.titles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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

func TestManagerModel(t *testing.T) {
	tasks := mocks.CreateSampleTaskManagerTasks()
	mgr := mocks.NewTaskManagerMock("task", "Taskfile runner", tasks)

	// Create list items
	var allItems []list.Item
	for _, t := range tasks {
		allItems = append(allItems, taskItem(t))
	}

	managerTitles := []manager.Title{mgr.GetTitle()}
	taskCounts := []int{len(tasks)}
	managerStartIndices := []int{0}

	delegate := itemDelegate{
		managerTitles:        managerTitles,
		taskCounts:           taskCounts,
		managerStartIndices:  managerStartIndices,
		showManagerIndicator: false,
	}

	listModel := list.New(allItems, delegate, 80, 14)

	model := managerModel{
		list:                listModel,
		managerTitles:       managerTitles,
		taskCounts:          taskCounts,
		managerStartIndices: managerStartIndices,
	}

	t.Run("Init", func(t *testing.T) {
		cmd := model.Init()
		assert.Nil(t, cmd)
	})

	t.Run("getCurrentManagerIndex", func(t *testing.T) {
		// Test with single manager
		assert.Equal(t, 0, model.getCurrentManagerIndex())

		// Test with multiple managers
		multiModel := model
		multiModel.managerStartIndices = []int{0, 3, 6}
		multiModel.list.Select(0) // First manager
		assert.Equal(t, 0, multiModel.getCurrentManagerIndex())

		multiModel.list.Select(3) // Second manager
		assert.Equal(t, 1, multiModel.getCurrentManagerIndex())

		multiModel.list.Select(7) // Third manager
		assert.Equal(t, 2, multiModel.getCurrentManagerIndex())
	})

	t.Run("Update - window size", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		assert.IsType(t, managerModel{}, updatedModel)
	})

	t.Run("Update - quit keys", func(t *testing.T) {
		quitKeys := []string{"q", "ctrl+c"}

		for _, key := range quitKeys {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			if key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			updatedModel, cmd := model.Update(msg)

			assert.NotNil(t, cmd)
			modelTyped := updatedModel.(managerModel)
			assert.True(t, modelTyped.quitting)
		}
	})

	t.Run("Update - enter key", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		modelTyped := updatedModel.(managerModel)
		assert.NotEmpty(t, modelTyped.choice.Name)
	})

	t.Run("Update - navigation keys", func(t *testing.T) {
		navigationKeys := []struct {
			key     string
			keyType tea.KeyType
		}{
			{"left", tea.KeyLeft},
			{"right", tea.KeyRight},
		}

		for _, nav := range navigationKeys {
			msg := tea.KeyMsg{Type: nav.keyType}
			updatedModel, cmd := model.Update(msg)

			assert.Nil(t, cmd)
			assert.IsType(t, managerModel{}, updatedModel)
		}
	})

	t.Run("View - normal state", func(t *testing.T) {
		view := model.View()

		assert.Contains(t, view, "task")
		assert.Contains(t, view, "Taskfile runner")
		assert.Contains(t, view, "tasks 5")
	})

	t.Run("View - with choice", func(t *testing.T) {
		modelWithChoice := model
		modelWithChoice.choice = task.Task{Name: "build"}

		view := modelWithChoice.View()
		assert.Equal(t, "Selected: build", view)
	})

	t.Run("View - quitting", func(t *testing.T) {
		modelQuitting := model
		modelQuitting.quitting = true

		view := modelQuitting.View()
		assert.Empty(t, view)
	})
}

func TestRenderManagerList_ErrorCases(t *testing.T) {
	t.Run("no managers provided", func(t *testing.T) {
		resultManager, resultTask, err := RenderManagerList([]manager.Manager{})

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "no managers provided")
	})

	t.Run("manager with list error", func(t *testing.T) {
		errorManager := mocks.CreateErrorManager()

		resultManager, resultTask, err := RenderManagerList([]manager.Manager{errorManager})

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "failed to list tasks")
	})

	t.Run("all managers have no tasks", func(t *testing.T) {
		emptyManager1 := mocks.NewTaskManagerMock("empty1", "Empty manager 1", []task.Task{})
		emptyManager2 := mocks.NewTaskManagerMock("empty2", "Empty manager 2", []task.Task{})

		resultManager, resultTask, err := RenderManagerList([]manager.Manager{emptyManager1, emptyManager2})

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "no tasks found")
	})
}

func TestRenderManagerList_Setup(t *testing.T) {
	t.Run("single manager with tasks", func(t *testing.T) {
		// Create a manager with tasks
		tasks := mocks.CreateSampleTaskManagerTasks()
		mgr := mocks.NewTaskManagerMock("task", "Taskfile runner", tasks)

		// Since RenderManagerList starts an interactive UI, we can't test the full function
		// But we can test the setup logic by examining the manager directly
		title := mgr.GetTitle()
		assert.Equal(t, "task", title.Name)
		assert.Equal(t, "Taskfile runner", title.Description)

		listedTasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, listedTasks, 5)

		// Verify task names
		expectedNames := []string{"build", "test", "lint", "clean", "deploy"}
		actualNames := make([]string, len(listedTasks))
		for i, task := range listedTasks {
			actualNames[i] = task.Name
		}

		for _, name := range expectedNames {
			assert.Contains(t, actualNames, name)
		}
	})

	t.Run("multiple managers with different types", func(t *testing.T) {
		_, taskMocks, workspaceMocks := mocks.CreateManagersForMultiManagerTest()

		assert.Len(t, taskMocks, 1)
		assert.Len(t, workspaceMocks, 1)

		// Test task manager
		taskManager := taskMocks["task"]
		assert.Equal(t, "task", taskManager.GetTitle().Name)
		taskTasks, err := taskManager.ListTasks()
		require.NoError(t, err)
		assert.Len(t, taskTasks, 5)

		// Test workspace manager
		workspaceManager := workspaceMocks["npm"]
		assert.Equal(t, "npm", workspaceManager.GetTitle().Name)
		workspaceTasks, err := workspaceManager.ListTasks()
		require.NoError(t, err)
		assert.Len(t, workspaceTasks, 6)
	})
}

func TestComplexScenarios(t *testing.T) {
	t.Run("long content handling", func(t *testing.T) {
		longTasks := mocks.CreateLongDescriptionTasks()
		mgr := mocks.NewTaskManagerMock("test", "Test manager", longTasks)

		tasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Find the long task
		var longTask task.Task
		for _, task := range tasks {
			if task.Name == "very-long-task-name-that-exceeds-normal-length" {
				longTask = task
				break
			}
		}

		assert.NotEmpty(t, longTask.Name)
		assert.Greater(t, len(longTask.Name), 18)        // Longer than title width
		assert.Greater(t, len(longTask.Description), 50) // Longer than description limit
	})

	t.Run("special characters handling", func(t *testing.T) {
		specialTasks := mocks.CreateTasksWithSpecialCharacters()
		mgr := mocks.NewTaskManagerMock("test", "Test manager", specialTasks)

		tasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, tasks, 4)

		expectedNames := []string{"test:unit", "test:integration", "build-prod", "deploy_staging"}
		actualNames := make([]string, len(tasks))
		for i, task := range tasks {
			actualNames[i] = task.Name
		}

		for _, name := range expectedNames {
			assert.Contains(t, actualNames, name)
		}
	})
}

func TestMockManagerFunctionality(t *testing.T) {
	t.Run("task manager mock operations", func(t *testing.T) {
		tasks := mocks.CreateSampleTaskManagerTasks()
		mgr := mocks.NewTaskManagerMock("task", "Taskfile runner", tasks)

		// Test basic operations
		title := mgr.GetTitle()
		assert.Equal(t, "task", title.Name)
		assert.Equal(t, "Taskfile runner", title.Description)

		listedTasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, listedTasks, 5)

		// Test task execution tracking
		testTask := &listedTasks[0]
		mgr.ExecuteTask(testTask, "arg1", "arg2")

		executed := mgr.GetExecutedTasks()
		assert.Len(t, executed, 1)
		assert.Equal(t, testTask, executed[0].Task)
		assert.Equal(t, []string{"arg1", "arg2"}, executed[0].Args)

		// Test clearing executed tasks
		mgr.ClearExecutedTasks()
		executed = mgr.GetExecutedTasks()
		assert.Len(t, executed, 0)

		// Test adding tasks
		newTask := task.Task{Name: "new-task", Description: "New task"}
		mgr.AddTask(newTask)

		updatedTasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, updatedTasks, 6)
	})

	t.Run("workspace manager mock operations", func(t *testing.T) {
		tasks := mocks.CreateSampleWorkspaceTasks()
		mgr := mocks.NewWorkspaceManagerMock("npm", "Node.js package manager", tasks)

		// Test basic operations
		title := mgr.GetTitle()
		assert.Equal(t, "npm", title.Name)
		assert.Equal(t, "Node.js package manager", title.Description)

		listedTasks, err := mgr.ListTasks()
		require.NoError(t, err)
		assert.Len(t, listedTasks, 6)

		// Test execution tracking
		testTask := &listedTasks[0]
		mgr.ExecuteTask(testTask, "--verbose")

		executed := mgr.GetExecutedTasks()
		assert.Len(t, executed, 1)
		assert.Equal(t, testTask, executed[0].Task)
		assert.Equal(t, []string{"--verbose"}, executed[0].Args)
	})

	t.Run("error manager functionality", func(t *testing.T) {
		errorMgr := mocks.CreateErrorManager()

		title := errorMgr.GetTitle()
		assert.Equal(t, "error-manager", title.Name)
		assert.Equal(t, "Manager that returns errors", title.Description)

		// Should return error when listing tasks
		tasks, err := errorMgr.ListTasks()
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Contains(t, err.Error(), "failed to list tasks")
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("sample task creation", func(t *testing.T) {
		taskManagerTasks := mocks.CreateSampleTaskManagerTasks()
		assert.Len(t, taskManagerTasks, 5)

		workspaceTasks := mocks.CreateSampleWorkspaceTasks()
		assert.Len(t, workspaceTasks, 6)

		longTasks := mocks.CreateLongDescriptionTasks()
		assert.Len(t, longTasks, 2)

		specialTasks := mocks.CreateTasksWithSpecialCharacters()
		assert.Len(t, specialTasks, 4)
	})

	t.Run("multi-manager test setup", func(t *testing.T) {
		managers, taskMocks, workspaceMocks := mocks.CreateManagersForMultiManagerTest()

		assert.Len(t, managers, 2)
		assert.Len(t, taskMocks, 1)
		assert.Len(t, workspaceMocks, 1)

		// Verify the managers are properly configured
		for _, mgr := range managers {
			title := mgr.GetTitle()
			assert.NotEmpty(t, title.Name)
			assert.NotEmpty(t, title.Description)

			tasks, err := mgr.ListTasks()
			require.NoError(t, err)
			assert.Greater(t, len(tasks), 0)
		}
	})
}

func TestAllMockHelperFunctions(t *testing.T) {
	t.Run("CreateSampleTaskManagerTasks", func(t *testing.T) {
		tasks := mocks.CreateSampleTaskManagerTasks()
		assert.Len(t, tasks, 5)

		expectedTasks := map[string]bool{
			"build":  true,
			"test":   true,
			"lint":   true,
			"clean":  true,
			"deploy": true,
		}

		for _, task := range tasks {
			assert.True(t, expectedTasks[task.Name], "Unexpected task: %s", task.Name)
			assert.NotEmpty(t, task.Description)
		}
	})

	t.Run("CreateSampleWorkspaceTasks", func(t *testing.T) {
		tasks := mocks.CreateSampleWorkspaceTasks()
		assert.Len(t, tasks, 6)

		expectedTasks := map[string]bool{
			"install": true,
			"dev":     true,
			"build":   true,
			"test":    true,
			"add":     true,
			"remove":  true,
		}

		for _, task := range tasks {
			assert.True(t, expectedTasks[task.Name], "Unexpected task: %s", task.Name)
			assert.NotEmpty(t, task.Description)
		}
	})

	t.Run("CreateLongDescriptionTasks", func(t *testing.T) {
		tasks := mocks.CreateLongDescriptionTasks()
		assert.Len(t, tasks, 2)

		// Find long task
		var longTask, shortTask task.Task
		for _, task := range tasks {
			switch task.Name {
			case "very-long-task-name-that-exceeds-normal-length":
				longTask = task
			case "short":
				shortTask = task
			}
		}

		assert.NotEmpty(t, longTask.Name)
		assert.Greater(t, len(longTask.Description), 50)
		assert.Contains(t, longTask.Aliases, "vl")

		assert.Equal(t, "short", shortTask.Name)
		assert.Equal(t, "Short description", shortTask.Description)
	})

	t.Run("CreateTasksWithSpecialCharacters", func(t *testing.T) {
		tasks := mocks.CreateTasksWithSpecialCharacters()
		assert.Len(t, tasks, 4)

		expectedTasks := []string{"test:unit", "test:integration", "build-prod", "deploy_staging"}
		actualTasks := make([]string, len(tasks))
		for i, task := range tasks {
			actualTasks[i] = task.Name
		}

		for _, expected := range expectedTasks {
			assert.Contains(t, actualTasks, expected)
		}
	})

	t.Run("CreateManagersForMultiManagerTest", func(t *testing.T) {
		managers, taskMocks, workspaceMocks := mocks.CreateManagersForMultiManagerTest()

		assert.Len(t, managers, 2)
		assert.Len(t, taskMocks, 1)
		assert.Len(t, workspaceMocks, 1)

		// Verify task manager
		taskManager, exists := taskMocks["task"]
		assert.True(t, exists)
		assert.Equal(t, "task", taskManager.GetTitle().Name)
		assert.Equal(t, "Taskfile runner", taskManager.GetTitle().Description)

		// Verify workspace manager
		workspaceManager, exists := workspaceMocks["npm"]
		assert.True(t, exists)
		assert.Equal(t, "npm", workspaceManager.GetTitle().Name)
		assert.Equal(t, "Node.js package manager", workspaceManager.GetTitle().Description)
	})

	t.Run("CreateErrorManager", func(t *testing.T) {
		errorManager := mocks.CreateErrorManager()

		assert.Equal(t, "error-manager", errorManager.GetTitle().Name)
		assert.Equal(t, "Manager that returns errors", errorManager.GetTitle().Description)

		tasks, err := errorManager.ListTasks()
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Contains(t, err.Error(), "failed to list tasks")
	})
}

func TestMockManagerErrorHandling(t *testing.T) {
	t.Run("task manager with list error", func(t *testing.T) {
		mgr := mocks.NewTaskManagerMock("test", "test", []task.Task{})
		mgr.SetListError(fmt.Errorf("custom list error"))

		tasks, err := mgr.ListTasks()
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Contains(t, err.Error(), "custom list error")
	})

	t.Run("workspace manager with list error", func(t *testing.T) {
		mgr := mocks.NewWorkspaceManagerMock("test", "test", []task.Task{})
		mgr.SetListError(fmt.Errorf("workspace list error"))

		tasks, err := mgr.ListTasks()
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Contains(t, err.Error(), "workspace list error")
	})

	t.Run("task manager execute error", func(t *testing.T) {
		tasks := []task.Task{{Name: "test", Description: "test task"}}
		mgr := mocks.NewTaskManagerMock("test", "test", tasks)
		mgr.SetExecuteError(fmt.Errorf("execute error"))

		// Execute error doesn't affect the ExecuteTask method in mocks
		// but we can verify the error is set
		mgr.ExecuteTask(&tasks[0])
		executed := mgr.GetExecutedTasks()
		assert.Len(t, executed, 1)
	})
}

func TestIntegrationScenarios(t *testing.T) {
	t.Run("full setup simulation", func(t *testing.T) {
		// Simulate the full setup process that RenderManagerList would do
		managers, _, _ := mocks.CreateManagersForMultiManagerTest()

		var allItems []interface{}
		var managerTitles []manager.Title
		var managerTaskCounts []int
		var managerStartIndices []int
		var taskToManagerMap = make(map[string]manager.Manager)

		taskIndex := 0
		for _, mgr := range managers {
			tasks, err := mgr.ListTasks()
			require.NoError(t, err)

			if len(tasks) == 0 {
				continue
			}

			managerTitles = append(managerTitles, mgr.GetTitle())
			managerTaskCounts = append(managerTaskCounts, len(tasks))
			managerStartIndices = append(managerStartIndices, taskIndex)

			for _, taskItem := range tasks {
				allItems = append(allItems, taskItem)
				// Note: if same task names exist across managers, the last one wins
				taskToManagerMap[taskItem.Name] = mgr
				taskIndex++
			}
		}

		// Verify setup
		assert.Len(t, allItems, 11) // 5 + 6 tasks total
		assert.Len(t, managerTitles, 2)
		assert.Equal(t, []int{5, 6}, managerTaskCounts)
		assert.Equal(t, []int{0, 5}, managerStartIndices)
		// taskToManagerMap may have fewer entries if task names overlap
		assert.Greater(t, len(taskToManagerMap), 0)

		// Test manager indicator decision
		showIndicator := ShouldShowManagerIndicator(managerTitles)
		assert.True(t, showIndicator) // Different manager names
	})

	t.Run("edge case handling", func(t *testing.T) {
		// Test with mixed managers (some empty, some with tasks)
		mgr1 := mocks.NewTaskManagerMock("empty", "Empty manager", []task.Task{})
		mgr2 := mocks.NewTaskManagerMock("task", "Task manager", mocks.CreateSampleTaskManagerTasks())
		mgr3 := mocks.NewWorkspaceManagerMock("npm", "NPM manager", []task.Task{}) // Empty

		managers := []manager.Manager{mgr1, mgr2, mgr3}

		var validManagers []manager.Manager
		totalTasks := 0

		for _, mgr := range managers {
			tasks, err := mgr.ListTasks()
			require.NoError(t, err)

			if len(tasks) > 0 {
				validManagers = append(validManagers, mgr)
				totalTasks += len(tasks)
			}
		}

		// Only one manager should have tasks
		assert.Len(t, validManagers, 1)
		assert.Equal(t, 5, totalTasks)
		assert.Equal(t, "task", validManagers[0].GetTitle().Name)
	})
}
