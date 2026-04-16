package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return config.ValidKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := config.Set(key, value); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
