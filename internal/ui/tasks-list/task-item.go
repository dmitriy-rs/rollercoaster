package ui

import (
	"fmt"
	"io"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
)

// RenderCache caches rendered strings for better performance
type RenderCache struct {
	taskStrings map[string]string
	mutex       sync.RWMutex
}

// Global render cache instance
var DefaultRenderCache = &RenderCache{
	taskStrings: make(map[string]string),
}

// GetCachedString retrieves a cached render string
func (rc *RenderCache) GetCachedString(key string) (string, bool) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	str, exists := rc.taskStrings[key]
	return str, exists
}

// SetCachedString stores a rendered string in cache
func (rc *RenderCache) SetCachedString(key, value string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Simple cache size management
	if len(rc.taskStrings) > 500 {
		// Clear oldest half of cache entries
		count := 0
		for k := range rc.taskStrings {
			if count > 250 {
				break
			}
			delete(rc.taskStrings, k)
			count++
		}
	}

	rc.taskStrings[key] = value
}

// Pre-allocated styles to avoid creating new ones in render loop
var (
	itemStyle             = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle     = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
	boldTitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	highlightedDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	managerIndicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
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

	// Create cache key for this specific render configuration - include description length for correct caching
	cacheKey := fmt.Sprintf("%s_%s_%d_%t_%d_%d", taskTitle, managerTitle.Name, index, d.showManagerIndicator, m.Index(), len(taskDescription))

	// Check cache first
	if cachedStr, found := DefaultRenderCache.GetCachedString(cacheKey); found {
		_, _ = fmt.Fprint(w, cachedStr)
		return
	}

	// Pre-compute truncated strings
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
		managerIndicator = managerIndicatorStyle.Render(fmt.Sprintf("%-8s", indicator))
	}

	str := fmt.Sprintf("%2d. %s%s %s", index+1, managerIndicator, paddedTitle, description)

	var finalStr string
	if index == m.Index() {
		boldTitle := boldTitleStyle.Render(paddedTitle)
		highlightedDescription := highlightedDescStyle.Render(description)
		boldStr := fmt.Sprintf("%2d. %s%s %s", index+1, managerIndicator, boldTitle, highlightedDescription)
		finalStr = selectedItemStyle.Render("> " + boldStr)
	} else {
		finalStr = itemStyle.Render(str)
	}

	// Cache the result for future renders
	DefaultRenderCache.SetCachedString(cacheKey, finalStr)

	_, _ = fmt.Fprint(w, finalStr)
}
