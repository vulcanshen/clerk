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
	"github.com/vulcanshen/clerk/internal/feed"
	"github.com/vulcanshen/clerk/internal/hook"
	"github.com/vulcanshen/clerk/internal/logger"
	mcpinstall "github.com/vulcanshen/clerk/internal/mcp"
	"github.com/vulcanshen/clerk/internal/progress"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register clerk with Claude Code and verify environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		issues := 0
		fixed := 0

		w := os.Stderr

		// Executable
		exe, err := os.Executable()
		if err != nil || exe == "" {
			fmt.Fprintf(w, "Executable:  (unable to detect)\n")
		} else {
			fmt.Fprintf(w, "Executable:  %s\n", exe)
		}
		fmt.Fprintf(w, "Version:     %s\n", Version)

		// Claude CLI
		claudeOut, err := exec.Command("claude", "--version").Output()
		if err != nil {
			fmt.Fprintf(w, "Claude CLI:  NOT FOUND — clerk requires Claude Code to be installed\n")
			issues++
		} else {
			fmt.Fprintf(w, "Claude CLI:  OK (%s)\n", strings.TrimSpace(string(claudeOut)))
		}

		cfg, cfgErr := config.Load()

		// Read settings.json once for all hook checks
		settings, _ := readSettingsJSON()

		// Hook
		if hook.IsInstalled() {
			hookIssue := checkHookPathWith(settings)
			if hookIssue != "" {
				fmt.Fprintf(w, "Hook:        FIXING — %s\n", hookIssue)
				if err := hook.ForceInstall(); err != nil {
					fmt.Fprintf(w, "Hook:        FAILED — %v\n", err)
					if cfgErr == nil {
						logger.Errorf(cfg, "register: hook fix failed: %v", err)
					}
					issues++
				} else {
					fresh, _ := readSettingsJSON()
					fmt.Fprintf(w, "Hook:        FIXED → %s\n", extractHookBinaryFrom(fresh))
					fixed++
				}
			} else {
				binPath := extractHookBinaryFrom(settings)
				if binPath != "" {
					if _, err := os.Stat(binPath); os.IsNotExist(err) {
						fmt.Fprintf(w, "Hook:        FIXING — binary not found at %s\n", binPath)
						if err := hook.ForceInstall(); err != nil {
							fmt.Fprintf(w, "Hook:        FAILED — %v\n", err)
							issues++
						} else {
							fresh, _ := readSettingsJSON()
							fmt.Fprintf(w, "Hook:        FIXED → %s\n", extractHookBinaryFrom(fresh))
							fixed++
						}
					} else {
						fmt.Fprintf(w, "Hook:        OK (%s)\n", binPath)
					}
				} else {
					fmt.Fprintf(w, "Hook:        OK\n")
				}
			}
		} else {
			fmt.Fprintf(w, "Hook:        FIXING — not installed\n")
			if err := hook.ForceInstall(); err != nil {
				fmt.Fprintf(w, "Hook:        FAILED — %v\n", err)
				if cfgErr == nil {
					logger.Errorf(cfg, "register: hook fix failed: %v", err)
				}
				issues++
			} else {
				fresh, _ := readSettingsJSON()
				fmt.Fprintf(w, "Hook:        FIXED → %s\n", extractHookBinaryFrom(fresh))
				fixed++
			}
		}

		// MCP (single `claude mcp list` call)
		mcpInstalled, mcpIssue := mcpinstall.CheckStatus()
		if !mcpInstalled {
			mcpIssue = "not registered"
		}
		if mcpIssue == "" {
			if exe != "" {
				fmt.Fprintf(w, "MCP:         OK (%s mcp)\n", exe)
			} else {
				fmt.Fprintf(w, "MCP:         OK\n")
			}
		} else {
			fmt.Fprintf(w, "MCP:         FIXING — %s\n", mcpIssue)
			if err := mcpinstall.ForceInstall(); err != nil {
				fmt.Fprintf(w, "MCP:         FAILED — %v\n", err)
				if cfgErr == nil {
					logger.Errorf(cfg, "register: mcp fix failed: %v", err)
				}
				issues++
			} else {
				if exe != "" {
					fmt.Fprintf(w, "MCP:         FIXED → %s mcp\n", exe)
				} else {
					fmt.Fprintf(w, "MCP:         FIXED\n")
				}
				fixed++
			}
		}

		// Skills — always write latest content
		skillsDir := commands.SkillsDir()
		if err := commands.WriteSkills(); err != nil {
			fmt.Fprintf(w, "Skills:      FAILED — %v\n", err)
			if cfgErr == nil {
				logger.Errorf(cfg, "register: skills fix failed: %v", err)
			}
			issues++
		} else {
			fmt.Fprintf(w, "Skills:      OK (%s)\n", skillsDir)
		}

		// Output dir
		if cfgErr != nil {
			fmt.Fprintf(w, "Config:      ERROR — %v\n", cfgErr)
			issues++
		} else {
			outDir := config.ExpandPath(cfg.Output.Dir)
			if info, err := os.Stat(outDir); err != nil {
				fmt.Fprintf(w, "Output dir:  NOT FOUND — %s (will be created on first feed)\n", outDir)
			} else if !info.IsDir() {
				fmt.Fprintf(w, "Output dir:  ERROR — %s exists but is not a directory\n", outDir)
				issues++
			} else {
				fmt.Fprintf(w, "Output dir:  OK (%s)\n", outDir)

				if n, err := migrateHiddenDirs(outDir); err != nil {
					fmt.Fprintf(w, "Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "register: hidden dir migration failed: %v", err)
					issues++
				} else if n > 0 {
					fmt.Fprintf(w, "Migration:   FIXED — renamed %d hidden directories\n", n)
					fixed++
				}

				if n, err := migrateSummaryDirs(outDir); err != nil {
					fmt.Fprintf(w, "Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "register: summary dir migration failed: %v", err)
					issues++
				} else if n > 0 {
					fmt.Fprintf(w, "Migration:   FIXED — moved %d date directories into summary/\n", n)
					fixed++
				}

				if n, err := migrateTagsToIndex(outDir); err != nil {
					fmt.Fprintf(w, "Migration:   FAILED — %v\n", err)
					logger.Errorf(cfg, "register: tags to index migration failed: %v", err)
					issues++
				} else if n > 0 {
					fmt.Fprintf(w, "Migration:   FIXED — rebuilt index from summaries\n")
					fixed++
				}

				if !checkMigration(outDir) {
					fmt.Fprintf(w, "Migration:   OK\n")
				}
			}

			// Config details
			fmt.Fprintln(w)
			fmt.Fprintf(w, "Config files:\n")
			fmt.Fprintf(w, "  global:    %s\n", config.GlobalConfigPath())
			projectCfg := config.ProjectConfigPath("")
			if _, err := os.Stat(projectCfg); err == nil {
				fmt.Fprintf(w, "  project:   %s\n", projectCfg)
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "Config values:\n")
			for _, s := range config.LoadSources() {
				if s.Value == "" {
					fmt.Fprintf(w, "  %-20s (not set)\n", s.Key)
				} else {
					fmt.Fprintf(w, "  %-20s %s  ← %s\n", s.Key, s.Value, s.Source)
				}
			}
		}

		// Claude API test
		fmt.Fprintln(w)
		p := progress.New()
		p.Start("Claude API test")
		testConv := "[User]\nHello, this is a test.\n\n[Assistant]\nHi! How can I help?\n"
		testPrompt := feed.BuildPrompt(testConv, "", "en")
		testOut, err := feed.CallClaude(testPrompt, "", "1m")
		if err != nil {
			p.Fail(err)
			issues++
		} else {
			summary, tags := feed.ParseSummaryAndTags(testOut)
			if strings.TrimSpace(summary) == "" {
				p.Fail(fmt.Errorf("empty summary (API format may have changed)"))
				issues++
			} else if len(tags) == 0 {
				p.DoneMsg("Claude API:  WARNING — summary OK but no tags extracted")
			} else {
				p.DoneMsg(fmt.Sprintf("Claude API:  OK (%d tags)", len(tags)))
			}
		}

		// Summary
		fmt.Fprintln(w)
		if issues == 0 && fixed == 0 {
			fmt.Fprintln(w, "All checks passed.")
		} else if issues == 0 && fixed > 0 {
			fmt.Fprintf(w, "Fixed %d issue(s). All checks passed now.\n", fixed)
		} else {
			fmt.Fprintf(w, "%d issue(s) found, %d fixed.\n", issues, fixed)
			fmt.Fprintln(w, "If issues persist, run 'clerk logs --error' and report to GitHub.")
			cmd.SilenceErrors = true
			return fmt.Errorf("%d issue(s) could not be resolved", issues)
		}
		return nil
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

func extractHookBinaryFrom(settings map[string]interface{}) string {
	if settings == nil {
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

func checkHookPathWith(settings map[string]interface{}) string {
	if settings == nil {
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
	rootCmd.AddCommand(registerCmd)
}
