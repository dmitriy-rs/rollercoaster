package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
)

// managerTaskItem wraps manager.ManagerTask to implement list.Item interface
type managerTaskItem struct {
	ManagerTask manager.ManagerTask
}

func (t managerTaskItem) Title() string {
	if len(t.ManagerTask.Aliases) > 0 {
		return t.ManagerTask.Aliases[0]
	}
	return t.ManagerTask.Name
}

func (t managerTaskItem) FilterValue() string { return t.ManagerTask.Name }

type itemDelegate struct {
	showManagerIndicator bool
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var taskTitle, taskDescription string
	var managerTitle manager.Title

	// Handle ManagerTask directly
	if item, ok := listItem.(manager.ManagerTask); ok {
		taskTitle = item.Title()
		taskDescription = item.Description
		managerTitle = (*item.Manager).GetTitle()
	} else {
		// Fallback for other item types
		return
	}

	description := taskDescription
	if len(description) > 50 { // Shorter to make room for manager indicator
		description = description[:47] + "..."
	}

	titleWidth := 18
	title := taskTitle
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	paddedTitle := fmt.Sprintf("%-*s", titleWidth, title)

	// Add manager indicator with fixed width for alignment - only if needed
	managerIndicator := ""
	if d.showManagerIndicator {
		managerName := managerTitle.Name
		if len(managerName) > 8 {
			managerName = managerName[:8]
		}
		// Create indicator like "[task]" then pad the whole thing to fixed width
		indicator := fmt.Sprintf("[%s]", managerName)
		managerIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%-8s", indicator))
	}

	str := fmt.Sprintf("%2d. %s%s %s", index+1, managerIndicator, paddedTitle, description)

	fn := itemStyle.Render
	if index == m.Index() {
		boldTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(paddedTitle)
		highlightedDescription := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(description)
		boldStr := fmt.Sprintf("%2d. %s%s %s", index+1, managerIndicator, boldTitle, highlightedDescription)
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + boldStr)
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}
