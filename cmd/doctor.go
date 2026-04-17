package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/hook"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check if your environment is set up correctly",
	Run: func(cmd *cobra.Command, args []string) {
		issues := 0

		// Executable
		exe, _ := os.Executable()
		exe, _ = filepath.EvalSymlinks(exe)
		fmt.Printf("Executable:  %s\n", exe)
		fmt.Printf("Version:     %s\n", Version)

		// Claude CLI
		claudeOut, err := exec.Command("claude", "--version").Output()
		if err != nil {
			fmt.Printf("Claude CLI:  NOT FOUND — clerk requires Claude Code to be installed\n")
			issues++
		} else {
			fmt.Printf("Claude CLI:  OK (%s)\n", strings.TrimSpace(string(claudeOut)))
		}

		// Hook
		if hook.IsInstalled() {
			fmt.Printf("Hook:        OK\n")

			// Check hook path for backslashes (Windows issue)
			hookPath := checkHookPath()
			if hookPath != "" {
				fmt.Printf("Hook path:   WARNING — contains backslashes: %s\n", hookPath)
				fmt.Printf("             Run 'clerk install --force' to fix\n")
				issues++
			}
		} else {
			fmt.Printf("Hook:        NOT INSTALLED — run 'clerk install'\n")
			issues++
		}

		// MCP
		if mcpinstall.IsInstalled() {
			fmt.Printf("MCP:         OK\n")
		} else {
			fmt.Printf("MCP:         NOT INSTALLED — run 'clerk install'\n")
			issues++
		}

		// Skills
		if commands.IsInstalled() {
			fmt.Printf("Skills:      OK\n")
		} else {
			fmt.Printf("Skills:      NOT INSTALLED — run 'clerk install'\n")
			issues++
		}

		// Config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Config:      ERROR — %v\n", err)
			issues++
		} else {
			fmt.Printf("Config:      OK\n")
		}

		// Output dir
		if err == nil {
			outDir := config.ExpandPath(cfg.Output.Dir)
			if info, err := os.Stat(outDir); err != nil {
				fmt.Printf("Output dir:  NOT FOUND — %s (will be created on first feed)\n", outDir)
			} else if !info.IsDir() {
				fmt.Printf("Output dir:  ERROR — %s exists but is not a directory\n", outDir)
				issues++
			} else {
				fmt.Printf("Output dir:  OK (%s)\n", outDir)

				// Check for old hidden directories needing migration
				migrationNeeded := checkMigration(outDir)
				if migrationNeeded {
					fmt.Printf("Migration:   NEEDED — run 'clerk migrate'\n")
					issues++
				} else {
					fmt.Printf("Migration:   OK\n")
				}
			}
		}

		// Summary
		fmt.Println()
		if issues == 0 {
			fmt.Println("All checks passed.")
		} else {
			fmt.Printf("%d issue(s) found.\n", issues)
		}
	},
}

func checkHookPath() string {
	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".claude", "settings.json"))
	if err != nil {
		home, _ := os.UserHomeDir()
		data, err = os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
		if err != nil {
			return ""
		}
	}
	content := string(data)
	// Look for clerk commands with backslashes
	if strings.Contains(content, "clerk") && strings.Contains(content, "\\") {
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "clerk") && strings.Contains(line, "\\") {
				return strings.TrimSpace(line)
			}
		}
	}
	return ""
}

func checkMigration(outDir string) bool {
	oldDirs := []string{".tags", ".sessions", ".cursor", ".running", ".log"}
	for _, d := range oldDirs {
		if _, err := os.Stat(filepath.Join(outDir, d)); err == nil {
			return true
		}
	}
	// Check for YYYYMMDD dirs in root (should be in summary/)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && dateDirPattern.MatchString(e.Name()) {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
