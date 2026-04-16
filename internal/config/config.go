package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type OutputConfig struct {
	Dir      string `json:"dir"`
	Language string `json:"language"`
}

type SummaryConfig struct {
	Model string `json:"model"`
}

type LogConfig struct {
	RetentionDays int `json:"retention_days"`
}

type Config struct {
	Output  OutputConfig  `json:"output"`
	Summary SummaryConfig `json:"summary"`
	Log     LogConfig     `json:"log"`
}

func DefaultConfig() Config {
	return Config{
		Output: OutputConfig{
			Dir:      "~/.clerk/",
			Language: "zh-TW",
		},
		Summary: SummaryConfig{
			Model: "",
		},
		Log: LogConfig{
			RetentionDays: 30,
		},
	}
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "clerk", "config.json")
}

func Load() (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func ValidKeys() []string {
	return []string{
		"output.dir",
		"output.language",
		"summary.model",
		"log.retention_days",
	}
}

func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "output.dir":
		cfg.Output.Dir = value
	case "output.language":
		cfg.Output.Language = value
	case "summary.model":
		cfg.Summary.Model = value
	case "log.retention_days":
		var days int
		if _, err := fmt.Sscanf(value, "%d", &days); err != nil {
			return fmt.Errorf("invalid value for log.retention_days: %s (must be an integer)", value)
		}
		cfg.Log.RetentionDays = days
	default:
		return fmt.Errorf("unknown key: %s\nvalid keys: %s", key, strings.Join(ValidKeys(), ", "))
	}

	return Save(cfg)
}

func Save(cfg Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
