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
)

var diagnosisDays int
var diagnosisErrorMask bool

var diagnosisErrorCmd = &cobra.Command{
	Use:   "error",
	Short: "Show error logs for troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		lines, err := collectLogLines(cfg, diagnosisDays, true)
		if err != nil {
			return err
		}

		if len(lines) == 0 {
			fmt.Printf("No errors found in the last %d day(s).\n", diagnosisDays)
			return nil
		}

		if diagnosisErrorMask {
			fmt.Println(maskOutput(lines, cfg))
		} else {
			for _, line := range lines {
				fmt.Println(line)
			}
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

func init() {
	diagnosisErrorCmd.Flags().IntVar(&diagnosisDays, "days", 1, "Number of days to include (default: today only)")
	diagnosisErrorCmd.Flags().BoolVar(&diagnosisErrorMask, "mask", false, "Redact personal information using Claude (uses API calls)")
	diagnosisCmd.AddCommand(diagnosisErrorCmd)
}
