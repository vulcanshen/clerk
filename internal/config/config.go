package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OutputConfig struct {
	Dir      string `json:"dir"`
	Language string `json:"language"`
}

type SummaryConfig struct {
	Model   string `json:"model"`
	Timeout string `json:"timeout"`
}

type LogConfig struct {
	RetentionDays int `json:"retention_days"`
}

type FeedConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type Config struct {
	Output  OutputConfig  `json:"output"`
	Summary SummaryConfig `json:"summary"`
	Log     LogConfig     `json:"log"`
	Feed    FeedConfig    `json:"feed"`
}

func boolPtr(b bool) *bool {
	return &b
}

func DefaultConfig() Config {
	return Config{
		Output: OutputConfig{
			Dir:      "~/.clerk/",
			Language: "en",
		},
		Summary: SummaryConfig{
			Model:   "",
			Timeout: "5m",
		},
		Log: LogConfig{
			RetentionDays: 30,
		},
		Feed: FeedConfig{
			Enabled: boolPtr(true),
		},
	}
}

func IsFeedEnabled(cfg Config) bool {
	if cfg.Feed.Enabled == nil {
		return true
	}
	return *cfg.Feed.Enabled
}

func GlobalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "clerk", ".clerk.json")
}

func ProjectConfigPath(cwd string) string {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	return filepath.Join(cwd, ".clerk.json")
}

// ConfigPath returns the global config path (for backward compatibility with config show)
func ConfigPath() string {
	return GlobalConfigPath()
}

func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parsing config %s: %w", path, err)
	}
	return nil
}

// Load merges: defaults → global → project (cwd)
func Load() (Config, error) {
	return LoadWithCwd("")
}

func LoadWithCwd(cwd string) (Config, error) {
	cfg := DefaultConfig()

	if err := loadFile(GlobalConfigPath(), &cfg); err != nil {
		return cfg, err
	}

	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	if cwd != "" {
		if err := loadFile(ProjectConfigPath(cwd), &cfg); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

func ValidKeys() []string {
	return []string{
		"output.dir",
		"output.language",
		"summary.model",
		"summary.timeout",
		"log.retention_days",
		"feed.enabled",
	}
}

func applyKeyValue(cfg *Config, key, value string) error {
	switch key {
	case "output.dir":
		cfg.Output.Dir = value
	case "output.language":
		cfg.Output.Language = value
	case "summary.model":
		cfg.Summary.Model = value
	case "summary.timeout":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid value for summary.timeout: %s (use format like 5m, 2m30s, 1h)", value)
		}
		cfg.Summary.Timeout = value
	case "log.retention_days":
		var days int
		if _, err := fmt.Sscanf(value, "%d", &days); err != nil {
			return fmt.Errorf("invalid value for log.retention_days: %s (must be an integer)", value)
		}
		cfg.Log.RetentionDays = days
	case "feed.enabled":
		switch strings.ToLower(value) {
		case "true", "1":
			cfg.Feed.Enabled = boolPtr(true)
		case "false", "0":
			cfg.Feed.Enabled = boolPtr(false)
		default:
			return fmt.Errorf("invalid value for feed.enabled: %s (must be true or false)", value)
		}
	default:
		return fmt.Errorf("unknown key: %s\nvalid keys: %s", key, strings.Join(ValidKeys(), ", "))
	}
	return nil
}

func Set(key, value string, global bool) error {
	var path string
	if global {
		path = GlobalConfigPath()
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}
		path = ProjectConfigPath(cwd)
	}

	// load existing file as raw map to preserve only set fields
	raw := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &raw)
	}

	// validate key and value
	var tmp Config
	if err := applyKeyValue(&tmp, key, value); err != nil {
		return err
	}

	// set the value in the raw map
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 2 {
		section, _ := raw[parts[0]].(map[string]interface{})
		if section == nil {
			section = make(map[string]interface{})
		}
		switch key {
		case "log.retention_days":
			var days int
			fmt.Sscanf(value, "%d", &days)
			section[parts[1]] = days
		case "feed.enabled":
			section[parts[1]] = strings.ToLower(value) == "true" || value == "1"
		default:
			section[parts[1]] = value
		}
		raw[parts[0]] = section
	}

	return saveRawToPath(path, raw)
}

func saveRawToPath(path string, raw map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

func saveToPath(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

// Save saves to global config (backward compat)
func Save(cfg Config) error {
	return saveToPath(GlobalConfigPath(), cfg)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
