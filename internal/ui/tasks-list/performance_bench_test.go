package ui

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

// Benchmark render performance
func BenchmarkItemDelegate_Render(b *testing.B) {
	// Setup test data
	var mgr manager.Manager = &benchMockManager{title: manager.Title{Name: "task", Description: "Taskfile runner"}}
	managerTask := manager.ManagerTask{
		Task:    task.Task{Name: "build", Description: "Build application with lots of dependencies and complex configuration"},
		Manager: &mgr,
	}

	tasks := []list.Item{managerTask}
	listModel := list.New(tasks, itemDelegate{showManagerIndicator: true}, 80, 10)
	delegate := itemDelegate{showManagerIndicator: true}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		delegate.Render(&buf, listModel, 0, tasks[0])
	}
}

// Benchmark cache performance
func BenchmarkRenderCache_Operations(b *testing.B) {
	cache := &RenderCache{
		entries: make(map[string]*CacheEntry, 1000),
		maxSize: 500,
	}

	b.Run("SetCachedString", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := "test_key_" + string(rune(i%1000))
			value := "cached_render_string_with_styling_and_formatting"
			cache.SetCachedString(key, value)
		}
	})

	b.Run("GetCachedString", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 100; i++ {
			key := "test_key_" + string(rune(i))
			value := "cached_render_string_with_styling_and_formatting"
			cache.SetCachedString(key, value)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := "test_key_" + string(rune(i%100))
			cache.GetCachedString(key)
		}
	})
}

// Benchmark manager task creation
func BenchmarkGetManagerTasksFromList(b *testing.B) {
	// Create test managers
	tasks1 := make([]task.Task, 50)
	for i := range tasks1 {
		tasks1[i] = task.Task{
			Name:        "task_" + string(rune(i)),
			Description: "Description for task " + string(rune(i)),
		}
	}

	tasks2 := make([]task.Task, 30)
	for i := range tasks2 {
		tasks2[i] = task.Task{
			Name:        "workspace_task_" + string(rune(i)),
			Description: "Workspace description for task " + string(rune(i)),
		}
	}

	manager1 := &benchMockManager{title: manager.Title{Name: "task", Description: "Task manager"}, tasks: tasks1}
	manager2 := &benchMockManager{title: manager.Title{Name: "npm", Description: "NPM manager"}, tasks: tasks2}
	managers := []manager.Manager{manager1, manager2}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = manager.GetManagerTasksFromList(managers)
	}
}

// Mock manager for benchmarks
type benchMockManager struct {
	title manager.Title
	tasks []task.Task
}

func (m *benchMockManager) GetTitle() manager.Title                     { return m.title }
func (m *benchMockManager) ListTasks() ([]task.Task, error)             { return m.tasks, nil }
func (m *benchMockManager) ExecuteTask(task *task.Task, args ...string) {}
