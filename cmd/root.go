package cmd

import (
	"fmt"
	"os"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/parser"
	"github.com/dmitriy-rs/rollercoaster/internal/ui"
	"github.com/spf13/cobra"
)

var VERSION string = "dev"

var rootCmd = &cobra.Command{
	Use:           "rollercoaster [TASK_NAME|TASK_NAME_QUERY]",
	Short:         "rollercoaster is a cli tool for running tasks/scripts in current directory",
	Long:          "rollercoaster is a cli tool for running tasks/scripts in current directory.\nIt allows you to run it without knowing the name of the manager and script.",
	SilenceErrors: false,
	Version:       VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		if err := execute(cmd, args); err != nil {
			logger.Error("", err)
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
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	managers, err := parser.ParseManager(&dir)
	if err != nil {
		return err
	}
	if len(managers) == 0 {
		return nil
	}

	if len(args) == 0 {
		selectedManager, selectedTask, err := ui.RenderManagerList(managers, "")
		if err != nil {
			return err
		}

		// If user selected a task, execute it
		if selectedManager != nil && selectedTask != nil {
			(*selectedManager).ExecuteTask(selectedTask)
		}
		// If user quit without selecting, just return (no error)
		return nil
	} else {
		commandName := args[0]
		commandArgs := args[1:]
		taskManager, closestTask, err := manager.FindClosestTaskFromList(managers, commandName)
		if err != nil {
			logger.Info("No tasks found")
			selectedManager, selectedTask, err := ui.RenderManagerList(managers, "")
			if err != nil {
				return err
			}

			// If user selected a task, execute it
			if selectedManager != nil && selectedTask != nil {
				(*selectedManager).ExecuteTask(selectedTask)
			}
			return nil
		}
		taskManager.ExecuteTask(closestTask, commandArgs...)
		return nil
	}
}
