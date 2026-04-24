package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var configSetGlobalFlag bool

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
		if err := config.Set(key, value, configSetGlobalFlag); err != nil {
			return err
		}
		scope := "project"
		if configSetGlobalFlag {
			scope = "global"
		}
		displayValue := value
		if strings.HasSuffix(key, ".api_key") {
			displayValue = maskKey(value)
		}
		fmt.Printf("Set %s = %s (%s)\n", key, displayValue, scope)
		return nil
	},
}

func init() {
	configSetCmd.Flags().BoolVarP(&configSetGlobalFlag, "global", "g", false, "Set in global config instead of project config")
	configCmd.AddCommand(configSetCmd)
}
