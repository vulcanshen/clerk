package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type HookEntry struct {
	Hooks []HookCommand `json:"hooks"`
}

type HooksConfig struct {
	SessionEnd []HookEntry `json:"SessionEnd,omitempty"`
}

type ClaudeSettings struct {
	Hooks   HooksConfig            `json:"hooks,omitempty"`
	Unknown map[string]interface{} `json:"-"`
}

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func clerkCommand() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}
	return exe + " feed", nil
}

func isClerkCommand(cmd string) bool {
	return strings.Contains(cmd, "clerk") && strings.Contains(cmd, "feed")
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

func Install() error {
	command, err := clerkCommand()
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

	sessionEnd, _ := hooks["SessionEnd"].([]interface{})

	for _, entry := range sessionEnd {
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
			if isClerkCommand(cmd) {
				fmt.Println("Clerk hook is already installed.")
				fmt.Printf("Settings file: %s\n", settingsPath())
				return nil
			}
		}
	}

	newEntry := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": command,
			},
		},
	}

	sessionEnd = append(sessionEnd, newEntry)
	hooks["SessionEnd"] = sessionEnd
	settings["hooks"] = hooks

	if err := writeSettings(settings); err != nil {
		return err
	}

	fmt.Printf("Installed clerk hook: %s\n", command)
	fmt.Printf("Settings file: %s\n", settingsPath())
	return nil
}

func Uninstall() error {
	settings, err := readSettings()
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		fmt.Println("No hooks found, nothing to uninstall.")
		return nil
	}

	sessionEnd, _ := hooks["SessionEnd"].([]interface{})
	if len(sessionEnd) == 0 {
		fmt.Println("No SessionEnd hooks found, nothing to uninstall.")
		return nil
	}

	var filtered []interface{}
	removed := false

	for _, entry := range sessionEnd {
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
			if isClerkCommand(cmd) {
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

	if !removed {
		fmt.Println("No clerk hook found, nothing to uninstall.")
		return nil
	}

	if len(filtered) == 0 {
		delete(hooks, "SessionEnd")
	} else {
		hooks["SessionEnd"] = filtered
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	if err := writeSettings(settings); err != nil {
		return err
	}

	fmt.Println("Clerk hook uninstalled.")
	fmt.Printf("Settings file: %s\n", settingsPath())
	return nil
}
