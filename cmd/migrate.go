package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/hook"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var dateDirPattern = regexp.MustCompile(`^\d{8}$`)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate data directory structure to the latest format",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		root := config.ExpandPath(cfg.Output.Dir)
		migrated := 0

		// Migration 1: rename hidden directories to non-hidden
		if n, err := migrateHiddenDirs(root); err != nil {
			return err
		} else {
			migrated += n
		}

		// Migration 2: move YYYYMMDD dirs from root into summary/
		if n, err := migrateSummaryDirs(root); err != nil {
			return err
		} else {
			migrated += n
		}

		if migrated == 0 {
			fmt.Println("Nothing to migrate. Already up to date.")
		} else {
			fmt.Println("\nUpdating installed components...")
			if hook.IsInstalled() {
				hook.Install()
			}
			if mcpinstall.IsInstalled() {
				mcpinstall.Install()
			}
			if commands.IsInstalled() {
				commands.Install()
			}
		}
		return nil
	},
}

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

func init() {
	rootCmd.AddCommand(migrateCmd)
}
