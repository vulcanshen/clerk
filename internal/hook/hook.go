package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func clerkExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}
	return exe, nil
}

func isClerkHook(cmd, subcommand string) bool {
	return strings.Contains(cmd, "clerk") && strings.Contains(cmd, subcommand)
}

func readSettings() (map[string]interface{}, error) {
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	return settings, nil
}

func writeSettings(settings map[string]interface{}) error {
	path := settingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

// IsInstalled checks if clerk hooks are currently installed.
func IsInstalled() bool {
	settings, err := readSettings()
	if err != nil {
		return false
	}
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return false
	}
	return hasHook(hooks, "SessionEnd", "feed") || hasHook(hooks, "SessionStart", "punch")
}

func hasHook(hooks map[string]interface{}, event, subcommand string) bool {
	entries, _ := hooks[event].([]interface{})
	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		hooksList, _ := entryMap["hooks"].([]interface{})
		for _, h := range hooksList {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			cmd, _ := hMap["command"].(string)
			if isClerkHook(cmd, subcommand) {
				return true
			}
		}
	}
	return false
}

func addHook(hooks map[string]interface{}, event, command string) {
	entries, _ := hooks[event].([]interface{})
	newEntry := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": command,
			},
		},
	}
	entries = append(entries, newEntry)
	hooks[event] = entries
}

func removeHook(hooks map[string]interface{}, event, subcommand string) bool {
	entries, _ := hooks[event].([]interface{})
	if len(entries) == 0 {
		return false
	}

	var filtered []interface{}
	removed := false

	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			filtered = append(filtered, entry)
			continue
		}

		hooksList, _ := entryMap["hooks"].([]interface{})
		var filteredHooks []interface{}
		for _, h := range hooksList {
			hMap, ok := h.(map[string]interface{})
			if !ok {
				filteredHooks = append(filteredHooks, h)
				continue
			}
			cmd, _ := hMap["command"].(string)
			if isClerkHook(cmd, subcommand) {
				removed = true
				continue
			}
			filteredHooks = append(filteredHooks, h)
		}

		if len(filteredHooks) > 0 {
			entryMap["hooks"] = filteredHooks
			filtered = append(filtered, entryMap)
		}
	}

	if len(filtered) == 0 {
		delete(hooks, event)
	} else {
		hooks[event] = filtered
	}

	return removed
}

func Install() error {
	exe, err := clerkExePath()
	if err != nil {
		return err
	}

	settings, err := readSettings()
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	feedCmd := exe + " feed"
	punchCmd := exe + " punch"

	feedExists := hasHook(hooks, "SessionEnd", "feed")
	punchExists := hasHook(hooks, "SessionStart", "punch")

	if feedExists && punchExists {
		fmt.Println("  hook: already installed")
		return nil
	}

	if !feedExists {
		addHook(hooks, "SessionEnd", feedCmd)
	}
	if !punchExists {
		addHook(hooks, "SessionStart", punchCmd)
	}

	settings["hooks"] = hooks

	if err := writeSettings(settings); err != nil {
		return err
	}

	fmt.Println("  hook: installed (SessionStart + SessionEnd)")
	return nil
}

func Uninstall() error {
	settings, err := readSettings()
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		fmt.Println("  hook: not installed")
		return nil
	}

	feedRemoved := removeHook(hooks, "SessionEnd", "feed")
	punchRemoved := removeHook(hooks, "SessionStart", "punch")

	if !feedRemoved && !punchRemoved {
		fmt.Println("  hook: not installed")
		return nil
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	if err := writeSettings(settings); err != nil {
		return err
	}

	fmt.Println("  hook: removed (SessionStart + SessionEnd)")
	return nil
}
