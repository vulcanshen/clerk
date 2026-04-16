package cmd

import "github.com/spf13/cobra"

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage Claude Code SessionEnd hook",
}

func init() {
	rootCmd.AddCommand(hookCmd)
}
