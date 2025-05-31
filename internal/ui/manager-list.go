package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

var (
	titleStyle      = lipgloss.NewStyle().MarginLeft(2).Align(lipgloss.Center)
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle   = lipgloss.NewStyle().Margin(0, 0, 0, 0)
)

type managerModel struct {
	list                list.Model
	choice              task.Task
	quitting            bool
	managerTitles       []manager.Title
	taskCounts          []int
	managerStartIndices []int // Track where each manager's tasks start
}

// getCurrentManagerIndex returns the index of the manager that contains the currently selected task
func (m managerModel) getCurrentManagerIndex() int {
	currentIndex := m.list.Index()
	for i := len(m.managerStartIndices) - 1; i >= 0; i-- {
		if currentIndex >= m.managerStartIndices[i] {
			return i
		}
	}
	return 0
}

func (m managerModel) Init() tea.Cmd {
	return nil
}

func (m managerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Use the list's built-in SetSize method which handles pagination properly
		m.list.SetSize(msg.Width, msg.Height-4) // Leave space for help text
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			selectedItem := m.list.SelectedItem()
			if itemWithManager, ok := selectedItem.(taskItemWithManager); ok {
				m.choice = itemWithManager.Task
			} else if item, ok := selectedItem.(taskItem); ok {
				m.choice = task.Task(item)
			}
			return m, tea.Quit

		case "left":
			m.list.PrevPage()
			// Adjust selection if current index is beyond available items on this page
			visibleItems := m.list.VisibleItems()
			if len(visibleItems) > 0 && m.list.Index() >= len(visibleItems) {
				m.list.Select(len(visibleItems) - 1)
			}
			return m, nil

		case "right":
			m.list.NextPage()
			// Adjust selection if current index is beyond available items on this page
			visibleItems := m.list.VisibleItems()
			if len(visibleItems) > 0 && m.list.Index() >= len(visibleItems) {
				m.list.Select(len(visibleItems) - 1)
			}
			return m, nil

		case "/":
			// Start filtering mode
			m.list.SetFilteringEnabled(true)
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case "esc":
			// Let the list handle ESC first (to exit filtering if active)
			wasFiltering := m.list.FilterState() == list.Filtering || m.list.FilterState() == list.FilterApplied
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)

			// If we were filtering and the list handled it, don't quit
			if wasFiltering {
				return m, cmd
			}

			// Otherwise, quit the application
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m managerModel) View() string {
	if m.choice.Name != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected: %s", m.choice.Name))
	}
	if m.quitting {
		return quitTextStyle.Render("")
	}

	// Show current manager context at the top
	currentManagerIndex := m.getCurrentManagerIndex()
	var header strings.Builder

	if currentManagerIndex < len(m.managerTitles) {
		currentManager := m.managerTitles[currentManagerIndex]
		titleText := TaskNameStyle.Render(currentManager.Name) + " " + TextColor.Render(currentManager.Description)

		managerTitle := lipgloss.NewStyle().PaddingLeft(5).Bold(true).Render(titleText)
		header.WriteString(managerTitle)
	}

	// Use the built-in list view for proper scrolling and pagination
	listView := m.list.View()

	totalItems := len(m.list.Items())

	statusInfo := fmt.Sprintf("tasks %d", totalItems)

	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(4).
		Render(statusInfo)

	return header.String() + listView + "\n" + statusBar
}

func RenderManagerList(managers []manager.Manager) (*manager.Manager, *task.Task, error) {
	if len(managers) == 0 {
		return nil, nil, fmt.Errorf("no managers provided")
	}

	// Collect all tasks from all managers with manager info
	var allItems []list.Item
	var managerTitles []manager.Title
	var managerTaskCounts []int
	var managerStartIndices []int
	var taskToManagerMap = make(map[string]manager.Manager) // Map task name to its manager

	taskIndex := 0
	for managerIdx, mgr := range managers {
		tasks, err := mgr.ListTasks()
		if err != nil {
			return nil, nil, err
		}

		if len(tasks) == 0 {
			continue // Skip managers with no tasks
		}

		managerTitles = append(managerTitles, mgr.GetTitle())
		managerTaskCounts = append(managerTaskCounts, len(tasks))
		managerStartIndices = append(managerStartIndices, taskIndex)

		for _, t := range tasks {
			// Use taskItemWithManager to store the manager index directly
			itemWithManager := taskItemWithManager{
				Task:         t,
				ManagerIndex: managerIdx,
			}
			allItems = append(allItems, itemWithManager)
			taskToManagerMap[t.Name] = mgr // Store the mapping
			taskIndex++
		}
	}

	if len(allItems) == 0 {
		return nil, nil, fmt.Errorf("no tasks found")
	}

	// Determine if manager indicators should be shown
	showManagerIndicator := ShouldShowManagerIndicator(managerTitles)

	const defaultWidth = 80
	const defaultHeight = 14

	l := list.New(allItems, itemDelegate{managerTitles, managerTaskCounts, managerStartIndices, showManagerIndicator}, defaultWidth, defaultHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := managerModel{
		list:                l,
		managerTitles:       managerTitles,
		taskCounts:          managerTaskCounts,
		managerStartIndices: managerStartIndices,
	}

	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, nil, err
	}

	// Extract the selected task from the final model
	if model, ok := finalModel.(managerModel); ok {
		if model.choice.Name != "" {
			// Find the manager for this task
			if mgr, exists := taskToManagerMap[model.choice.Name]; exists {
				return &mgr, &model.choice, nil
			}
		}
	}

	// User quit without selecting
	return nil, nil, nil
}
