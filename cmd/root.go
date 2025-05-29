package cmd

import (
	"os"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/parser"
	"github.com/dmitriy-rs/rollercoaster/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "rollercoaster [TASK_NAME|PARTIAL_TASK_NAME]",
	Short:         "rollercoaster is a cli tool for running tasks/scripts in current directory",
	Long:          "rollercoaster is a cli tool for running tasks/scripts in current directory.\nIt allows you to run it without knowing the name of the manager and script.",
	SilenceErrors: false,
	Run: func(cmd *cobra.Command, args []string) {
		if err := execute(cmd, args); err != nil {
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// logger.Error("Oops. An error occurred while executing rollercoaster", err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", err)
		return err
	}

	taskManager, err := parser.ParseManager(&dir)
	if err != nil {
		logger.Error("", err)
		return err
	}
	if taskManager == nil {
		return nil
	}

	if len(args) == 0 {
		return ui.RenderTaskList(taskManager)
	} else {
		commandName := args[0]
		commandArgs := args[1:]
		closestTask, err := manager.FindClosestTask(taskManager, commandName)
		if err != nil {
			logger.Warning("No tasks found")
			return ui.RenderTaskList(taskManager)
		}
		taskManager.ExecuteTask(closestTask, commandArgs...)
		return nil
	}
}
