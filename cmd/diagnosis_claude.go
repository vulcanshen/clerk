package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/feed"
)

var diagnosisClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Test if clerk can communicate with Claude API",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Testing feed pipeline (BuildPrompt → CallClaude → ParseSummaryAndTags)...")
		fmt.Println()

		testConv := "[User]\nHello, this is a test.\n\n[Assistant]\nHi! How can I help?\n"
		testPrompt := feed.BuildPrompt(testConv, "", "en")
		testOut, err := feed.CallClaude(testPrompt, "", "1m")
		if err != nil {
			fmt.Printf("FAILED — claude -p error: %v\n", err)
			return nil
		}

		summary, tags := feed.ParseSummaryAndTags(testOut)
		if strings.TrimSpace(summary) == "" {
			fmt.Println("FAILED — empty summary")
			fmt.Println()
			fmt.Println("Claude API response format may have changed.")
			fmt.Println("Run 'clerk version' to check for updates.")
			fmt.Println("If already latest, please report at https://github.com/vulcanshen/clerk/issues")
		} else if len(tags) == 0 {
			fmt.Println("WARNING — summary OK but no tags extracted")
			fmt.Println("Tag format may have changed. Run 'clerk version' to check for updates.")
		} else {
			fmt.Printf("OK — summary generated + %d tags extracted\n", len(tags))
		}

		return nil
	},
}

func init() {
	diagnosisCmd.AddCommand(diagnosisClaudeCmd)
}
