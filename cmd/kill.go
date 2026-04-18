package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

var killAllFlag bool

var killCmd = &cobra.Command{
	Use:   "kill [slug]",
	Short: "Kill active clerk feed processes",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		states, err := feed.RunningStates(cfg)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var slugs []string
		for _, s := range states {
			slugs = append(slugs, s.Slug)
		}
		return slugs, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		states, err := feed.RunningStates(cfg)
		if err != nil {
			return err
		}

		if len(states) == 0 {
			fmt.Println("No active clerk feed processes.")
			return nil
		}

		if !killAllFlag && len(args) == 1 {
			slug := args[0]
			var matched []feed.RunningInfo
			for _, s := range states {
				if s.Slug == slug {
					matched = append(matched, s)
				}
			}
			if len(matched) == 0 {
				return fmt.Errorf("no active process found for: %s", slug)
			}
			states = matched
		} else if !killAllFlag && len(args) == 0 {
			return fmt.Errorf("specify a slug or use --all to kill all %d active process(es)", len(states))
		}

		for _, s := range states {
			proc, err := os.FindProcess(s.PID)
			if err != nil {
				fmt.Printf("  PID %-8d %s ... not found\n", s.PID, s.Slug)
				continue
			}
			if err := proc.Signal(os.Kill); err != nil {
				fmt.Printf("  PID %-8d %s ... failed: %v\n", s.PID, s.Slug, err)
				continue
			}
			fmt.Printf("  PID %-8d %s ... killed\n", s.PID, s.Slug)
		}

		fmt.Println("\nState files preserved. Run 'clerk status retry' to regenerate summaries.")
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killAllFlag, "all", "a", false, "Kill all active feed processes")
}
