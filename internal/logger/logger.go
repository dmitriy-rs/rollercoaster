package logger

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var MODE = "DEV"

var (
	errStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#fe5069")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Italic(true)
	infoStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#5a7ba8")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Italic(true)
	warnStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ffef50")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Italic(true)
	debugStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#005500")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Italic(true)
)

var (
	errorMessageChip = errStyle.Render(" ERROR ")
	infoMessageChip  = infoStyle.Render(" INFO ")
	warnMessageChip  = warnStyle.Render(" WARN ")
	debugMessageChip = debugStyle.Render(" DEBG ")
)

func Error(message string, err error) {
	if message == "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", errorMessageChip, err)
	} else if err == nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", errorMessageChip, message)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s %s\n", errorMessageChip, message, err)
	}
}

func Fatal(err error) {
	if err == nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", errorMessageChip, "An unknown error occurred")
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", errorMessageChip, err)
	}
	os.Exit(1)
}

func Info(message string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", infoMessageChip, message)
}

func Warning(message string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", warnMessageChip, message)
}

func Debug(message ...any) {
	if MODE == "DEV" {
		_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", debugMessageChip, message)
	}
}
