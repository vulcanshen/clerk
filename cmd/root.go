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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if home, err := os.UserHomeDir(); err != nil || home == "" {
			fmt.Fprintln(os.Stderr, "fatal: cannot determine home directory — is $HOME set?")
			os.Exit(1)
		}
	},
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

// noFileComp is a ValidArgsFunction that disables file/directory completion.
var noFileComp = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.Version = Version
	rootCmd.ValidArgsFunction = noFileComp
}
