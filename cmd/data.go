package cmd

import (
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:               "data",
	Short:             "Manage clerk data",
	ValidArgsFunction: noFileComp,
}

func init() {
	dataCmd.AddCommand(movetoCmd)
	rootCmd.AddCommand(dataCmd)
}
