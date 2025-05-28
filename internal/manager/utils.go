package manager

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
)

func TaskExecute(cmd *exec.Cmd) {
	logger.Debug(fmt.Sprintf("Executing task: %s", cmd.Args))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// logger.Fatal(err)
		return
	}
}
