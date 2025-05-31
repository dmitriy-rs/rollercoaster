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
	list             list.Model
	choice           task.Task
	chosenManager    *manager.Manager
	quitting         bool
	managerTasks     []manager.ManagerTask
	hasInitialFilter bool // Track if initial filter was provided
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
			if item, ok := selectedItem.(managerTaskItem); ok {
				m.choice = item.ManagerTask.Task
				m.chosenManager = item.ManagerTask.Manager
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
			// If initial filter was provided, always quit on ESC
			if m.hasInitialFilter && m.list.FilterState() == list.FilterApplied {
				m.quitting = true
				return m, tea.Quit
			}

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
	var header strings.Builder

	// Get the current task to show its manager
	if m.list.Index() < len(m.managerTasks) {
		currentTask := m.managerTasks[m.list.Index()]
		currentManager := *currentTask.Manager
		currentTitle := currentManager.GetTitle()

		titleText := TaskNameStyle.Render(currentTitle.Name) + " " + TextColor.Render(currentTitle.Description)
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

func RenderTasksList(managerTasks []manager.ManagerTask, initialFilter string) (*manager.Manager, *task.Task, error) {
	if len(managerTasks) == 0 {
		return nil, nil, fmt.Errorf("no tasks provided")
	}

	// Convert manager tasks to list items
	var allItems []list.Item
	for _, mgr := range managerTasks {
		allItems = append(allItems, managerTaskItem{ManagerTask: mgr})
	}

	// Collect unique manager titles for determining if indicators should be shown
	var managerTitles []manager.Title
	seenManagers := make(map[string]bool)

	for _, mgr := range managerTasks {
		title := (*mgr.Manager).GetTitle()
		managerKey := title.Name + title.Description
		if !seenManagers[managerKey] {
			managerTitles = append(managerTitles, title)
			seenManagers[managerKey] = true
		}
	}

	// Determine if manager indicators should be shown
	showManagerIndicator := ShouldShowManagerIndicator(managerTitles)

	const defaultWidth = 80
	const defaultHeight = 14

	l := list.New(allItems, itemDelegate{showManagerIndicator: showManagerIndicator}, defaultWidth, defaultHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	hasInitialFilter := initialFilter != ""

	// If initial filter is provided, set it up before creating the model
	if hasInitialFilter {
		l.SetFilterText(initialFilter)
		l.SetFilterState(list.FilterApplied)
	}

	m := managerModel{
		list:             l,
		managerTasks:     managerTasks,
		hasInitialFilter: hasInitialFilter,
	}

	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, nil, err
	}

	// Extract the selected task from the final model
	if model, ok := finalModel.(managerModel); ok {
		if model.choice.Name != "" && model.chosenManager != nil {
			return model.chosenManager, &model.choice, nil
		}
	}

	// User quit without selecting
	return nil, nil, nil
}
