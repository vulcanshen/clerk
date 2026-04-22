package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:               "config",
	Short:             "Manage clerk configuration",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configShowCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
