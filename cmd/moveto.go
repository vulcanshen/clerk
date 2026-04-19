package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var movetoCmd = &cobra.Command{
	Use:   "moveto <path>",
	Short: "Move clerk data to a new directory and update config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest := config.ExpandPath(args[0])

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		src := config.ExpandPath(cfg.Output.Dir)

		// check source exists
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", src)
		}

		// check dest is not the same as src
		absSrc, _ := filepath.Abs(src)
		absDest, _ := filepath.Abs(dest)
		if absSrc == absDest {
			return fmt.Errorf("destination is the same as current output directory")
		}

		// check dest doesn't already have data
		if entries, err := os.ReadDir(dest); err == nil && len(entries) > 0 {
			return fmt.Errorf("destination %s already exists and is not empty", dest)
		}

		// create dest
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("creating destination: %w", err)
		}

		// move contents (not the directory itself)
		entries, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("reading source directory: %w", err)
		}

		// track entries that were copied (not renamed) for deferred removal
		var copied []string

		for _, entry := range entries {
			srcPath := filepath.Join(src, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			// try rename first (fast, same filesystem)
			if err := os.Rename(srcPath, destPath); err != nil {
				// cross-filesystem: copy all first, remove sources after all succeed
				if err := copyEntry(srcPath, destPath, entry); err != nil {
					return fmt.Errorf("moving %s: %w", entry.Name(), err)
				}
				copied = append(copied, srcPath)
			}
		}

		// all copies succeeded, now remove sources
		for _, p := range copied {
			os.RemoveAll(p)
		}

		// remove empty source directory
		os.Remove(src)

		// update global config
		if err := config.Set("output.dir", args[0], true); err != nil {
			return fmt.Errorf("updating config: %w", err)
		}

		fmt.Printf("Moved data from %s to %s\n", src, dest)
		fmt.Printf("Updated output.dir to %s\n", args[0])
		return nil
	},
}

func copyEntry(src, dest string, entry fs.DirEntry) error {
	if entry.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest)
}

func copyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, info.Mode())
}

func copyDir(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

