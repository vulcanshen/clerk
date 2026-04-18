package cmd

import (
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Manage clerk data",
}

func init() {
	dataCmd.AddCommand(movetoCmd)
	dataCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(dataCmd)
}
