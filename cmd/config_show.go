package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/logger"
)

var configShowJSON bool

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if configShowJSON {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		exe, _ := os.Executable()
		fmt.Printf("Executable:     %s\n", exe)
		fmt.Printf("Version:        %s\n", Version)
		fmt.Printf("Global config:  %s\n", config.GlobalConfigPath())
		fmt.Printf("Project config: %s\n", config.ProjectConfigPath(""))
		fmt.Printf("Log path:       %s\n\n", logger.LogPath(cfg))

		printOrNotSet := func(key, val string) {
			if val == "" {
				fmt.Printf("%-22s (not set)\n", key)
			} else {
				fmt.Printf("%-22s %s\n", key, val)
			}
		}

		printOrNotSet("output.dir", cfg.Output.Dir)
		printOrNotSet("output.language", cfg.Output.Language)
		printOrNotSet("summary.model", cfg.Summary.Model)
		printOrNotSet("summary.timeout", cfg.Summary.Timeout)
		printOrNotSet("summary.instruction", cfg.Summary.Instruction)
		fmt.Printf("%-22s %d\n", "log.retention_days", cfg.Log.RetentionDays)
		if cfg.Feed.Enabled != nil {
			fmt.Printf("%-22s %v\n", "feed.enabled", *cfg.Feed.Enabled)
		} else {
			fmt.Printf("%-22s true (default)\n", "feed.enabled")
		}
		return nil
	},
}

func init() {
	configShowCmd.Flags().BoolVar(&configShowJSON, "json", false, "Output in JSON format")
	configCmd.AddCommand(configShowCmd)
}
