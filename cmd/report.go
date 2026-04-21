package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/logger"
	"github.com/vulcanshen/clerk/internal/mcpserver"
	"github.com/vulcanshen/clerk/internal/progress"
)

var reportDays int
var reportActive bool
var reportOutput string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report from recent summaries",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CLERK_INTERNAL") == "1" {
			return nil
		}

		if reportDays < 1 || reportDays > 180 {
			return fmt.Errorf("--days must be between 1 and 180")
		}

		if reportOutput != "" {
			reportOutput = config.ExpandPath(reportOutput)
		}

		p := progress.New()

		// Step 1: Load config
		p.Start("Loading config")
		cfg, err := config.Load()
		if err != nil {
			p.Fail(err)
			return fmt.Errorf("loading config: %w", err)
		}
		p.Done()

		// Step 2: Flush active sessions (conditional)
		if reportActive {
			p.Start("Flushing active sessions")
			flushActiveSessions(cfg)
			p.Done()
		}

		// Step 3: Read summaries
		p.Start("Reading summaries")
		summaryDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")

		dirEntries, err := os.ReadDir(summaryDir)
		if err != nil {
			if os.IsNotExist(err) {
				p.Fail(fmt.Errorf("no summaries found"))
				return fmt.Errorf("no summaries found. Run some sessions with clerk feed first")
			}
			p.Fail(err)
			return fmt.Errorf("reading summary directory: %w", err)
		}

		cutoff := time.Now().AddDate(0, 0, -reportDays+1).Format("20060102")

		type entry struct {
			date    string
			slug    string
			content string
		}
		var all []entry

		for _, e := range dirEntries {
			if !e.IsDir() || !dateDirPattern.MatchString(e.Name()) {
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
			p.Fail(fmt.Errorf("no summaries in the last %d day(s)", reportDays))
			return fmt.Errorf("no summaries found in the last %d day(s)", reportDays)
		}

		sort.Slice(all, func(i, j int) bool {
			if all[i].date != all[j].date {
				return all[i].date < all[j].date
			}
			return all[i].slug < all[j].slug
		})

		startDate := all[0].date
		endDate := all[len(all)-1].date
		p.Done()

		// Step 4: Build prompt
		p.Start("Building prompt")
		var sb strings.Builder
		for _, e := range all {
			fmt.Fprintf(&sb, "## [%s] %s\n\n%s\n\n", e.date, e.slug, strings.TrimSpace(e.content))
		}
		prompt := buildReportPrompt(sb.String(), startDate, endDate, cfg.Output.Language)
		p.Done()

		// Step 5: Call Claude
		p.Start(fmt.Sprintf("Generating report (%d summaries, %s ~ %s)", len(all), formatDate(startDate), formatDate(endDate)))
		output, err := feed.CallClaude(prompt, cfg.Summary.Model, cfg.Summary.Timeout, "")
		if err != nil {
			p.Fail(err)
			logger.Errorf(cfg, "report: claude -p failed: %v", err)
			return fmt.Errorf("claude -p failed: %w", err)
		}
		p.Done()

		// Determine output target
		outPath := reportOutput
		if outPath == "" && isStdoutTerminal() {
			outPath = defaultReportPath(cfg, reportDays)
		}

		// Output
		if outPath != "" {
			p.Start(fmt.Sprintf("Saving to %s", outPath))
			os.MkdirAll(filepath.Dir(outPath), 0755)
			if err := atomicWriteFile(outPath, output+"\n"); err != nil {
				p.Fail(err)
				logger.Errorf(cfg, "report: write failed: %v", err)
				return fmt.Errorf("writing report file: %w", err)
			}
			p.Done()
			fmt.Fprintf(os.Stderr, "Saved to %s\n", outPath)
		} else {
			fmt.Println(output)
		}
		return nil
	},
}

func isStdoutTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func defaultReportPath(cfg config.Config, days int) string {
	dir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "reports")
	date := time.Now().Format("20060102")
	base := fmt.Sprintf("clerk-report-%s-%dd", date, days)

	path := filepath.Join(dir, base+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	for i := 1; ; i++ {
		path = filepath.Join(dir, fmt.Sprintf("%s-%d.md", base, i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
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
		fmt.Fprintf(os.Stderr, "Warning: could not read session entries: %v\n", err)
		return
	}

	// Collect unique sessions to flush
	type flushJob struct {
		inputData []byte
		projCfg   config.Config
		cwd       string
	}

	seen := make(map[string]bool)
	var jobs []flushJob
	for _, entries := range allSessions {
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

			projCfg, err := config.LoadWithCwd(entry.Cwd)
			if err != nil || !config.IsFeedEnabled(projCfg) {
				continue
			}

			inputData, err := json.Marshal(feed.HookInput{
				SessionID:      entry.SessionID,
				TranscriptPath: entry.TranscriptPath,
				Cwd:            entry.Cwd,
			})
			if err != nil {
				continue
			}

			jobs = append(jobs, flushJob{inputData: inputData, projCfg: projCfg, cwd: entry.Cwd})
		}
	}

	// Run feed.Run calls in parallel with bounded concurrency
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(j flushJob) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := feed.Run(j.inputData, j.projCfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to flush session for %s: %v\n", j.cwd, err)
			}
		}(job)
	}
	wg.Wait()
}

func init() {
	reportCmd.Flags().IntVar(&reportDays, "days", 1, "Number of days to include (default: today only)")
	reportCmd.Flags().BoolVar(&reportActive, "active", false, "Include active sessions (uses extra Claude API calls)")
	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "Save report to specific file (default: auto-save to output.dir/reports/)")
	reportCmd.MarkFlagFilename("output", "md", "txt")
	rootCmd.AddCommand(reportCmd)
}
