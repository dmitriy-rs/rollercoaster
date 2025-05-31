package task

import (
	"slices"
	"strings"
)

type Task struct {
	Name        string
	Description string
	Aliases     []string
}

type TaskSource []Task

func (ts TaskSource) String(i int) string {
	return ts[i].Name
}

func (ss TaskSource) Len() int { return len(ss) }

func SortTasks(tasks []Task) {
	slices.SortStableFunc(tasks, func(a, b Task) int {
		return strings.Compare(a.Name, b.Name)
	})
}
