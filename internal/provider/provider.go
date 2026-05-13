package provider

import "context"

// Provider abstracts the AI completion backend.
type Provider interface {
	Complete(ctx context.Context, prompt, systemPrompt string) (string, error)
}
