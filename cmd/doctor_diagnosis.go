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

var doctorDiagnosisCmd = &cobra.Command{
	Use:   "diagnosis",
	Short: "Show error logs for troubleshooting",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		logDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "log")
		entries, err := os.ReadDir(logDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No logs found.")
				return nil
			}
			return fmt.Errorf("reading log directory: %w", err)
		}

		cutoff := time.Now().AddDate(0, 0, -diagnosisDays+1).Format("20060102")

		// collect matching log files sorted by name (date)
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

		if len(logFiles) == 0 {
			fmt.Printf("No logs found in the last %d day(s).\n", diagnosisDays)
			return nil
		}

		found := 0
		for _, name := range logFiles {
			f, err := os.Open(filepath.Join(logDir, name))
			if err != nil {
				continue
			}

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "[ERROR]") {
					fmt.Println(line)
					found++
				}
			}
			f.Close()
		}

		if found == 0 {
			fmt.Printf("No errors found in the last %d day(s).\n", diagnosisDays)
		}

		return nil
	},
}

func init() {
	doctorDiagnosisCmd.Flags().IntVar(&diagnosisDays, "days", 1, "Number of days to include (default: today only)")
	doctorCmd.AddCommand(doctorDiagnosisCmd)
}
