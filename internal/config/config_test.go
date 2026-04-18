package config

import (
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
