package manager

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
)

var commandTextStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#6b9bd1")).
	Bold(true)

func CommandExecute(cmd *exec.Cmd, args ...string) {
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args...)
	}

	logger.Info(fmt.Sprintf("Executing: %s", commandTextStyle.Render(strings.Join(cmd.Args, " "))))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// logger.Fatal(err)
		return
	}
}
