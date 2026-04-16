package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/hook"
)

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install clerk as a Claude Code SessionEnd hook",
	RunE: func(cmd *cobra.Command, args []string) error {
		return hook.Install()
	},
}

func init() {
	hookCmd.AddCommand(hookInstallCmd)
}
