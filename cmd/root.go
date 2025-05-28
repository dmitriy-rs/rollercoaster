package cmd

import (
	"fmt"
	"os"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "rollercoaster [TASK_NAME|PARTIAL_TASK_NAME]",
	Short:         "rollercoaster is a cli tool for running tasks/scripts in current directory",
	Long:          "rollercoaster is a cli tool for running tasks/scripts in current directory.\nIt allows you to run it without knowing the name of the manager and script.",
	SilenceErrors: false,
	Args:          cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := execute(cmd, args); err != nil {
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Oops. An error occurred while executing rollercoaster", err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", err)
		return err
	}

	manager, err := manager.FindManager(&dir)
	if err != nil {
		logger.Error("", err)
		return err
	}

	if manager == nil {
		// Find closest .git directory and try to find task manager there
		logger.Info("No task manager found in current directory, trying to find in parent directories")
	}

	tasks, err := manager.ListTasks()
	if err != nil {
		logger.Error("Failed to list tasks", err)
		return err
	}
	fmt.Printf("Found tasks: %s\n", tasks)
	if len(tasks) != 0 {
		manager.ExecuteTask(tasks[0])
	}
	return nil
}
