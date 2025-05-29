package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

var (
	tasksTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			Underline(true)
	taskNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00aaff")).
			Bold(true)
	taskDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bebebe")).
				Bold(false).
				Align(lipgloss.Right)
)

func RenderTaskList(manager manager.Manager) error {
	tasks, err := manager.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		println("No tasks found.")
		return nil
	}

	maxSpaces := getMaxSpaces(tasks)

	println(tasksTitleStyle.Underline(false).Render("Choose a task to run:"))

	title := manager.GetTitle()
	println("\n" + TaskNameStyle.Render(title.Name) + " " + TextColor.Render(title.Description))

	println("\n" + tasksTitleStyle.Render("Name") + renderSpaces(4, maxSpaces) + tasksTitleStyle.Render("Description"))

	for _, t := range tasks {
		name := taskNameStyle.Render(t.Name)
		description := taskDescriptionStyle.Render(t.Description)
		println(name + renderSpaces(len(t.Name), maxSpaces) + description)
	}

	return nil
}

func getMaxSpaces(tasks []task.Task) int {
	spaces := 0
	for _, t := range tasks {
		if len(t.Name) > spaces {
			spaces = len(t.Name)
		}
	}
	return spaces + 12
}

func renderSpaces(nameLength int, maxSpaces int) string {
	if nameLength >= maxSpaces {
		return ""
	}
	spaces := strings.Repeat(" ", maxSpaces-nameLength)
	return spaces
}
