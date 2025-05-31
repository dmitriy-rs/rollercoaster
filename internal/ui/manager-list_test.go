package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/dmitriy-rs/rollercoaster/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShowManagerIndicator(tt.titles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManagerModel_CoreFunctionality(t *testing.T) {
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
		hasInitialFilter:    false,
	}

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

	t.Run("Update - quit keys", func(t *testing.T) {
		quitKeys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune("q")},
			{Type: tea.KeyCtrlC},
		}

		for _, key := range quitKeys {
			updatedModel, cmd := model.Update(key)

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

	t.Run("View - basic rendering", func(t *testing.T) {
		view := model.View()
		assert.Contains(t, view, "task")
		assert.Contains(t, view, "tasks 5")
	})
}

func TestManagerModel_AdditionalFeatures(t *testing.T) {
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
		hasInitialFilter:    false,
	}

	t.Run("Init", func(t *testing.T) {
		cmd := model.Init()
		assert.Nil(t, cmd)
	})

	t.Run("Update - window size", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		assert.IsType(t, managerModel{}, updatedModel)
	})

	t.Run("Update - navigation keys", func(t *testing.T) {
		// Test left/right navigation
		leftMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")}
		updatedModel, cmd := model.Update(leftMsg)
		assert.Nil(t, cmd)
		assert.IsType(t, managerModel{}, updatedModel)

		rightMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("right")}
		updatedModel, cmd = model.Update(rightMsg)
		assert.Nil(t, cmd)
		assert.IsType(t, managerModel{}, updatedModel)
	})

	t.Run("Update - slash key for filtering", func(t *testing.T) {
		slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
		updatedModel, cmd := model.Update(slashMsg)

		// Should return a command and model
		assert.NotNil(t, cmd)
		assert.IsType(t, managerModel{}, updatedModel)
	})

	t.Run("Update - ESC key when not filtering", func(t *testing.T) {
		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(escMsg)

		// Should quit when not filtering
		assert.NotNil(t, cmd)
		modelTyped := updatedModel.(managerModel)
		assert.True(t, modelTyped.quitting)
	})

	t.Run("Update - ESC key when filtering", func(t *testing.T) {
		// Enable filtering first
		filterModel := model
		filterModel.list.SetFilteringEnabled(true)
		filterModel.list.SetFilterText("test")
		filterModel.list.SetFilterState(list.FilterApplied)

		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := filterModel.Update(escMsg)

		// Should handle ESC for filtering, not quit
		assert.IsType(t, managerModel{}, updatedModel)
		// cmd may or may not be nil depending on how the list handles it
		_ = cmd
	})

	t.Run("Update - other keys fall through to list", func(t *testing.T) {
		// Test that other keys are passed to the list
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, cmd := model.Update(upMsg)

		assert.IsType(t, managerModel{}, updatedModel)
		// cmd may or may not be nil depending on list behavior
		_ = cmd
	})

	t.Run("View - with choice selected", func(t *testing.T) {
		modelWithChoice := model
		modelWithChoice.choice = task.Task{Name: "build"}

		view := modelWithChoice.View()
		assert.Equal(t, "Selected: build", view)
	})

	t.Run("View - when quitting", func(t *testing.T) {
		modelQuitting := model
		modelQuitting.quitting = true

		view := modelQuitting.View()
		assert.Empty(t, view)
	})

	t.Run("Update - enter with taskItemWithManager", func(t *testing.T) {
		// Create model with taskItemWithManager items
		var itemsWithManager []list.Item
		for i, t := range tasks {
			itemsWithManager = append(itemsWithManager, taskItemWithManager{
				Task:         t,
				ManagerIndex: i % 2, // Alternate manager indices
			})
		}

		listWithManager := list.New(itemsWithManager, delegate, 80, 14)
		modelWithManager := managerModel{
			list:                listWithManager,
			managerTitles:       managerTitles,
			taskCounts:          taskCounts,
			managerStartIndices: managerStartIndices,
			hasInitialFilter:    false,
		}

		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := modelWithManager.Update(enterMsg)

		assert.NotNil(t, cmd)
		modelTyped := updatedModel.(managerModel)
		assert.NotEmpty(t, modelTyped.choice.Name)
	})

	t.Run("View - current manager out of bounds", func(t *testing.T) {
		// Test when current manager index is out of bounds
		badModel := model
		badModel.managerTitles = []manager.Title{} // Empty titles

		view := badModel.View()
		// Should not panic and still render
		assert.Contains(t, view, "tasks 5")
	})
}

