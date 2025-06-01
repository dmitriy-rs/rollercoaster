package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
)

// CacheEntry holds both the rendered string and access time for LRU eviction
type CacheEntry struct {
	value      string
	lastAccess time.Time
}

// RenderCache caches rendered strings for better performance with LRU eviction
type RenderCache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	maxSize int
}

// Global render cache instance
var DefaultRenderCache = &RenderCache{
	entries: make(map[string]*CacheEntry, 1000), // Pre-allocate for better performance
	maxSize: 500,
}

// GetCachedString retrieves a cached render string and updates access time
func (rc *RenderCache) GetCachedString(key string) (string, bool) {
	rc.mutex.Lock() // Use Lock instead of RLock to update access time
	defer rc.mutex.Unlock()

	entry, exists := rc.entries[key]
	if !exists {
		return "", false
	}

	// Update access time for LRU
	entry.lastAccess = time.Now()
	return entry.value, true
}

// SetCachedString stores a rendered string in cache with efficient LRU eviction
func (rc *RenderCache) SetCachedString(key, value string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Check if entry already exists (update case)
	if entry, exists := rc.entries[key]; exists {
		entry.value = value
		entry.lastAccess = time.Now()
		return
	}

	// Evict entries if at capacity - use efficient batch eviction
	if len(rc.entries) >= rc.maxSize {
		rc.evictOldestEntries()
	}

	// Add new entry
	rc.entries[key] = &CacheEntry{
		value:      value,
		lastAccess: time.Now(),
	}
}

// evictOldestEntries removes 20% of oldest entries efficiently
func (rc *RenderCache) evictOldestEntries() {
	removeCount := rc.maxSize / 5
	if removeCount < 10 {
		removeCount = 10 // Minimum batch size
	}

	// Pre-allocate slice for keys to remove
	type keyTime struct {
		key  string
		time time.Time
	}

	candidates := make([]keyTime, 0, len(rc.entries))
	for key, entry := range rc.entries {
		candidates = append(candidates, keyTime{key: key, time: entry.lastAccess})
	}

	// Sort by time (oldest first) - only sort what we need
	if len(candidates) > removeCount {
		// Use partial sort - find the N oldest without full sort
		for i := 0; i < removeCount; i++ {
			minIdx := i
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].time.Before(candidates[minIdx].time) {
					minIdx = j
				}
			}
			if minIdx != i {
				candidates[i], candidates[minIdx] = candidates[minIdx], candidates[i]
			}
		}
	} else {
		removeCount = len(candidates)
	}

	// Remove oldest entries
	for i := 0; i < removeCount; i++ {
		delete(rc.entries, candidates[i].key)
	}
}

// Pre-allocated styles to avoid creating new ones in render loop
var (
	itemStyle             = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
	selectedItemStyle     = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
	boldTitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	highlightedDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	managerIndicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Cache for manager titles to avoid repeated interface calls
var managerTitleCache = struct {
	sync.RWMutex
	titles map[*manager.Manager]manager.Title
}{
	titles: make(map[*manager.Manager]manager.Title),
}

// getCachedManagerTitle retrieves manager title with caching
func getCachedManagerTitle(mgr *manager.Manager) manager.Title {
	managerTitleCache.RLock()
	if title, exists := managerTitleCache.titles[mgr]; exists {
		managerTitleCache.RUnlock()
		return title
	}
	managerTitleCache.RUnlock()

	// Cache miss - get title and cache it
	title := (*mgr).GetTitle()
	managerTitleCache.Lock()
	managerTitleCache.titles[mgr] = title
	managerTitleCache.Unlock()

	return title
}

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
		managerTitle = getCachedManagerTitle(item.Manager) // Use cached version
	} else {
		// Fallback for other item types
		return
	}

	// Use efficient cache key building with StringBuilder
	var keyBuilder strings.Builder
	keyBuilder.Grow(len(taskTitle) + len(managerTitle.Name) + 50) // Pre-allocate capacity
	keyBuilder.WriteString(taskTitle)
	keyBuilder.WriteByte('_')
	keyBuilder.WriteString(managerTitle.Name)
	keyBuilder.WriteByte('_')
	keyBuilder.WriteString(fmt.Sprintf("%d_%t_%d_%d", index, d.showManagerIndicator, m.Index(), len(taskDescription)))
	cacheKey := keyBuilder.String()

	// Check cache first
	if cachedStr, found := DefaultRenderCache.GetCachedString(cacheKey); found {
		_, _ = fmt.Fprint(w, cachedStr)
		return
	}

	// Pre-compute truncated strings efficiently
	description := taskDescription
	if len(description) > 50 {
		description = description[:47] + "..."
	}

	const titleWidth = 18
	title := taskTitle
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	// Use StringBuilder for efficient string building
	var builder strings.Builder
	builder.Grow(100) // Pre-allocate reasonable capacity

	// Build the string efficiently
	builder.WriteString(fmt.Sprintf("%2d. ", index+1))

	// Add manager indicator with fixed width for alignment - only if needed
	if d.showManagerIndicator {
		managerName := managerTitle.Name
		if len(managerName) > 8 {
			managerName = managerName[:8]
		}
		indicator := fmt.Sprintf("[%s]%-8s", managerName, "")[:8] // Efficient padding
		builder.WriteString(managerIndicatorStyle.Render(indicator))
	}

	// Add padded title and description
	builder.WriteString(fmt.Sprintf("%-*s %s", titleWidth, title, description))
	baseStr := builder.String()

	var finalStr string
	if index == m.Index() {
		// For selected items, rebuild with styling
		var selectedBuilder strings.Builder
		selectedBuilder.Grow(len(baseStr) + 50)
		selectedBuilder.WriteString(fmt.Sprintf("%2d. ", index+1))

		if d.showManagerIndicator {
			managerName := managerTitle.Name
			if len(managerName) > 8 {
				managerName = managerName[:8]
			}
			indicator := fmt.Sprintf("[%s]%-8s", managerName, "")[:8]
			selectedBuilder.WriteString(managerIndicatorStyle.Render(indicator))
		}

		boldTitle := boldTitleStyle.Render(fmt.Sprintf("%-*s", titleWidth, title))
		highlightedDesc := highlightedDescStyle.Render(description)
		selectedBuilder.WriteString(boldTitle)
		selectedBuilder.WriteByte(' ')
		selectedBuilder.WriteString(highlightedDesc)

		finalStr = selectedItemStyle.Render("> " + selectedBuilder.String())
	} else {
		finalStr = itemStyle.Render(baseStr)
	}

	// Cache the result for future renders
	DefaultRenderCache.SetCachedString(cacheKey, finalStr)

	_, _ = fmt.Fprint(w, finalStr)
}
