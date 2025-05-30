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

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Align(lipgloss.Center)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("75"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(0, 0, 0, 0)
)

// taskItem wraps task.Task to implement list.Item interface
type taskItem task.Task

func (t taskItem) Title() string       { return t.Name }
func (t taskItem) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(taskItem)
	if !ok {
		return
	}

	description := i.Description
	if len(description) > 50 {
		description = description[:47] + "..."
	}

	titleWidth := 20
	title := i.Name
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	paddedTitle := fmt.Sprintf("%-*s", titleWidth, title)

	str := fmt.Sprintf("%d. %s %s", index+1, paddedTitle, description)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			boldTitle := lipgloss.NewStyle().Bold(true).Render(paddedTitle)
			highlightedDescription := lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(description)
			boldStr := fmt.Sprintf("%d. %s %s", index+1, boldTitle, highlightedDescription)
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
		m.list.SetWidth(msg.Width)
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
			currentManager := m.getCurrentManagerIndex()
			prevManager := currentManager - 1
			if prevManager < 0 {
				prevManager = len(m.managerStartIndices) - 1 // Circle to last manager
			}
			m.navigateToManager(prevManager)
			return m, nil

		case "right":
			currentManager := m.getCurrentManagerIndex()
			nextManager := currentManager + 1
			if nextManager >= len(m.managerStartIndices) {
				nextManager = 0 // Circle to first manager
			}
			m.navigateToManager(nextManager)
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

	var result strings.Builder
	taskIndex := 0

	for i, managerTitle := range m.managerTitles {
		// Render manager title
		titleText := TaskNameStyle.Render(managerTitle.Name) + " " + TextColor.Render(managerTitle.Description)
		titleWithPadding := lipgloss.NewStyle().MarginLeft(3).Render(titleText)
		result.WriteString(titleWithPadding)
		result.WriteString("\n")

		// Render tasks for this manager
		taskCount := m.taskCounts[i]
		for j := 0; j < taskCount; j++ {
			if taskIndex < len(m.list.Items()) {
				item := m.list.Items()[taskIndex]
				if taskItem, ok := item.(taskItem); ok {
					// Format task similar to itemDelegate.Render
					description := taskItem.Description
					if len(description) > 50 {
						description = description[:47] + "..."
					}

					titleWidth := 20
					title := taskItem.Name
					if len(title) > titleWidth {
						title = title[:titleWidth-3] + "..."
					}

					paddedTitle := fmt.Sprintf("%-*s", titleWidth, title)
					taskStr := fmt.Sprintf("%d. %s %s", j+1, paddedTitle, description)

					// Apply selection styling if this is the current item
					if taskIndex == m.list.Index() {
						boldTitle := lipgloss.NewStyle().Bold(true).Render(paddedTitle)
						highlightedDescription := lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(description)
						taskStr = fmt.Sprintf("%d. %s %s", j+1, boldTitle, highlightedDescription)
						taskStr = selectedItemStyle.Render("> " + taskStr)
					} else {
						taskStr = itemStyle.Render(taskStr)
					}

					result.WriteString(taskStr)
					result.WriteString("\n")
				}
			}
			taskIndex++
		}

		// Add spacing between managers
		if i < len(m.managerTitles)-1 {
			result.WriteString("\n")
		}
	}

	// Add pagination if there are multiple pages
	if m.list.Paginator.TotalPages > 1 {
		result.WriteString("\n")
		result.WriteString(m.list.Styles.PaginationStyle.Render(m.list.Paginator.View()))
	}

	// Add help text
	result.WriteString("\n")
	result.WriteString(m.list.Styles.HelpStyle.Render(m.list.Help.View(m.list)))

	return result.String()
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

	l := list.New(allItems, itemDelegate{}, defaultWidth, listHeight)
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
