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

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	configShowCmd.Flags().BoolVar(&configShowJSON, "json", false, "Output in JSON format")
	configCmd.AddCommand(configShowCmd)
}
