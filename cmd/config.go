package cmd

import (
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	ui "github.com/dmitriy-rs/rollercoaster/internal/ui/config-list"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Open configuration interface",
	Long:    "Open a TUI interface to manage rollercoaster configuration settings including default JS manager, auto-select behavior, and other preferences.",
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.RenderConfigList(); err != nil {
			logger.Error("Failed to open config interface", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
