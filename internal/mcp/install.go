package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vulcanshen/clerk/internal/mcpserver"
)

// CheckStatus runs `claude mcp list` once and returns whether clerk is
// registered (installed) and any path mismatch issue (empty string if OK).
func CheckStatus() (installed bool, issue string) {
	c := exec.Command("claude", "mcp", "list")
	output, err := c.Output()
	if err != nil {
		return false, ""
	}
	if !strings.Contains(string(output), "clerk") {
		return false, ""
	}

	// installed — now check path
	exe, err := os.Executable()
	if err != nil {
		return true, ""
	}
	exe = filepath.ToSlash(exe)
	if strings.Contains(string(output), exe) {
		return true, ""
	}

	// extract old path from output: "clerk: /old/path mcp - ✓ Connected"
	old := ""
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "clerk:") {
			parts := strings.TrimPrefix(line, "clerk:")
			parts = strings.TrimSpace(parts)
			if idx := strings.Index(parts, " - "); idx > 0 {
				old = parts[:idx]
			}
			break
		}
	}
	if old != "" {
		return true, fmt.Sprintf("MCP points to %s (expected %s mcp)", old, exe)
	}
	return true, fmt.Sprintf("MCP points to a different executable (expected %s mcp)", exe)
}

func ForceInstall() error {
	Uninstall()
	return Install()
}

func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable: %w", err)
	}
	// Don't resolve symlinks — keep package manager symlinks stable across upgrades
	exe = filepath.ToSlash(exe)

	c := exec.Command("claude", "mcp", "add", "--transport", "stdio", "--scope", "user", "clerk", "--", exe, "mcp")
	output, err := c.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "already exists") {
			fmt.Fprintln(os.Stderr, "  mcp:   already installed")
			return nil
		}
		return fmt.Errorf("failed to register MCP server: %s", strings.TrimSpace(string(output)))
	}

	// trigger tool registration to get names
	mcpserver.NewServer("")
	fmt.Fprintf(os.Stderr, "  mcp:   installed (%s)\n", strings.Join(mcpserver.ToolNames(), ", "))
	return nil
}

func Uninstall() error {
	c := exec.Command("claude", "mcp", "remove", "clerk")
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "  mcp:  not installed")
		return nil
	}

	fmt.Fprintln(os.Stderr, "  mcp:  removed")
	return nil
}
