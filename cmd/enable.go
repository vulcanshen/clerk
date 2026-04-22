package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var enableCmd = &cobra.Command{
	Use:   "enable <slug>",
	Short: "Enable feed for a project",
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
		return setFeedEnabled(args[0], true)
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <slug>",
	Short: "Disable feed for a project",
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
		return setFeedEnabled(args[0], false)
	},
}

func setFeedEnabled(slug string, enabled bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cwd, err := feed.LoadSlugMeta(cfg, slug)
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(cwd, ".clerk.json")

	// Load existing or create new
	raw := make(map[string]interface{})
	if data, err := os.ReadFile(cfgPath); err == nil {
		json.Unmarshal(data, &raw)
	}

	feedSection, _ := raw["feed"].(map[string]interface{})
	if feedSection == nil {
		feedSection = make(map[string]interface{})
	}
	feedSection["enabled"] = enabled
	raw["feed"] = feedSection

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(cfgPath)
	os.MkdirAll(dir, 0755)

	tmp, err := os.CreateTemp(dir, ".clerk-config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing config: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpPath, cfgPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("saving config: %w", err)
	}

	action := "Enabled"
	if !enabled {
		action = "Disabled"
	}
	fmt.Fprintf(os.Stderr, "%s feed for %s (%s)\n", action, slug, cwd)
	return nil
}

func init() {
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
}
