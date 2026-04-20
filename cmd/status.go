package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var watchFlag bool
var statusJSON bool

const (
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorReset = "\033[0m"
	colorDim   = "\033[2m"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running clerk feed processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if statusJSON {
			return printStatusJSON(cfg)
		}

		if !watchFlag {
			return printStatus(cfg)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		clearScreen()
		printStatus(cfg)

		for {
			select {
			case <-ticker.C:
				clearScreen()
				printStatus(cfg)
			case <-sig:
				fmt.Println()
				return nil
			}
		}
	},
}

func printStatusJSON(cfg config.Config) error {
	type jsonEntry struct {
		PID       int    `json:"pid,omitempty"`
		Slug      string `json:"slug"`
		StartedAt string `json:"started_at"`
		Status    string `json:"status"`
	}

	entries := make([]jsonEntry, 0)

	states, err := feed.RunningStates(cfg)
	if err != nil {
		return err
	}
	for _, s := range states {
		entries = append(entries, jsonEntry{
			PID:       s.PID,
			Slug:      s.Slug,
			StartedAt: s.StartedAt.Format(time.RFC3339),
			Status:    "active",
		})
	}

	orphans, err := feed.OrphanStates(cfg)
	if err != nil {
		return err
	}
	for _, o := range orphans {
		entries = append(entries, jsonEntry{
			Slug:      o.State.Slug,
			StartedAt: o.State.StartedAt,
			Status:    "interrupted",
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printStatus(cfg config.Config) error {
	now := time.Now().Format("15:04:05")

	states, err := feed.RunningStates(cfg)
	if err != nil {
		return err
	}

	orphans, err := feed.OrphanStates(cfg)
	if err != nil {
		return err
	}

	if len(states) == 0 && len(orphans) == 0 {
		fmt.Printf("No active clerk feed processes.%s%s%s\n", colorDim, pad(36, now), colorReset)
		return nil
	}

	if len(states) > 0 {
		header := fmt.Sprintf("Active (%d)", len(states))
		fmt.Printf("%s%s%s%s%s%s\n\n", colorBold, colorCyan, header, colorReset, colorDim, pad(len(header), now))
		fmt.Print(colorReset)
		for _, s := range states {
			elapsed := formatDuration(time.Since(s.StartedAt))
			started := s.StartedAt.Format("15:04:05")
			fmt.Printf("  PID %-8d %-28s %s %s(%s)%s\n", s.PID, s.Slug, elapsed, colorDim, started, colorReset)
		}
	}

	if len(orphans) > 0 {
		if len(states) > 0 {
			fmt.Println()
		}
		header := fmt.Sprintf("Interrupted (%d)", len(orphans))
		timeStr := ""
		if len(states) == 0 {
			timeStr = now
		}
		fmt.Printf("%s%s%s%s%s%s\n", colorBold, colorCyan, header, colorReset, colorDim, pad(len(header), timeStr))
		fmt.Print(colorReset)
		fmt.Printf("  Run 'clerk status retry' to recover.\n\n")
		for _, o := range orphans {
			started := ""
			if t, err := time.Parse(time.RFC3339, o.State.StartedAt); err == nil {
				started = t.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("  %-30s %s%s%s\n", o.State.Slug, colorDim, started, colorReset)
		}
	}

	return nil
}

func pad(headerLen int, timeStr string) string {
	if timeStr == "" {
		return ""
	}
	total := 60
	spaces := total - headerLen - len(timeStr)
	if spaces < 2 {
		spaces = 2
	}
	p := make([]byte, spaces)
	for i := range p {
		p[i] = ' '
	}
	return string(p) + timeStr
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func init() {
	statusCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "Watch mode: refresh every second")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output in JSON format")
	statusCmd.AddCommand(retryCmd)
	statusCmd.AddCommand(killCmd)
	rootCmd.AddCommand(statusCmd)
}
