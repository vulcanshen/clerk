package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		tilde bool // whether result should NOT contain ~/
	}{
		{"tilde prefix", "~/.clerk/", true},
		{"absolute path", "/usr/local/bin", false},
		{"relative path", "data/", false},
		{"windows path", `C:\Users\test`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if tt.tilde && len(result) > 1 && result[:2] == "~/" {
				t.Errorf("ExpandPath(%q) = %q, still contains ~/", tt.input, result)
			}
			if result == "" {
				t.Error("result should not be empty")
			}
		})
	}
}

func TestApplyKeyValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"output.dir", "output.dir", "~/test/", false},
		{"output.language", "output.language", "zh-TW", false},
		{"summary.model", "summary.model", "haiku", false},
		{"log.retention_days valid", "log.retention_days", "30", false},
		{"log.retention_days invalid", "log.retention_days", "abc", true},
		{"feed.enabled true", "feed.enabled", "true", false},
		{"feed.enabled false", "feed.enabled", "false", false},
		{"feed.enabled invalid", "feed.enabled", "maybe", true},
		{"unknown key", "unknown.key", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := applyKeyValue(cfg, tt.key, tt.value)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for key=%q value=%q", tt.key, tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestApplyKeyValueResults(t *testing.T) {
	cfg := &Config{}

	applyKeyValue(cfg, "output.dir", "~/custom/")
	if cfg.Output.Dir != "~/custom/" {
		t.Errorf("output.dir = %q, want ~/custom/", cfg.Output.Dir)
	}

	applyKeyValue(cfg, "output.language", "ja")
	if cfg.Output.Language != "ja" {
		t.Errorf("output.language = %q, want ja", cfg.Output.Language)
	}

	applyKeyValue(cfg, "log.retention_days", "7")
	if cfg.Log.RetentionDays != 7 {
		t.Errorf("log.retention_days = %d, want 7", cfg.Log.RetentionDays)
	}

	applyKeyValue(cfg, "feed.enabled", "false")
	if cfg.Feed.Enabled == nil || *cfg.Feed.Enabled != false {
		t.Error("feed.enabled should be false")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Output.Dir != "~/.clerk/" {
		t.Errorf("default output.dir = %q, want ~/.clerk/", cfg.Output.Dir)
	}
	if cfg.Output.Language != "en" {
		t.Errorf("default output.language = %q, want en", cfg.Output.Language)
	}
	if cfg.Log.RetentionDays != 30 {
		t.Errorf("default retention_days = %d, want 30", cfg.Log.RetentionDays)
	}
	if !IsFeedEnabled(cfg) {
		t.Error("feed should be enabled by default")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, ".clerk.json")

	// Write a config file
	raw := map[string]interface{}{
		"output": map[string]interface{}{
			"dir":      "~/custom/",
			"language": "ja",
		},
		"summary": map[string]interface{}{
			"model":   "haiku",
			"timeout": "3m",
		},
	}
	data, _ := json.MarshalIndent(raw, "", "  ")
	os.WriteFile(cfgPath, data, 0644)

	// Load it
	cfg := DefaultConfig()
	err := loadFile(cfgPath, &cfg)
	if err != nil {
		t.Fatalf("loadFile failed: %v", err)
	}

	if cfg.Output.Dir != "~/custom/" {
		t.Errorf("output.dir = %q, want ~/custom/", cfg.Output.Dir)
	}
	if cfg.Output.Language != "ja" {
		t.Errorf("output.language = %q, want ja", cfg.Output.Language)
	}
	if cfg.Summary.Model != "haiku" {
		t.Errorf("summary.model = %q, want haiku", cfg.Summary.Model)
	}
	if cfg.Summary.Timeout != "3m" {
		t.Errorf("summary.timeout = %q, want 3m", cfg.Summary.Timeout)
	}
	// Defaults should still be in place for unset fields
	if cfg.Log.RetentionDays != 30 {
		t.Errorf("log.retention_days = %d, want 30 (default)", cfg.Log.RetentionDays)
	}
}

func TestSaveRawToPath(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "subdir", ".clerk.json")

	raw := map[string]interface{}{
		"output": map[string]interface{}{
			"dir": "~/test/",
		},
	}

	err := saveRawToPath(cfgPath, raw)
	if err != nil {
		t.Fatalf("saveRawToPath failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading saved config: %v", err)
	}

	var loaded map[string]interface{}
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parsing saved config: %v", err)
	}

	output, _ := loaded["output"].(map[string]interface{})
	if output["dir"] != "~/test/" {
		t.Errorf("saved output.dir = %v, want ~/test/", output["dir"])
	}
}

func TestLoadWithCwd(t *testing.T) {
	tmp := t.TempDir()

	// Create a project config
	projCfg := map[string]interface{}{
		"output": map[string]interface{}{
			"language": "ko",
		},
		"feed": map[string]interface{}{
			"enabled": false,
		},
	}
	data, _ := json.MarshalIndent(projCfg, "", "  ")
	os.WriteFile(filepath.Join(tmp, ".clerk.json"), data, 0644)

	cfg, err := LoadWithCwd(tmp)
	if err != nil {
		t.Fatalf("LoadWithCwd failed: %v", err)
	}

	if cfg.Output.Language != "ko" {
		t.Errorf("output.language = %q, want ko", cfg.Output.Language)
	}
	if IsFeedEnabled(cfg) {
		t.Error("feed should be disabled")
	}
	// Defaults should remain for unset
	if cfg.Output.Dir != "~/.clerk/" {
		t.Errorf("output.dir = %q, want ~/.clerk/ (default)", cfg.Output.Dir)
	}
}

func TestLoadFileNotExist(t *testing.T) {
	cfg := DefaultConfig()
	err := loadFile("/nonexistent/path/.clerk.json", &cfg)
	if err != nil {
		t.Error("loadFile should return nil for nonexistent file")
	}
	// Config should remain default
	if cfg.Output.Dir != "~/.clerk/" {
		t.Errorf("config should remain default, got dir=%q", cfg.Output.Dir)
	}
}

func TestLoadFileInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".clerk.json")
	os.WriteFile(path, []byte("not json"), 0644)

	cfg := DefaultConfig()
	err := loadFile(path, &cfg)
	if err == nil {
		t.Error("loadFile should return error for invalid JSON")
	}
}

func TestValidKeys(t *testing.T) {
	keys := ValidKeys()
	if len(keys) == 0 {
		t.Error("ValidKeys should not be empty")
	}

	expected := []string{"output.dir", "output.language", "summary.model", "summary.timeout", "log.retention_days", "feed.enabled"}
	for _, e := range expected {
		found := false
		for _, k := range keys {
			if k == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidKeys missing %q", e)
		}
	}
}
