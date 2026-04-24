package provider

import (
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
)

func isClaudeProvider(p string) bool {
	return p == "" || strings.ToLower(p) == "claude"
}

func applyPreset(pc config.ProviderConfig) config.ProviderConfig {
	if preset := FindPreset(strings.ToLower(pc.Name)); preset != nil {
		if pc.Endpoint == "" {
			pc.Endpoint = preset.Endpoint
		}
		if pc.Model == "" {
			pc.Model = preset.DefaultModel
		}
	}
	return pc
}

func resolveFromConfig(pc config.ProviderConfig) Provider {
	pc = applyPreset(pc)
	if isClaudeProvider(pc.Name) {
		return &ClaudeCliProvider{Model: pc.Model}
	}
	return &OpenAIProvider{
		Endpoint: pc.Endpoint,
		Model:    pc.Model,
		APIKey:   pc.APIKey,
	}
}

func fallback(value, fallbackValue string) string {
	if value != "" {
		return value
	}
	return fallbackValue
}

func mergeProvider(primary, secondary config.ProviderConfig) config.ProviderConfig {
	return config.ProviderConfig{
		Name:     fallback(primary.Name, secondary.Name),
		Model:    fallback(primary.Model, secondary.Model),
		Endpoint: fallback(primary.Endpoint, secondary.Endpoint),
		APIKey:   fallback(primary.APIKey, secondary.APIKey),
	}
}

// ResolveForSummary returns a Provider based on summary config.
func ResolveForSummary(cfg config.Config) Provider {
	return resolveFromConfig(cfg.Summary.Provider)
}

// ResolveForReport returns a Provider based on report config, falling back to summary config.
func ResolveForReport(cfg config.Config) Provider {
	merged := mergeProvider(cfg.Report.Provider, cfg.Summary.Provider)
	return resolveFromConfig(merged)
}
