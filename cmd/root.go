package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "clerk",
	Short: "The Claude Code Clerk — auto-summarize your sessions",
	Long:  "clerk is the Claude Code Clerk that automatically summarizes your sessions and saves them as organized markdown files.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.Version = Version
}
