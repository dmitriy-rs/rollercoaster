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

// taskItemWithManager extends taskItem to include manager information
type taskItemWithManager struct {
	task.Task
	ManagerIndex int
}

func (t taskItem) Title() string {
	if len(t.Aliases) > 0 {
		return t.Aliases[0]
	}
	return t.Name
}

func (t taskItem) FilterValue() string { return t.Name }

func (t taskItemWithManager) Title() string {
	if len(t.Aliases) > 0 {
		return t.Aliases[0]
	}
	return t.Name
}

func (t taskItemWithManager) FilterValue() string { return t.Name }

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
	// Try to get taskItemWithManager first, then fall back to taskItem
	var taskName, taskTitle, taskDescription string
	var managerIndex int
	var hasManagerIndex bool

	if itemWithManager, ok := listItem.(taskItemWithManager); ok {
		taskName = itemWithManager.Name
		taskTitle = itemWithManager.Title()
		taskDescription = itemWithManager.Description
		managerIndex = itemWithManager.ManagerIndex
		hasManagerIndex = true
	} else if item, ok := listItem.(taskItem); ok {
		taskName = item.Name
		taskTitle = item.Title()
		taskDescription = item.Description
		hasManagerIndex = false
	} else {
		return
	}

	// Find which manager this task belongs to
	if !hasManagerIndex {
		// Fall back to the old logic for backward compatibility
		if m.IsFiltered() || m.SettingFilter() {
			// When filtering is active or user is typing in filter, find the task in the original items list
			allItems := m.Items()
			originalIndex := -1

			// Find the original index by comparing task names
			for idx, item := range allItems {
				if taskItem, ok := item.(taskItem); ok {
					if taskItem.Name == taskName {
						originalIndex = idx
						break
					}
				} else if taskItemWithManager, ok := item.(taskItemWithManager); ok {
					if taskItemWithManager.Name == taskName {
						originalIndex = idx
						break
					}
				}
			}

			// Use the original index to determine manager
			if originalIndex >= 0 {
				for j := len(d.managerStartIndices) - 1; j >= 0; j-- {
					if originalIndex >= d.managerStartIndices[j] {
						managerIndex = j
						break
					}
				}
			}
		} else {
			// When not filtering, use the current index as before
			for j := len(d.managerStartIndices) - 1; j >= 0; j-- {
				if index >= d.managerStartIndices[j] {
					managerIndex = j
					break
				}
			}
		}
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
