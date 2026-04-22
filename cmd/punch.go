package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/punch"
)

var punchCmd = &cobra.Command{
	Use:               "punch",
	Short:             "Record a session start (called by SessionStart hook)",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CLERK_INTERNAL") == "1" {
			return nil
		}

		data, err := io.ReadAll(io.LimitReader(os.Stdin, 1024*1024)) // 1MB limit
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		// Parse hook input to get cwd for project-level config
		var input feed.HookInput
		if err := json.Unmarshal(data, &input); err != nil {
			return fmt.Errorf("parsing hook input: %w", err)
		}

		cfg, err := config.LoadWithCwd(input.Cwd)
		if err != nil {
			return err
		}

		return punch.Run(data, cfg)
	},
}

func init() {
	rootCmd.AddCommand(punchCmd)
}