func TestRenderManagerList_ErrorCases(t *testing.T) {
	t.Run("no managers provided", func(t *testing.T) {
		resultManager, resultTask, err := RenderManagerList([]manager.Manager{}, "")

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "no managers provided")
	})

	t.Run("manager with list error", func(t *testing.T) {
		errorManager := mocks.CreateErrorManager()

		resultManager, resultTask, err := RenderManagerList([]manager.Manager{errorManager}, "")

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "failed to list tasks")
	})

	t.Run("all managers have no tasks", func(t *testing.T) {
		emptyManager1 := mocks.NewTaskManagerMock("empty1", "Empty manager 1", []task.Task{})
		emptyManager2 := mocks.NewTaskManagerMock("empty2", "Empty manager 2", []task.Task{})

		resultManager, resultTask, err := RenderManagerList([]manager.Manager{emptyManager1, emptyManager2}, "")

		assert.Error(t, err)
		assert.Nil(t, resultManager)
		assert.Nil(t, resultTask)
		assert.Contains(t, err.Error(), "no tasks found")
	})
}

func TestFilteringBasics(t *testing.T) {
	t.Run("basic filtering functionality", func(t *testing.T) {
		tasks := []task.Task{
			{Name: "build", Description: "Build the application"},
			{Name: "test", Description: "Run all tests"},
			{Name: "test-unit", Description: "Run unit tests"},
			{Name: "deploy", Description: "Deploy to production"},
		}

		var allItems []list.Item
		for _, task := range tasks {
			allItems = append(allItems, taskItem(task))
		}

		listModel := list.New(allItems, itemDelegate{}, 80, 14)
		listModel.SetFilteringEnabled(true)

		// Test filtering is enabled
		assert.True(t, listModel.FilteringEnabled())

		// Test setting filter
		listModel.SetFilterText("test")
		filteredItems := listModel.VisibleItems()

		// Should show tasks containing "test" (test and test-unit)
		assert.Equal(t, 2, len(filteredItems))
	})
}

func TestInitialFilter(t *testing.T) {
	tasks := []task.Task{
		{Name: "build", Description: "Build the application"},
		{Name: "test", Description: "Run all tests"},
		{Name: "test-unit", Description: "Run unit tests"},
		{Name: "deploy", Description: "Deploy to production"},
	}

	t.Run("ESC always quits when initial filter provided", func(t *testing.T) {
		var allItems []list.Item
		for _, task := range tasks {
			allItems = append(allItems, taskItem(task))
		}

		listModel := list.New(allItems, itemDelegate{}, 80, 14)
		listModel.SetFilteringEnabled(true)
		listModel.SetFilterText("test")
		listModel.SetFilterState(list.FilterApplied)

		model := managerModel{
			list:                listModel,
			managerTitles:       []manager.Title{{Name: "Test", Description: "Test manager"}},
			taskCounts:          []int{len(tasks)},
			managerStartIndices: []int{0},
			hasInitialFilter:    true, // This is the key - initial filter was provided
		}

		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(escMsg)

		// Should always quit when initial filter was provided
		assert.NotNil(t, cmd)
		modelTyped := updatedModel.(managerModel)
		assert.True(t, modelTyped.quitting)
	})

	t.Run("ESC behaves normally when no initial filter", func(t *testing.T) {
		var allItems []list.Item
		for _, task := range tasks {
			allItems = append(allItems, taskItem(task))
		}

		listModel := list.New(allItems, itemDelegate{}, 80, 14)
		listModel.SetFilteringEnabled(true)
		listModel.SetFilterText("test")
		listModel.SetFilterState(list.FilterApplied)

		model := managerModel{
			list:                listModel,
			managerTitles:       []manager.Title{{Name: "Test", Description: "Test manager"}},
			taskCounts:          []int{len(tasks)},
			managerStartIndices: []int{0},
			hasInitialFilter:    false, // No initial filter was provided
		}

		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(escMsg)

		// Should handle ESC for filtering, not quit immediately
		assert.IsType(t, managerModel{}, updatedModel)
		modelTyped := updatedModel.(managerModel)
		// Should not quit immediately when filtering is active and no initial filter
		assert.False(t, modelTyped.quitting)
		// cmd may or may not be nil depending on how the list handles it
		_ = cmd
	})
}
