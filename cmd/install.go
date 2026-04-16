package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/hook"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install clerk components",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installing clerk...")
		if err := hook.Install(); err != nil {
			return err
		}
		if err := mcpinstall.Install(); err != nil {
			return err
		}
		if err := commands.Install(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

var installHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Install clerk as Claude Code SessionStart/SessionEnd hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installing clerk hook...")
		if err := hook.Install(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

var installMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Register clerk as a Claude Code MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installing clerk mcp...")
		if err := mcpinstall.Install(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

var installSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Install clerk skills for Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installing clerk skills...")
		if err := commands.Install(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

func init() {
	installCmd.AddCommand(installHookCmd)
	installCmd.AddCommand(installMcpCmd)
	installCmd.AddCommand(installSkillsCmd)
	rootCmd.AddCommand(installCmd)
}
