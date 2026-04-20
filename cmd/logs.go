package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var logsDays int
var logsErrorOnly bool
var logsNoMask bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs for troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if logsDays < 1 || logsDays > 180 {
			return fmt.Errorf("--days must be between 1 and 180")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		lines, err := collectLogLines(cfg, logsDays, logsErrorOnly)
		if err != nil {
			return err
		}

		if len(lines) == 0 {
			if logsErrorOnly {
				fmt.Printf("No errors found in the last %d day(s).\n", logsDays)
			} else {
				fmt.Printf("No logs found in the last %d day(s).\n", logsDays)
			}
			return nil
		}

		if logsNoMask {
			for _, line := range lines {
				fmt.Println(line)
			}
		} else {
			fmt.Println(maskOutput(lines, cfg))
		}

		return nil
	},
}

func collectLogLines(cfg config.Config, days int, errorOnly bool) ([]string, error) {
	logDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "log")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading log directory: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -days+1).Format("20060102")

	var logFiles []string
	for _, e := range entries {
		name := e.Name()
		if len(name) < 8 || !strings.HasSuffix(name, "-clerk.log") {
			continue
		}
		if name[:8] >= cutoff {
			logFiles = append(logFiles, name)
		}
	}
	sort.Strings(logFiles)

	var lines []string
	for _, name := range logFiles {
		f, err := os.Open(filepath.Join(logDir, name))
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if errorOnly && !strings.Contains(line, "[ERROR]") {
				continue
			}
			lines = append(lines, line)
		}
		f.Close()
	}

	return lines, nil
}

func maskOutput(lines []string, cfg config.Config) string {
	raw := strings.Join(lines, "\n")

	prompt := fmt.Sprintf(`You are a log redaction tool. Replace any personally identifiable information in the following log lines with # symbols. This includes:
- Usernames in file paths (e.g. /Users/john/ → /Users/####/)
- Home directory names
- Any personal names, emails, or identifiers

Keep the log structure, timestamps, log levels, and error messages intact. Only mask the personal parts.
Output the redacted log lines only, no explanation.

%s`, raw)

	output, err := feed.CallClaude(prompt, cfg.Summary.Model, cfg.Summary.Timeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: --mask failed (Claude API error). Raw output suppressed to protect personal information.")
		return ""
	}
	return output
}

func init() {
	logsCmd.Flags().IntVar(&logsDays, "days", 1, "Number of days to include (default: today only)")
	logsCmd.Flags().BoolVar(&logsErrorOnly, "error", false, "Show only error logs")
	logsCmd.Flags().BoolVar(&logsNoMask, "no-mask", false, "Show raw logs without redacting personal information")
	rootCmd.AddCommand(logsCmd)
}
