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

var exportOutput string
var exportSlug string
var exportDate string

var exportCmd = &cobra.Command{
	Use:               "export",
	Short:             "Export summaries by project or date",
	ValidArgsFunction: noFileComp,
	Long: `Without flags, list available slugs and dates.
With --slug <slug>, merge all dates for a project.
With --date <YYYYMMDD>, merge all projects for a date.
With both flags, export a specific slug on a specific date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if exportSlug != "" && exportDate != "" {
			return exportSingle(cfg, exportSlug, exportDate)
		}
		if exportSlug != "" {
			return exportBySlug(cfg, exportSlug)
		}
		if exportDate != "" {
			return exportByDate(cfg, exportDate)
		}
		return listExportTargets(cfg)
	},
}

func listExportTargets(cfg config.Config) error {
	slugs := listSlugs(cfg)
	dates := listDates(cfg)

	if len(slugs) == 0 && len(dates) == 0 {
		fmt.Fprintln(os.Stderr, "No summaries found.")
		return nil
	}

	if len(slugs) > 0 {
		fmt.Fprintln(os.Stderr, "Slugs:")
		for _, s := range slugs {
			fmt.Printf("  %s\n", s)
		}
	}
	if len(dates) > 0 {
		if len(slugs) > 0 {
			fmt.Fprintln(os.Stderr)
		}
		fmt.Fprintln(os.Stderr, "Dates:")
		for _, d := range dates {
			fmt.Printf("  %s\n", d)
		}
	}
	return nil
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

func listDates(cfg config.Config) []string {
	summaryDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")
	dateDirs, err := os.ReadDir(summaryDir)
	if err != nil {
		return nil
	}

	var dates []string
	for _, d := range dateDirs {
		if d.IsDir() && dateDirPattern.MatchString(d.Name()) {
			dates = append(dates, d.Name())
		}
	}
	sort.Strings(dates)
	return dates
}

func exportSingle(cfg config.Config, slug, date string) error {
	if !dateDirPattern.MatchString(date) {
		return fmt.Errorf("invalid date format %q (expected YYYYMMDD)", date)
	}

	path := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary", date, slug+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no summary found for %s on %s", slug, date)
		}
		return fmt.Errorf("reading summary: %w", err)
	}

	body := stripFrontmatter(string(data))
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("no summary found for %s on %s", slug, date)
	}

	formattedDate := date[:4] + "-" + date[4:6] + "-" + date[6:]
	output := fmt.Sprintf("## %s — %s\n\n%s", formattedDate, slug, strings.TrimSpace(body))

	return writeExportOutput(cfg, output, fmt.Sprintf("clerk-export-%s-%s", slug, date))
}

func exportBySlug(cfg config.Config, slug string) error {
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

	return writeExportOutput(cfg, strings.TrimSpace(sb.String()), "clerk-export-"+slug)
}

func exportByDate(cfg config.Config, date string) error {
	if !dateDirPattern.MatchString(date) {
		return fmt.Errorf("invalid date format %q (expected YYYYMMDD)", date)
	}

	dateDir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary", date)
	files, err := os.ReadDir(dateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no summaries found for date %s", date)
		}
		return fmt.Errorf("reading date directory: %w", err)
	}

	type slugContent struct {
		slug    string
		content string
	}
	var entries []slugContent

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dateDir, f.Name()))
		if err != nil {
			continue
		}
		body := stripFrontmatter(string(data))
		if strings.TrimSpace(body) == "" {
			continue
		}
		slug := strings.TrimSuffix(f.Name(), ".md")
		entries = append(entries, slugContent{slug: slug, content: body})
	}

	if len(entries) == 0 {
		return fmt.Errorf("no summaries found for date %s", date)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].slug < entries[j].slug
	})

	formattedDate := date[:4] + "-" + date[4:6] + "-" + date[6:]
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", formattedDate)
	for _, e := range entries {
		fmt.Fprintf(&sb, "## %s\n\n%s\n\n", e.slug, strings.TrimSpace(e.content))
	}

	return writeExportOutput(cfg, strings.TrimSpace(sb.String()), "clerk-export-"+date)
}

func writeExportOutput(cfg config.Config, output, baseName string) error {
	outPath := exportOutput
	if outPath != "" {
		outPath = config.ExpandPath(outPath)
	} else if isStdoutTerminal() {
		outPath = defaultExportPath(cfg, baseName)
	}

	if outPath != "" {
		dir := filepath.Dir(outPath)
		os.MkdirAll(dir, 0755)
		if err := atomicWriteFile(outPath, output+"\n"); err != nil {
			return fmt.Errorf("writing export file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to %s\n", outPath)
	} else {
		fmt.Println(output)
	}
	return nil
}

func defaultExportPath(cfg config.Config, baseName string) string {
	dir := filepath.Join(config.ExpandPath(cfg.Output.Dir), "exports")

	path := filepath.Join(dir, baseName+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	for i := 1; ; i++ {
		path = filepath.Join(dir, fmt.Sprintf("%s-%d.md", baseName, i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
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
	exportCmd.Flags().StringVar(&exportSlug, "slug", "", "Export merged summaries for a slug (across all dates)")
	exportCmd.Flags().StringVar(&exportSlug, "summary", "", "Alias for --slug (deprecated)")
	exportCmd.Flags().MarkHidden("summary")
	exportCmd.Flags().StringVar(&exportDate, "date", "", "Export merged summaries for a date (across all slugs)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Save to specific file")
	exportCmd.MarkFlagFilename("output", "md")

	exportCmd.RegisterFlagCompletionFunc("slug", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return listSlugs(cfg), cobra.ShellCompDirectiveNoFileComp
	})
	exportCmd.RegisterFlagCompletionFunc("date", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return listDates(cfg), cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(exportCmd)
}
