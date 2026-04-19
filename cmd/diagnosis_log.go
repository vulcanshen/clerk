package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var logDays int
var diagnosisLogMask bool

var diagnosisLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Show all logs for troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if logDays < 1 || logDays > 180 {
			return fmt.Errorf("--days must be between 1 and 180")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		lines, err := collectLogLines(cfg, logDays, false)
		if err != nil {
			return err
		}

		if len(lines) == 0 {
			fmt.Printf("No logs found in the last %d day(s).\n", logDays)
			return nil
		}

		if diagnosisLogMask {
			fmt.Println(maskOutput(lines, cfg))
		} else {
			for _, line := range lines {
				fmt.Println(line)
			}
		}

		return nil
	},
}

func init() {
	diagnosisLogCmd.Flags().IntVar(&logDays, "days", 1, "Number of days to include (default: today only)")
	diagnosisLogCmd.Flags().BoolVar(&diagnosisLogMask, "mask", false, "Redact personal information using Claude (uses API calls)")
	diagnosisCmd.AddCommand(diagnosisLogCmd)
}
