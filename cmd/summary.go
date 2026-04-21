package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var summaryOutput string

var summaryCmd = &cobra.Command{
	Use:   "summary [slug]",
	Short: "View project summaries",
	Long:  "Without arguments, list all project slugs. With a slug, show merged summaries across all dates.",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		slugs := listSlugs(cfg)
		return slugs, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if len(args) == 0 {
			return listSummaries(cfg)
		}
		return showSummary(cfg, args[0])
	},
}

func listSlugs(cfg config.Config) []string {
	summaryDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")
	dateDirs, err := os.ReadDir(summaryDir)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	for _, d := range dateDirs {
		if !d.IsDir() || !dateDirPattern.MatchString(d.Name()) {
			continue
		}
		files, err := os.ReadDir(filepath.Join(summaryDir, d.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				seen[strings.TrimSuffix(f.Name(), ".md")] = true
			}
		}
	}

	slugs := make([]string, 0, len(seen))
	for s := range seen {
		slugs = append(slugs, s)
	}
	sort.Strings(slugs)
	return slugs
}

func listSummaries(cfg config.Config) error {
	slugs := listSlugs(cfg)
	if len(slugs) == 0 {
		fmt.Fprintln(os.Stderr, "No summaries found.")
		return nil
	}
	for _, s := range slugs {
		fmt.Println(s)
	}
	return nil
}

func showSummary(cfg config.Config, slug string) error {
	summaryDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")
	dateDirs, err := os.ReadDir(summaryDir)
	if err != nil {
		return fmt.Errorf("reading summary directory: %w", err)
	}

	type dateContent struct {
		date    string
		content string
	}
	var entries []dateContent

	for _, d := range dateDirs {
		if !d.IsDir() || !dateDirPattern.MatchString(d.Name()) {
			continue
		}
		path := filepath.Join(summaryDir, d.Name(), slug+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		body := stripFrontmatter(string(data))
		if strings.TrimSpace(body) == "" {
			continue
		}
		entries = append(entries, dateContent{date: d.Name(), content: body})
	}

	if len(entries) == 0 {
		return fmt.Errorf("no summaries found for slug %q", slug)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].date < entries[j].date
	})

	var sb strings.Builder
	for _, e := range entries {
		date := e.date[:4] + "-" + e.date[4:6] + "-" + e.date[6:]
		fmt.Fprintf(&sb, "## %s\n\n%s\n\n", date, strings.TrimSpace(e.content))
	}
	output := strings.TrimSpace(sb.String())

	// Output: same logic as report (terminal → file, pipe → stdout, -o → explicit)
	outPath := summaryOutput
	if outPath != "" {
		outPath = config.ExpandPath(outPath)
	} else if isStdoutTerminal() {
		outPath = defaultSummaryPath(cfg, slug)
	}

	if outPath != "" {
		dir := filepath.Dir(outPath)
		os.MkdirAll(dir, 0755)
		if err := atomicWriteFile(outPath, output+"\n"); err != nil {
			return fmt.Errorf("writing summary file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to %s\n", outPath)
	} else {
		fmt.Println(output)
	}
	return nil
}

func stripFrontmatter(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---\n")
	if end == -1 {
		return content
	}
	return content[4+end+5:]
}

func defaultSummaryPath(cfg config.Config, slug string) string {
	dir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "reports")
	base := "clerk-summary-" + slug

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

func atomicWriteFile(path, content string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".clerk-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func init() {
	summaryCmd.Flags().StringVarP(&summaryOutput, "output", "o", "", "Save to specific file")
	summaryCmd.MarkFlagFilename("output", "md")
	rootCmd.AddCommand(summaryCmd)
}
