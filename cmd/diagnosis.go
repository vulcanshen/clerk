package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/hook"
	"github.com/vulcanshen/clerk/internal/logger"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var doctorCmd = &cobra.Command{
	Use:   "diagnosis",
	Short: "Check environment and auto-fix issues",
	Run: func(cmd *cobra.Command, args []string) {
		issues := 0
		fixed := 0

		// Executable
		exe, _ := os.Executable()
		fmt.Printf("Executable:  %s\n", exe)
		fmt.Printf("Version:     %s\n", Version)

		// Claude CLI
		claudeOut, err := exec.Command("claude", "--version").Output()
		if err != nil {
			fmt.Printf("Claude CLI:  NOT FOUND — clerk requires Claude Code to be installed\n")
			issues++
		} else {
			fmt.Printf("Claude CLI:  OK (%s)\n", strings.TrimSpace(string(claudeOut)))

			// Test feed pipeline
			fmt.Print("Feed test:   Run API test? This uses a small amount of tokens. (Y/n): ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "" || answer == "y" || answer == "yes" {
				testConv := "[User]\nHello, this is a test.\n\n[Assistant]\nHi! How can I help?\n"
				testPrompt := feed.BuildPrompt(testConv, "", "en")
				testOut, err := feed.CallClaude(testPrompt, "")
				if err != nil {
					fmt.Printf("Feed test:   FAILED — claude -p error: %v\n", err)
					issues++
				} else {
					summary, tags := feed.ParseSummaryAndTags(testOut)
					if strings.TrimSpace(summary) == "" {
						fmt.Printf("Feed test:   FAILED — empty summary\n")
						fmt.Printf("             Claude API response format may have changed.\n")
						fmt.Printf("             Run 'clerk version' to check for updates.\n")
						fmt.Printf("             If already latest, please report at https://github.com/vulcanshen/clerk/issues\n")
						issues++
					} else if len(tags) == 0 {
						fmt.Printf("Feed test:   WARNING — summary OK but no tags extracted\n")
						fmt.Printf("             Tag format may have changed. Run 'clerk version' to check for updates.\n")
					} else {
						fmt.Printf("Feed test:   OK (summary + %d tags)\n", len(tags))
					}
				}
			} else {
				fmt.Printf("Feed test:   SKIPPED\n")
			}
		}

		cfg, cfgErr := config.Load()

		// Hook
		if hook.IsInstalled() {
			hookIssue := checkHookPath()
			if hookIssue != "" {
				fmt.Printf("Hook:        FIXING — %s\n", hookIssue)
				if err := hook.ForceInstall(); err != nil {
					fmt.Printf("Hook:        FAILED — %v\n", err)
					if cfgErr == nil {
						logger.Errorf(cfg, "diagnosis: hook fix failed: %v", err)
					}
					issues++
				} else {
					fmt.Printf("Hook:        FIXED\n")
					fixed++
				}
			} else {
				// Verify hook binary exists
				binPath := extractHookBinary()
				if binPath != "" {
					if _, err := os.Stat(binPath); os.IsNotExist(err) {
						fmt.Printf("Hook:        FIXING — binary not found at %s\n", binPath)
						if err := hook.ForceInstall(); err != nil {
							fmt.Printf("Hook:        FAILED — %v\n", err)
							issues++
						} else {
							fmt.Printf("Hook:        FIXED\n")
							fixed++
						}
					} else {
						fmt.Printf("Hook:        OK\n")
					}
				} else {
					fmt.Printf("Hook:        OK\n")
				}
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
		if cfgErr != nil {
			fmt.Printf("Config:      ERROR — %v\n", cfgErr)
			issues++
		} else {
			fmt.Printf("Config:      OK\n")
		}

		// Output dir
		if cfgErr == nil {
			outDir := config.ExpandPath(cfg.Output.Dir)
			if info, err := os.Stat(outDir); err != nil {
				fmt.Printf("Output dir:  NOT FOUND — %s (will be created on first feed)\n", outDir)
			} else if !info.IsDir() {
				fmt.Printf("Output dir:  ERROR — %s exists but is not a directory\n", outDir)
				issues++
			} else {
				fmt.Printf("Output dir:  OK (%s)\n", outDir)

				// Auto-fix: rename hidden directories
				if n, err := migrateHiddenDirs(outDir); err != nil {
					fmt.Printf("Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "diagnosis: hidden dir migration failed: %v", err)
					issues++
				} else if n > 0 {
					fmt.Printf("Migration:   FIXED — renamed %d hidden directories\n", n)
					fixed++
				}

				// Auto-fix: move YYYYMMDD dirs into summary/
				if n, err := migrateSummaryDirs(outDir); err != nil {
					fmt.Printf("Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "diagnosis: summary dir migration failed: %v", err)
					issues++
				} else if n > 0 {
					fixed++
				}

				// Auto-fix: rename tags/ to index/
				if n, err := migrateTagsToIndex(outDir); err != nil {
					fmt.Printf("Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "diagnosis: tags to index migration failed: %v", err)
					issues++
				} else if n > 0 {
					fixed++
				}

				if !checkMigration(outDir) {
					fmt.Printf("Migration:   OK\n")
				}
			}
		}

		// Summary
		fmt.Println()
		if issues == 0 && fixed == 0 {
			fmt.Println("All checks passed.")
		} else if issues == 0 && fixed > 0 {
			fmt.Printf("Fixed %d issue(s). All checks passed now.\n", fixed)
		} else {
			fmt.Printf("%d issue(s) found, %d fixed.\n", issues, fixed)
			fmt.Println("If issues persist, run 'clerk diagnosis error --mask' and report to GitHub.")
		}
	},
}

func extractHookBinary() string {
	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".claude", "settings.json"))
	if err != nil {
		home, _ := os.UserHomeDir()
		data, err = os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
		if err != nil {
			return ""
		}
	}
	// Extract the first clerk command path from hooks
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "command") || !strings.Contains(line, "clerk") {
			continue
		}
		// format: "command": "path/to/clerk feed"
		parts := strings.SplitN(line, "\"", 5)
		if len(parts) >= 4 {
			cmd := parts[3]
			// extract binary path (first token before subcommand)
			fields := strings.Fields(cmd)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	return ""
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

	// Check for cmd.exe wrapper (outdated v3.4.0 format, breaks stdin)
	if strings.Contains(content, "cmd.exe") && strings.Contains(content, "clerk") {
		return "hook uses cmd.exe wrapper (outdated, breaks feed)"
	}

	// Check for resolved symlink paths (e.g. /opt/homebrew/Cellar/clerk/3.6.1/bin/clerk)
	// These break on brew upgrade because the version number changes
	if strings.Contains(content, "/Cellar/") && strings.Contains(content, "clerk") {
		return "hook path contains versioned Cellar path (breaks on brew upgrade)"
	}

	// Check for backslashes in clerk paths (Windows issue)
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "clerk") && strings.Contains(line, "\\") {
			// Ignore JSON escape sequences like \"
			unescaped := strings.ReplaceAll(line, "\\\"", "")
			if strings.Contains(unescaped, "\\") {
				return "hook path contains backslashes"
			}
		}
	}

	return ""
}

func checkMigration(outDir string) bool {
	oldDirs := []string{".tags", ".sessions", ".cursor", ".running", ".log", "tags"}
	for _, d := range oldDirs {
		if _, err := os.Stat(filepath.Join(outDir, d)); err == nil {
			return true
		}
	}
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
