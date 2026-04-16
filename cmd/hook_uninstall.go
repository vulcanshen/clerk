package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/hook"
)

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove clerk from Claude Code SessionEnd hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return hook.Uninstall()
	},
}

func init() {
	hookCmd.AddCommand(hookUninstallCmd)
}
