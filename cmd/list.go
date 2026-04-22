package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var listCmd = &cobra.Command{
	Use:               "list",
	Short:             "List available slugs and dates",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return listExportTargets(cfg)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
