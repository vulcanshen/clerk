package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/platform"
)

var feedInternalFlag bool
var feedInputFlag string

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

		// If called without --internal, fork to background and return immediately.
		// This prevents Claude Code from cancelling the hook.
		if !feedInternalFlag {
			return forkFeed(data)
		}

		// --internal: do the actual work

		// read from file if specified
		if feedInputFlag != "" {
			data, err = os.ReadFile(feedInputFlag)
			if err != nil {
				return fmt.Errorf("reading input file: %w", err)
			}
			os.Remove(feedInputFlag)
		}

		// parse hook input to get cwd for project-level config
		hookInput, err := feed.ParseHookInput(data)
		if err != nil {
			return err
		}

		cfg, err := config.LoadWithCwd(hookInput.Cwd)
		if err != nil {
			return err
		}

		if !config.IsFeedEnabled(cfg) {
			return nil
		}

		return feed.Run(data, cfg)
	},
}

func forkFeed(data []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable: %w", err)
	}

	// write stdin data to temp file so the child can read it after parent exits
	tmp, err := os.CreateTemp("", "clerk-feed-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	child := exec.Command(exe, "feed", "--internal", "--input", tmp.Name())
	platform.DetachProcess(child)

	if err := child.Start(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("forking feed process: %w", err)
	}

	// detach — don't wait for child
	return nil
}

func init() {
	feedCmd.Flags().BoolVar(&feedInternalFlag, "internal", false, "Run feed directly (used by fork)")
	feedCmd.Flags().StringVar(&feedInputFlag, "input", "", "Read input from file instead of stdin (used by fork)")
	feedCmd.Flags().MarkHidden("internal")
	feedCmd.Flags().MarkHidden("input")
	rootCmd.AddCommand(feedCmd)
}
