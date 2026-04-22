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

var unregisterCmd = &cobra.Command{
	Use:               "unregister",
	Short:             "Unregister clerk from Claude Code",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "Unregistering clerk...")
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
			fmt.Fprintf(os.Stderr, "\nAlso remove clerk data at %s? (y/N): ", outDir)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "y" || answer == "yes" {
				dirs := []string{"summary", "index", "tags", "sessions", "cursor", "running", "log"}
				for _, d := range dirs {
					path := filepath.Join(outDir, d)
					if _, err := os.Stat(path); err != nil {
						continue
					}
					if err := os.RemoveAll(path); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", d, err)
					}
				}
				fmt.Fprintf(os.Stderr, "Removed data from %s\n", outDir)
			}
		}

		fmt.Fprintln(os.Stderr, "\nDone.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
}
