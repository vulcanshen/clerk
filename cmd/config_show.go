package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("Config path: %s\n\n", config.ConfigPath())

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
}
