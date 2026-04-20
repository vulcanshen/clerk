package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func deprecatedCmd(old, new string) *cobra.Command {
	return &cobra.Command{
		Use:    old,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "Error: '%s' has been renamed to '%s'.\nRun: clerk %s\n", old, new, new)
			os.Exit(1)
		},
	}
}

func init() {
	rootCmd.AddCommand(deprecatedCmd("install", "register"))
	rootCmd.AddCommand(deprecatedCmd("uninstall", "unregister"))
	rootCmd.AddCommand(deprecatedCmd("diagnosis", "register"))
}
