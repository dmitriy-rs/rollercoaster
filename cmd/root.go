package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:           "rollercoaster [TASK_NAME|PARTIAL_TASK_NAME]",
	Short:         "rollercoaster is a cli tool for running tasks/scripts in current directory",
	Long:          "rollercoaster is a cli tool for running tasks/scripts in current directory. It allows you to run it without knowing the name of the manager and script.",
	SilenceErrors: false,
	Args:          cobra.MaximumNArgs(1),
	Run: execute,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing rollercoaster '%s'\n", err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) {
}

