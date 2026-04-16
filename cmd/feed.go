package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Process a Claude Code session transcript and generate a summary",
	Long:  "Reads SessionEnd hook JSON from stdin, extracts the conversation, and generates a summary using claude -p.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CLERK_INTERNAL") == "1" {
			return nil
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return feed.Run(data, cfg)
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)
}
