package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var deleteDate string
var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete all data for a slug",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return listSlugs(cfg), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		root := config.ExpandPath(cfg.Output.Dir)
		targets := collectDeleteTargets(root, slug, deleteDate)

		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "No data found for slug %q\n", slug)
			return nil
		}

		// Show what will be deleted
		fmt.Fprintf(os.Stderr, "Will delete %d item(s) for slug %q:\n", len(targets), slug)
		for _, t := range targets {
			fmt.Fprintf(os.Stderr, "  %s\n", t)
		}

		if !deleteForce {
			fmt.Fprintf(os.Stderr, "\nProceed? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}
		}

		deleted := 0
		for _, t := range targets {
			if err := os.Remove(t); err == nil {
				deleted++
			}
		}

		// Clean up index entries
		cleanIndexEntries(root, slug, deleteDate)

		// Clean up empty date directories in summary/
		cleanEmptyDirs(filepath.Join(root, "summary"))

		fmt.Fprintf(os.Stderr, "Deleted %d item(s) for slug %q\n", deleted, slug)
		return nil
	},
}

func collectDeleteTargets(root, slug, date string) []string {
	var targets []string

	summaryDir := filepath.Join(root, "summary")
	if date != "" {
		// Only specific date
		path := filepath.Join(summaryDir, date, slug+".md")
		if _, err := os.Stat(path); err == nil {
			targets = append(targets, path)
		}
		// cursor for that date
		cursorPath := filepath.Join(root, "cursor", date+"-"+slug)
		if _, err := os.Stat(cursorPath); err == nil {
			targets = append(targets, cursorPath)
		}
		return targets
	}

	// All dates
	dateDirs, _ := os.ReadDir(summaryDir)
	for _, d := range dateDirs {
		if !d.IsDir() {
			continue
		}
		path := filepath.Join(summaryDir, d.Name(), slug+".md")
		if _, err := os.Stat(path); err == nil {
			targets = append(targets, path)
		}
	}

	// cursor files: YYYYMMDD-<slug>
	cursorDir := filepath.Join(root, "cursor")
	if entries, err := os.ReadDir(cursorDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), "-"+slug) {
				targets = append(targets, filepath.Join(cursorDir, e.Name()))
			}
		}
	}

	// sessions file
	sessionsPath := filepath.Join(root, "sessions", slug+".md")
	if _, err := os.Stat(sessionsPath); err == nil {
		targets = append(targets, sessionsPath)
	}

	// slug metadata
	slugMetaPath := filepath.Join(root, "slug", slug+".json")
	if _, err := os.Stat(slugMetaPath); err == nil {
		targets = append(targets, slugMetaPath)
	}

	return targets
}

func cleanIndexEntries(root, slug, date string) {
	indexDir := filepath.Join(root, "index")
	entries, err := os.ReadDir(indexDir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(indexDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		var kept []string
		changed := false
		for _, line := range lines {
			if line == "" {
				continue
			}
			// Index lines look like: - [slug+YYYYMMDD](../summary/YYYYMMDD/slug.md)
			if strings.Contains(line, "/"+slug+".md") {
				if date == "" || strings.Contains(line, "/"+date+"/") {
					changed = true
					continue
				}
			}
			kept = append(kept, line)
		}

		if !changed {
			continue
		}

		if len(kept) == 0 {
			os.Remove(path)
		} else {
			os.WriteFile(path, []byte(strings.Join(kept, "\n")+"\n"), 0644)
		}
	}
}

func cleanEmptyDirs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(dir, e.Name())
		subEntries, err := os.ReadDir(subDir)
		if err == nil && len(subEntries) == 0 {
			os.Remove(subDir)
		}
	}
}

func init() {
	deleteCmd.Flags().StringVar(&deleteDate, "date", "", "Only delete summary for a specific date (YYYYMMDD)")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")

	deleteCmd.RegisterFlagCompletionFunc("date", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return listDates(cfg), cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.AddCommand(deleteCmd)
}
