package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Align(lipgloss.Center)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(0, 0, 0, 0)
)

// taskItem wraps task.Task to implement list.Item interface
type taskItem task.Task

func (t taskItem) Title() string       { return t.Name }
func (t taskItem) FilterValue() string { return "" }

type itemDelegate struct {
	managerTitles       []manager.Title
	taskCounts          []int
	managerStartIndices []int
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(taskItem)
	if !ok {
		return
	}

	// Find which manager this task belongs to
	managerIndex := 0
	for j := len(d.managerStartIndices) - 1; j >= 0; j-- {
		if index >= d.managerStartIndices[j] {
			managerIndex = j
			break
		}
	}

	description := i.Description
	if len(description) > 40 { // Shorter to make room for manager indicator
		description = description[:37] + "..."
	}

	titleWidth := 18
	title := i.Name
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	paddedTitle := fmt.Sprintf("%-*s", titleWidth, title)

	// Add manager indicator with fixed width for alignment
	managerIndicator := ""
	if managerIndex < len(d.managerTitles) {
		managerName := d.managerTitles[managerIndex].Name
		if len(managerName) > 8 {
			managerName = managerName[:8]
		}
		// Create indicator like "[task]" then pad the whole thing to fixed width
		indicator := fmt.Sprintf("[%s]", managerName)
		managerIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%-8s", indicator))
	}

	str := fmt.Sprintf("%d. %s%s %s", index+1, managerIndicator, paddedTitle, description)

	fn := itemStyle.Render
	if index == m.Index() {
		boldTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(paddedTitle)
		highlightedDescription := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(description)
		boldStr := fmt.Sprintf("%d. %s%s %s", index+1, managerIndicator, boldTitle, highlightedDescription)
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + boldStr)
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

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

// navigateToManager moves the cursor to the first task of the specified manager
func (m *managerModel) navigateToManager(managerIndex int) {
	if managerIndex >= 0 && managerIndex < len(m.managerStartIndices) {
		m.list.Select(m.managerStartIndices[managerIndex])
	}
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
			i, ok := m.list.SelectedItem().(taskItem)
			if ok {
				m.choice = task.Task(i)
			}
			return m, tea.Quit

		case "left":
			m.list.PrevPage()
			return m, nil

		case "right":
			m.list.NextPage()
			return m, nil
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

		// Make "Manager:" greyish and the title bold
		managerLabel := lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("240")).Render("Manager: ")
		managerTitle := lipgloss.NewStyle().Bold(true).Render(titleText)
		header.WriteString(managerLabel + managerTitle)
	}

	// Use the built-in list view for proper scrolling and pagination
	listView := m.list.View()

	return header.String() + listView
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
	for _, mgr := range managers {
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
			allItems = append(allItems, taskItem(t))
			taskToManagerMap[t.Name] = mgr // Store the mapping
			taskIndex++
		}
	}

	if len(allItems) == 0 {
		return nil, nil, fmt.Errorf("no tasks found")
	}

	const defaultWidth = 80
	const defaultHeight = 14

	l := list.New(allItems, itemDelegate{managerTitles, managerTaskCounts, managerStartIndices}, defaultWidth, defaultHeight)
	l.Title = ""
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
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
