package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var dateDirPattern = regexp.MustCompile(`^\d{8}$`)

// Migration functions are called by diagnosis for auto-fix.

// migrateHiddenDirs renames .tags→tags, .sessions→sessions, .cursor→cursor, .running→running, .log→log.
func migrateHiddenDirs(root string) (int, error) {
	renames := [][2]string{
		{".tags", "tags"},
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

