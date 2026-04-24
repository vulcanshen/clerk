package provider

import (
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
)

func isClaudeProvider(p string) bool {
	return p == "" || strings.ToLower(p) == "claude"
}

func resolve(provider, model, endpoint, apiKey string) Provider {
	if isClaudeProvider(provider) {
		return &ClaudeCliProvider{Model: model}
	}
	return &OpenAIProvider{
		Endpoint: endpoint,
		Model:    model,
		APIKey:   apiKey,
	}
}

func fallback(value, fallbackValue string) string {
	if value != "" {
		return value
	}
	return fallbackValue
}

// ResolveForSummary returns a Provider based on summary config.
func ResolveForSummary(cfg config.Config) Provider {
	return resolve(cfg.Summary.Provider, cfg.Summary.Model, cfg.Summary.Endpoint, cfg.Summary.APIKey)
}

// ResolveForReport returns a Provider based on report config, falling back to summary config.
func ResolveForReport(cfg config.Config) Provider {
	p := fallback(cfg.Report.Provider, cfg.Summary.Provider)
	m := fallback(cfg.Report.Model, cfg.Summary.Model)
	ep := fallback(cfg.Report.Endpoint, cfg.Summary.Endpoint)
	key := fallback(cfg.Report.APIKey, cfg.Summary.APIKey)
	return resolve(p, m, ep, key)
}
