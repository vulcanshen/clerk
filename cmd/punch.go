package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/punch"
)

var punchCmd = &cobra.Command{
	Use:   "punch",
	Short: "Record a session start (called by SessionStart hook)",
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

		return punch.Run(data, cfg)
	},
}

func init() {
	rootCmd.AddCommand(punchCmd)
}
