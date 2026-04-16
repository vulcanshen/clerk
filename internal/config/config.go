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

type Config struct {
	Output  OutputConfig  `json:"output"`
	Summary SummaryConfig `json:"summary"`
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

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
