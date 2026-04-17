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

func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}
	// Use forward slashes — Claude Code uses bash on all platforms
	exe = filepath.ToSlash(exe)

	c := exec.Command("claude", "mcp", "add", "--transport", "stdio", "--scope", "local", "clerk", "--", exe, "mcp")
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
