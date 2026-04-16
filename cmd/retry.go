package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var retryAllFlag bool

var retryCmd = &cobra.Command{
	Use:   "retry [slug]",
	Short: "Retry interrupted feed processes",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		orphans, err := feed.OrphanStates(cfg)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var slugs []string
		for _, o := range orphans {
			slugs = append(slugs, o.State.Slug)
		}
		return slugs, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		orphans, err := feed.OrphanStates(cfg)
		if err != nil {
			return err
		}

		if len(orphans) == 0 {
			fmt.Println("No interrupted sessions to retry.")
			return nil
		}

		// filter by slug if specified
		if !retryAllFlag && len(args) == 1 {
			slug := args[0]
			var matched []feed.OrphanState
			for _, o := range orphans {
				if o.State.Slug == slug {
					matched = append(matched, o)
				}
			}
			if len(matched) == 0 {
				return fmt.Errorf("no interrupted session found for: %s", slug)
			}
			orphans = matched
		} else if !retryAllFlag && len(args) == 0 {
			return fmt.Errorf("specify a slug or use --all to retry all %d interrupted session(s)", len(orphans))
		}

		fmt.Printf("Retrying %d interrupted session(s)...\n\n", len(orphans))
		for _, o := range orphans {
			fmt.Printf("  %s ...", o.State.Slug)
			if err := feed.Retry(o, cfg); err != nil {
				fmt.Printf(" failed: %v\n", err)
				continue
			}
			fmt.Println(" done")
		}
		return nil
	},
}

func init() {
	retryCmd.Flags().BoolVarP(&retryAllFlag, "all", "a", false, "Retry all interrupted sessions")
	rootCmd.AddCommand(retryCmd)
}
