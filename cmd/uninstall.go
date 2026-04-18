package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/hook"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall clerk components",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling clerk...")
		if err := hook.Uninstall(); err != nil {
			return err
		}
		if err := mcpinstall.Uninstall(); err != nil {
			return err
		}
		if err := commands.Uninstall(); err != nil {
			return err
		}
		cfg, cfgErr := config.Load()
		if cfgErr == nil {
			outDir := config.ExpandPath(cfg.Output.Dir)
			fmt.Printf("\nAlso remove clerk data at %s? (y/N): ", outDir)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "y" || answer == "yes" {
				dirs := []string{"summary", "index", "tags", "sessions", "cursor", "running", "log"}
				for _, d := range dirs {
					if err := os.RemoveAll(filepath.Join(outDir, d)); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", d, err)
					}
				}
				fmt.Printf("Removed data from %s\n", outDir)
			}
		}

		fmt.Println("\nDone.")
		return nil
	},
}

var uninstallHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Remove clerk from Claude Code SessionStart/SessionEnd hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling clerk hook...")
		if err := hook.Uninstall(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

var uninstallMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Remove clerk from Claude Code MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling clerk mcp...")
		if err := mcpinstall.Uninstall(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

var uninstallSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Remove clerk skills from Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling clerk skills...")
		if err := commands.Uninstall(); err != nil {
			return err
		}
		fmt.Println("\nDone.")
		return nil
	},
}

func init() {
	uninstallCmd.AddCommand(uninstallHookCmd)
	uninstallCmd.AddCommand(uninstallMcpCmd)
	uninstallCmd.AddCommand(uninstallSkillsCmd)
	rootCmd.AddCommand(uninstallCmd)
}
