package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vulcanshen/clerk/internal/mcpserver"
)

// IsInstalled checks if clerk MCP server is currently registered.
func IsInstalled() bool {
	c := exec.Command("claude", "mcp", "list")
	output, err := c.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "clerk")
}

// CheckPath verifies the registered MCP server points to the current executable.
// Returns empty string if OK, or a description of the problem.
func CheckPath() string {
	c := exec.Command("claude", "mcp", "list")
	output, err := c.Output()
	if err != nil {
		return ""
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe = filepath.ToSlash(exe)
	if !strings.Contains(string(output), exe) {
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
			return fmt.Sprintf("MCP points to %s (expected %s mcp)", old, exe)
		}
		return fmt.Sprintf("MCP points to a different executable (expected %s mcp)", exe)
	}
	return ""
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
			fmt.Println("  mcp:   already installed")
			return nil
		}
		return fmt.Errorf("failed to register MCP server: %s", strings.TrimSpace(string(output)))
	}

	// trigger tool registration to get names
	mcpserver.NewServer("")
	fmt.Printf("  mcp:   installed (%s)\n", strings.Join(mcpserver.ToolNames(), ", "))
	return nil
}

func Uninstall() error {
	c := exec.Command("claude", "mcp", "remove", "clerk")
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Println("  mcp:  not installed")
		return nil
	}

	fmt.Println("  mcp:  removed")
	return nil
}
