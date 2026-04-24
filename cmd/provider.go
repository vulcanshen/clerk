package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/provider"
)

var providerCmd = &cobra.Command{
	Use:               "provider",
	Short:             "List supported providers",
	ValidArgsFunction: noFileComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listProviders()
	},
}

var providerModelsCmd = &cobra.Command{
	Use:   "models <provider>",
	Short: "List available models for a provider",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var names []string
		for _, p := range provider.Presets() {
			if p.Name != "claude" {
				names = append(names, p.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return listModels(args[0])
	},
}

func listProviders() error {
	presets := provider.Presets()
	fmt.Fprintln(os.Stderr, "Supported providers:")
	for _, p := range presets {
		if p.Name == "claude" {
			fmt.Printf("  %-10s (default, uses Claude Code CLI)\n", p.Name)
		} else {
			fmt.Printf("  %-10s %s  [%s]\n", p.Name, p.Endpoint, p.DefaultModel)
		}
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Quick setup:")
	fmt.Fprintln(os.Stderr, "  clerk config set summary.provider.name <provider>")
	fmt.Fprintln(os.Stderr, "  clerk config set summary.provider.api_key <key>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Query available models:")
	fmt.Fprintln(os.Stderr, "  clerk provider models <provider>")
	return nil
}

func listModels(name string) error {
	name = strings.ToLower(name)
	preset := provider.FindPreset(name)

	cfg, _ := config.Load()

	endpoint := cfg.Summary.Provider.Endpoint
	apiKey := cfg.Summary.Provider.APIKey

	if endpoint == "" && preset != nil {
		endpoint = preset.Endpoint
	}
	if endpoint == "" {
		return fmt.Errorf("no endpoint for provider %q — set summary.provider.endpoint or use a known provider name", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := provider.ListModels(ctx, endpoint, apiKey)
	if err != nil {
		return fmt.Errorf("listing models for %s: %w", name, err)
	}

	if len(models) == 0 {
		fmt.Fprintln(os.Stderr, "No models found.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Available models for %s:\n", name)
	for _, m := range models {
		fmt.Printf("  %s\n", m)
	}
	return nil
}

func init() {
	providerCmd.AddCommand(providerModelsCmd)
	rootCmd.AddCommand(providerCmd)
}
