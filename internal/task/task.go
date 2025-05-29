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

func SortTasks(tasks []Task) {
	slices.SortStableFunc(tasks, func(a, b Task) int {
		return strings.Compare(a.Name, b.Name)
	})
}
