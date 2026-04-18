package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/commands"
	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/hook"
	"github.com/vulcanshen/clerk/internal/logger"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
)

var diagnosisCmd = &cobra.Command{
	Use:   "diagnosis",
	Short: "Check environment and auto-fix issues",
	Run: func(cmd *cobra.Command, args []string) {
		issues := 0
		fixed := 0

		// Executable
		exe, err := os.Executable()
		if err != nil || exe == "" {
			fmt.Printf("Executable:  (unable to detect)\n")
		} else {
			fmt.Printf("Executable:  %s\n", exe)
		}
		fmt.Printf("Version:     %s\n", Version)

		// Claude CLI
		claudeOut, err := exec.Command("claude", "--version").Output()
		if err != nil {
			fmt.Printf("Claude CLI:  NOT FOUND — clerk requires Claude Code to be installed\n")
			issues++
		} else {
			fmt.Printf("Claude CLI:  OK (%s)\n", strings.TrimSpace(string(claudeOut)))
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
					fmt.Printf("Hook:        FIXED → %s\n", extractHookBinary())
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
							fmt.Printf("Hook:        FIXED → %s\n", extractHookBinary())
							fixed++
						}
					} else {
						fmt.Printf("Hook:        OK (%s)\n", binPath)
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
			fmt.Printf("MCP:         OK (%s mcp)\n", exe)
		} else {
			fmt.Printf("MCP:         FIXING — not registered\n")
			if err := mcpinstall.ForceInstall(); err != nil {
				fmt.Printf("MCP:         FAILED — %v\n", err)
				if cfgErr == nil {
					logger.Errorf(cfg, "diagnosis: mcp fix failed: %v", err)
				}
				issues++
			} else {
				fmt.Printf("MCP:         FIXED → %s mcp\n", exe)
				fixed++
			}
		}

		// Skills
		skillsDir := commands.SkillsDir()
		if commands.IsInstalled() {
			fmt.Printf("Skills:      OK (%s)\n", skillsDir)
		} else {
			fmt.Printf("Skills:      FIXING — not installed\n")
			if err := commands.Install(); err != nil {
				fmt.Printf("Skills:      FAILED — %v\n", err)
				if cfgErr == nil {
					logger.Errorf(cfg, "diagnosis: skills fix failed: %v", err)
				}
				issues++
			} else {
				fmt.Printf("Skills:      FIXED → %s\n", skillsDir)
				fixed++
			}
		}

		// Output dir
		if cfgErr != nil {
			fmt.Printf("Config:      ERROR — %v\n", cfgErr)
			issues++
		} else {
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

			// Config details
			fmt.Println()
			fmt.Printf("Config files:\n")
			fmt.Printf("  global:    %s\n", config.GlobalConfigPath())
			projectCfg := config.ProjectConfigPath("")
			if _, err := os.Stat(projectCfg); err == nil {
				fmt.Printf("  project:   %s\n", projectCfg)
			}
			fmt.Println()
			fmt.Printf("Config values:\n")
			for _, s := range config.LoadSources() {
				if s.Value == "" {
					fmt.Printf("  %-20s (not set)\n", s.Key)
				} else {
					fmt.Printf("  %-20s %s  ← %s\n", s.Key, s.Value, s.Source)
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

func readSettingsJSON() (map[string]interface{}, error) {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".claude", "settings.json"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			continue
		}
		return settings, nil
	}
	return nil, fmt.Errorf("settings.json not found")
}

func extractHookCommands(settings map[string]interface{}) []string {
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return nil
	}
	var cmds []string
	for _, event := range []string{"SessionStart", "SessionEnd"} {
		entries, _ := hooks[event].([]interface{})
		for _, entry := range entries {
			entryMap, _ := entry.(map[string]interface{})
			if entryMap == nil {
				continue
			}
			hooksList, _ := entryMap["hooks"].([]interface{})
			for _, h := range hooksList {
				hMap, _ := h.(map[string]interface{})
				if hMap == nil {
					continue
				}
				cmd, _ := hMap["command"].(string)
				if cmd != "" {
					cmds = append(cmds, cmd)
				}
			}
		}
	}
	return cmds
}

func extractHookBinary() string {
	settings, err := readSettingsJSON()
	if err != nil {
		return ""
	}
	for _, cmd := range extractHookCommands(settings) {
		if strings.Contains(cmd, "clerk") {
			fields := strings.Fields(cmd)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	return ""
}

func checkHookPath() string {
	settings, err := readSettingsJSON()
	if err != nil {
		return ""
	}
	for _, cmd := range extractHookCommands(settings) {
		if !strings.Contains(cmd, "clerk") {
			continue
		}
		if strings.Contains(cmd, "cmd.exe") {
			return "hook uses cmd.exe wrapper (outdated, breaks feed)"
		}
		if strings.Contains(cmd, "/Cellar/") {
			return "hook path contains versioned Cellar path (breaks on brew upgrade)"
		}
		if strings.Contains(cmd, "\\") {
			return "hook path contains backslashes"
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
	rootCmd.AddCommand(diagnosisCmd)
}
