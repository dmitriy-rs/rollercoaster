package cmd

import (
	"fmt"
	"os"

	"github.com/dmitriy-rs/rollercoaster/internal/config"
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/parser"
	ui "github.com/dmitriy-rs/rollercoaster/internal/ui/tasks-list"
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
		cfg := config.LoadConfig()
		if err := execute(cmd, args, cfg); err != nil {
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

func execute(cmd *cobra.Command, args []string, cfg *config.Config) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Handle case where config failed to load
	var defaultJSManager string
	if cfg != nil {
		defaultJSManager = cfg.DefaultJSManager
	}

	managers, err := parser.ParseManager(&dir, &parser.ParseManagerConfig{
		DefaultJSManager: defaultJSManager,
	})
	if err != nil {
		return err
	}
	if len(managers) == 0 {
		return nil
	}

	if len(args) == 0 {
		return executeWithoutArgs(managers)
	} else {
		return executeWithArgs(managers, args, cfg)
	}
}

func executeWithoutArgs(managers []manager.Manager) error {
	tasks, err := manager.GetManagerTasksFromList(managers)
	if err != nil {
		return err
	}

	return handleTasksListUI(tasks, "")
}

func executeWithArgs(managers []manager.Manager, args []string, cfg *config.Config) error {
	commandName := args[0]
	commandArgs := args[1:]

	tasks, err := findTasksWithFallback(managers, commandName, cfg.AutoSelectClosest)
	if err != nil {
		return handleNoTasksFound(managers)
	}

	return handleTaskSelection(tasks, commandName, commandArgs)
}

func handleTaskSelection(tasks []manager.ManagerTask, commandName string, commandArgs []string) error {
	if len(tasks) == 1 {
		return executeSingleTask(&tasks[0], commandArgs...)
	} else {
		return handleTasksListUI(tasks, commandName)
	}
}

func handleTasksListUI(tasks []manager.ManagerTask, initialSelection string) error {
	selectedManager, selectedTask, err := ui.RenderTasksList(tasks, initialSelection)
	if err != nil {
		return err
	}

	if selectedManager != nil && selectedTask != nil {
		return executeSingleTask(&manager.ManagerTask{
			Manager: selectedManager,
			Task:    *selectedTask,
		})
	}
	return nil
}

func findTasksWithFallback(managers []manager.Manager, commandName string, autoSelectClosest bool) ([]manager.ManagerTask, error) {
	if autoSelectClosest {
		closestTask, err := manager.FindClosestTaskFromList(managers, commandName)
		if err != nil {
			return nil, err
		}
		return []manager.ManagerTask{*closestTask}, nil
	} else {
		tasks, err := manager.FindAllClosestTasksFromList(managers, commandName)
		if err != nil {
			return nil, err
		}
		return tasks, nil
	}
}

func handleNoTasksFound(managers []manager.Manager) error {
	logger.Info("No tasks found")
	return executeWithoutArgs(managers)
}

func executeSingleTask(managerTask *manager.ManagerTask, args ...string) error {
	(*managerTask.Manager).ExecuteTask(&managerTask.Task, args...)
	return nil
}
