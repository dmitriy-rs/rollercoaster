package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
)

// taskItem wraps task.Task to implement list.Item interface
type taskItem task.Task

func (t taskItem) Title() string {
	if len(t.Aliases) > 0 {
		return t.Aliases[0]
	}
	return t.Name
}

func (t taskItem) FilterValue() string { return t.Name }

type itemDelegate struct {
	managerTitles        []manager.Title
	taskCounts           []int
	managerStartIndices  []int
	showManagerIndicator bool
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
	if len(description) > 50 { // Shorter to make room for manager indicator
		description = description[:47] + "..."
	}

	titleWidth := 18
	title := i.Title()
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	paddedTitle := fmt.Sprintf("%-*s", titleWidth, title)

	// Add manager indicator with fixed width for alignment - only if needed
	managerIndicator := ""
	if d.showManagerIndicator && managerIndex < len(d.managerTitles) {
		managerName := d.managerTitles[managerIndex].Name
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
