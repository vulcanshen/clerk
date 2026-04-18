package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var dateDirPattern = regexp.MustCompile(`^\d{8}$`)

// Migration functions are called by diagnosis for auto-fix.

// migrateHiddenDirs renames .tags→tags, .sessions→sessions, .cursor→cursor, .running→running, .log→log.
func migrateHiddenDirs(root string) (int, error) {
	renames := [][2]string{
		{".tags", "index"},
		{".sessions", "sessions"},
		{".cursor", "cursor"},
		{".running", "running"},
		{".log", "log"},
	}

	renamed := 0
	for _, r := range renames {
		src := filepath.Join(root, r[0])
		dest := filepath.Join(root, r[1])

		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		// skip if dest already exists
		if _, err := os.Stat(dest); err == nil {
			continue
		}

		if err := os.Rename(src, dest); err != nil {
			return renamed, fmt.Errorf("renaming %s to %s: %w", r[0], r[1], err)
		}
		renamed++
	}

	if renamed > 0 {
		fmt.Printf("Renamed %d hidden directories to non-hidden\n", renamed)
	}
	return renamed, nil
}

// migrateSummaryDirs moves YYYYMMDD directories from root into summary/.
func migrateSummaryDirs(root string) (int, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading output directory: %w", err)
	}

	// find YYYYMMDD dirs in root
	var dateDirs []string
	for _, e := range entries {
		if e.IsDir() && dateDirPattern.MatchString(e.Name()) {
			dateDirs = append(dateDirs, e.Name())
		}
	}

	if len(dateDirs) == 0 {
		return 0, nil
	}

	summaryDir := filepath.Join(root, "summary")
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return 0, fmt.Errorf("creating summary directory: %w", err)
	}

	moved := 0
	for _, name := range dateDirs {
		src := filepath.Join(root, name)
		dest := filepath.Join(summaryDir, name)

		if err := os.Rename(src, dest); err != nil {
			return moved, fmt.Errorf("moving %s to summary/: %w", name, err)
		}
		moved++
	}

	fmt.Printf("Migrated %d date directories into summary/\n", moved)
	return moved, nil
}

// migrateTagsToIndex removes old tags/ directory and rebuilds index from summaries.
func migrateTagsToIndex(root string) (int, error) {
	tagsDir := filepath.Join(root, "tags")
	hasTags := false
	if _, err := os.Stat(tagsDir); err == nil {
		hasTags = true
		os.RemoveAll(tagsDir)
	}

	indexDir := filepath.Join(root, "index")
	hasIndex := false
	if _, err := os.Stat(indexDir); err == nil {
		entries, _ := os.ReadDir(indexDir)
		hasIndex = len(entries) > 0
	}

	// Only rebuild if we had old tags or no index exists yet
	if !hasTags && hasIndex {
		return 0, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return 0, err
	}

	rebuilt, err := rebuildIndex(root, cfg)
	if err != nil {
		return 0, err
	}

	if hasTags {
		fmt.Printf("Migrated tags/ to index/ (%d summaries rebuilt)\n", rebuilt)
	} else if rebuilt > 0 {
		fmt.Printf("Rebuilt index/ (%d summaries)\n", rebuilt)
	}

	if hasTags || rebuilt > 0 {
		return 1, nil
	}
	return 0, nil
}

// rebuildIndex scans all summaries and rebuilds index + updates frontmatter.
func rebuildIndex(root string, cfg config.Config) (int, error) {
	summaryRoot := filepath.Join(root, "summary")
	if _, err := os.Stat(summaryRoot); os.IsNotExist(err) {
		return 0, nil
	}

	dateDirs, err := os.ReadDir(summaryRoot)
	if err != nil {
		return 0, err
	}

	rebuilt := 0
	for _, dateEntry := range dateDirs {
		if !dateEntry.IsDir() || !dateDirPattern.MatchString(dateEntry.Name()) {
			continue
		}
		date := dateEntry.Name()
		dateDir := filepath.Join(summaryRoot, date)

		files, err := os.ReadDir(dateDir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
				continue
			}

			slug := strings.TrimSuffix(f.Name(), ".md")
			summaryPath := filepath.Join(dateDir, f.Name())

			// Read existing frontmatter tags
			existingTags := readFrontmatterTags(summaryPath)

			// Build full terms
			// We don't have cwd, but we can reconstruct slug from filename
			terms := feed.BuildTerms(existingTags, slugToCwdFallback(slug), date)

			// Update summary frontmatter with full terms
			updateSummaryFrontmatter(summaryPath, terms)

			// Build index entries
			feed.RebuildIndexForSummary(cfg, summaryPath, terms)

			rebuilt++
		}
	}

	return rebuilt, nil
}

// readFrontmatterTags extracts tags from YAML frontmatter.
func readFrontmatterTags(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	inTags := false
	var tags []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if inFrontmatter {
				break // end of frontmatter
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter {
			continue
		}
		if line == "tags:" {
			inTags = true
			continue
		}
		if inTags {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				tag := strings.TrimPrefix(trimmed, "- ")
				tags = append(tags, tag)
			} else {
				inTags = false
			}
		}
	}
	return tags
}

// updateSummaryFrontmatter rewrites the frontmatter with full terms.
func updateSummaryFrontmatter(path string, terms []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	content := string(data)

	// Strip existing frontmatter
	body := content
	if strings.HasPrefix(content, "---\n") {
		end := strings.Index(content[4:], "\n---\n")
		if end != -1 {
			body = content[end+4+4+1:] // skip past closing ---\n
		}
	}

	// Write new frontmatter + body
	var sb strings.Builder
	if len(terms) > 0 {
		sb.WriteString("---\ntags:\n")
		for _, t := range terms {
			fmt.Fprintf(&sb, "  - %s\n", t)
		}
		sb.WriteString("---\n\n")
	}
	sb.WriteString(strings.TrimLeft(body, "\n"))

	os.WriteFile(path, []byte(sb.String()), 0644)
}

// slugToCwdFallback creates a fake cwd from slug for BuildTerms.
// BuildTerms only uses cwd to compute CwdToSlug, so we pass a path that produces the slug.
func slugToCwdFallback(slug string) string {
	// CwdToSlug strips home prefix, lowercases, replaces / with -
	// We can't reverse this perfectly, but slug itself is fine for word extraction
	home, _ := os.UserHomeDir()
	return filepath.Join(home, slug)
}

