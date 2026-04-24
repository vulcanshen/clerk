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
	Use:               "show",
	Short:             "Show current configuration",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if configShowJSON {
			cfg.Summary.Provider.APIKey = maskKey(cfg.Summary.Provider.APIKey)
			cfg.Report.Provider.APIKey = maskKey(cfg.Report.Provider.APIKey)
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
				fmt.Printf("%-28s (not set)\n", key)
			} else {
				fmt.Printf("%-28s %s\n", key, val)
			}
		}

		printOrNotSet("output.dir", cfg.Output.Dir)
		printOrNotSet("output.language", cfg.Output.Language)
		printOrNotSet("summary.provider.name", cfg.Summary.Provider.Name)
		printOrNotSet("summary.provider.model", cfg.Summary.Provider.Model)
		printOrNotSet("summary.provider.endpoint", cfg.Summary.Provider.Endpoint)
		printOrNotSet("summary.provider.api_key", maskKey(cfg.Summary.Provider.APIKey))
		printOrNotSet("summary.timeout", cfg.Summary.Timeout)
		printOrNotSet("summary.instruction", cfg.Summary.Instruction)
		printOrNotSet("report.provider.name", cfg.Report.Provider.Name)
		printOrNotSet("report.provider.model", cfg.Report.Provider.Model)
		printOrNotSet("report.provider.endpoint", cfg.Report.Provider.Endpoint)
		printOrNotSet("report.provider.api_key", maskKey(cfg.Report.Provider.APIKey))
		printOrNotSet("report.instruction", cfg.Report.Instruction)
		fmt.Printf("%-28s %d\n", "log.retention_days", cfg.Log.RetentionDays)
		if cfg.Feed.Enabled != nil {
			fmt.Printf("%-28s %v\n", "feed.enabled", *cfg.Feed.Enabled)
		} else {
			fmt.Printf("%-28s true (default)\n", "feed.enabled")
		}
		return nil
	},
}

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func init() {
	configShowCmd.Flags().BoolVar(&configShowJSON, "json", false, "Output in JSON format")
	configCmd.AddCommand(configShowCmd)
}
