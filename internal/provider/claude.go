package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ClaudeCliProvider calls the Claude Code CLI (`claude -p`) for completions.
type ClaudeCliProvider struct {
	Model string
}

func (c *ClaudeCliProvider) Complete(ctx context.Context, prompt, systemPrompt string) (string, error) {
	args := []string{"-p"}
	if c.Model != "" {
		args = append(args, "--model", c.Model)
	}
	if systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(), "CLERK_INTERNAL=1")

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude -p timed out")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude exited with error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("running claude: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
