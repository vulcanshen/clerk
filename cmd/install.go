package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/hook"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var forceInstall bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install clerk components",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := exec.LookPath("claude"); err != nil {
			fmt.Println("Warning: 'claude' not found in PATH. clerk requires Claude Code to function.")
			fmt.Println("Install Claude Code first: https://claude.ai/download")
			fmt.Println()
		}
		fmt.Println("Installing clerk...")
		if forceInstall {
			if err := hook.ForceInstall(); err != nil {
				return err
			}
			if err := mcpinstall.ForceInstall(); err != nil {
				return err
			}
			if err := commands.ForceInstall(); err != nil {
				return err
			}
		} else {
			if err := hook.Install(); err != nil {
				return err
			}
			if err := mcpinstall.Install(); err != nil {
				return err
			}
			if err := commands.Install(); err != nil {
				return err
			}
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
		if forceInstall {
			if err := hook.ForceInstall(); err != nil {
				return err
			}
		} else {
			if err := hook.Install(); err != nil {
				return err
			}
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
		if forceInstall {
			if err := mcpinstall.ForceInstall(); err != nil {
				return err
			}
		} else {
			if err := mcpinstall.Install(); err != nil {
				return err
			}
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
		if forceInstall {
			if err := commands.ForceInstall(); err != nil {
				return err
			}
		} else {
			if err := commands.Install(); err != nil {
				return err
			}
		}
		fmt.Println("\nDone.")
		return nil
	},
}

func init() {
	installCmd.PersistentFlags().BoolVar(&forceInstall, "force", false, "Force reinstall (uninstall then install)")
	installCmd.AddCommand(installHookCmd)
	installCmd.AddCommand(installMcpCmd)
	installCmd.AddCommand(installSkillsCmd)
	rootCmd.AddCommand(installCmd)
}
