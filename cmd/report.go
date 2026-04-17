package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/mcpserver"
)

var reportDays int
var reportRealtime bool

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report from recent summaries",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CLERK_INTERNAL") == "1" {
			return nil
		}

		if reportDays < 1 {
			return fmt.Errorf("--days must be at least 1")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if reportRealtime {
			fmt.Fprintln(os.Stderr, "Flushing active sessions (this will use additional Claude API calls)...")
			flushActiveSessions(cfg)
		}

		summaryDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")

		entries, err := os.ReadDir(summaryDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no summaries found. Run some sessions with clerk feed first")
			}
			return fmt.Errorf("reading summary directory: %w", err)
		}

		cutoff := time.Now().AddDate(0, 0, -reportDays+1).Format("20060102")

		// collect summaries grouped by date and slug
		type entry struct {
			date    string
			slug    string
			content string
		}
		var all []entry

		for _, e := range entries {
			if !e.IsDir() || len(e.Name()) != 8 {
				continue
			}
			if e.Name() < cutoff {
				continue
			}

			dateDir := filepath.Join(summaryDir, e.Name())
			files, err := os.ReadDir(dateDir)
			if err != nil {
				continue
			}

			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(dateDir, f.Name()))
				if err != nil {
					continue
				}
				slug := strings.TrimSuffix(f.Name(), ".md")
				all = append(all, entry{
					date:    e.Name(),
					slug:    slug,
					content: string(data),
				})
			}
		}

		if len(all) == 0 {
			return fmt.Errorf("no summaries found in the last %d day(s)", reportDays)
		}

		sort.Slice(all, func(i, j int) bool {
			if all[i].date != all[j].date {
				return all[i].date < all[j].date
			}
			return all[i].slug < all[j].slug
		})

		// build input for prompt
		var sb strings.Builder
		for _, e := range all {
			fmt.Fprintf(&sb, "## [%s] %s\n\n%s\n\n", e.date, e.slug, strings.TrimSpace(e.content))
		}

		startDate := all[0].date
		endDate := all[len(all)-1].date

		prompt := buildReportPrompt(sb.String(), startDate, endDate, cfg.Output.Language)

		output, err := feed.CallClaude(prompt, cfg.Summary.Model)
		if err != nil {
			return fmt.Errorf("claude -p failed: %w", err)
		}

		fmt.Println(output)
		return nil
	},
}

func formatDate(yyyymmdd string) string {
	t, err := time.Parse("20060102", yyyymmdd)
	if err != nil {
		return yyyymmdd
	}
	return t.Format("2006-01-02")
}

func buildReportPrompt(summaries string, startDate, endDate, language string) string {
	return fmt.Sprintf(`You are generating a work report from Claude Code session summaries.

Output language: %s
Time range: %s ~ %s

Organize the report into three sections:

### Summary (<start> ~ <end>)
Use the time range from above as <start> ~ <end>. A high-level overview of the entire time range. Summarize by project — what was accomplished overall for each project across all dates. Keep it concise, focus on outcomes and impact.
#### project-slug
One paragraph summary of what was accomplished.

### By Date
For each date, list each project and what was done. Structure:
#### YYYY-MM-DD
- **project-slug**: summary of work done
- **project-slug**: summary of work done

### By Project
For each project, list each date and what was done. Structure:
#### project-slug
- **YYYY-MM-DD**: summary of work done
- **YYYY-MM-DD**: summary of work done

Rules:
- Be concise and factual
- Prioritize by impact, not chronology
- Omit routine or trivial items
- Section titles and all content must be in the specified output language

---
Summaries:

%s`, language, formatDate(startDate), formatDate(endDate), summaries)
}

func flushActiveSessions(cfg config.Config) {
	allSessions, err := mcpserver.ReadAllSessionEntries(cfg)
	if err != nil {
		return
	}

	seen := make(map[string]bool)
	for _, entries := range allSessions {
		// iterate in reverse to get the latest entry per cwd
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			if entry.TranscriptPath == "" || entry.Cwd == "" {
				continue
			}
			if seen[entry.Cwd] {
				continue
			}
			seen[entry.Cwd] = true

			if _, err := os.Stat(entry.TranscriptPath); err != nil {
				continue
			}

			inputData, _ := json.Marshal(feed.HookInput{
				SessionID:      entry.SessionID,
				TranscriptPath: entry.TranscriptPath,
				Cwd:            entry.Cwd,
			})

			feed.Run(inputData, cfg)
		}
	}
}

func init() {
	reportCmd.Flags().IntVar(&reportDays, "days", 1, "Number of days to include (default: today only)")
	reportCmd.Flags().BoolVar(&reportRealtime, "realtime", false, "Include active sessions in real time (uses extra Claude API calls)")
	rootCmd.AddCommand(reportCmd)
}
