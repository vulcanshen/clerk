package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var purgeYes bool

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Delete all clerk data (summaries, tags, sessions, logs, cursors)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		root := config.ExpandPath(cfg.Output.Dir)

		if !purgeYes {
			fmt.Printf("This will delete all clerk data in %s\n", root)
			fmt.Print("Are you sure? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		dirs := []string{"summary", "tags", "sessions", "cursor", "running", "log"}

		removed := 0
		for _, d := range dirs {
			path := root + "/" + d
			if err := os.RemoveAll(path); err == nil {
				removed++
			}
		}

		fmt.Printf("Purged %d directories from %s\n", removed, root)
		return nil
	},
}

func init() {
	purgeCmd.Flags().BoolVarP(&purgeYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(purgeCmd)
}
