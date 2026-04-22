package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/config"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"20260421", "2026-04-21"},
		{"20260101", "2026-01-01"},
		{"invalid", "invalid"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatDate(tt.input)
		if got != tt.expected {
			t.Errorf("formatDate(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildReportPrompt(t *testing.T) {
	prompt := buildReportPrompt("some summaries", "20260420", "20260421")
	if !strings.Contains(prompt, "some summaries") {
		t.Error("prompt should contain summaries")
	}
	if !strings.Contains(prompt, "2026-04-20") {
		t.Error("prompt should contain formatted start date")
	}
	if !strings.Contains(prompt, "2026-04-21") {
		t.Error("prompt should contain formatted end date")
	}
	if strings.Contains(prompt, "Output language:") {
		t.Error("prompt should not contain Output language (moved to system prompt)")
	}
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"with frontmatter",
			"---\ntags: go, cli\n---\n# Summary\nContent here",
			"# Summary\nContent here",
		},
		{
			"no frontmatter",
			"# Just content\nNo frontmatter",
			"# Just content\nNo frontmatter",
		},
		{
			"CRLF frontmatter",
			"---\r\ntags: go\r\n---\r\n# Summary\r\nContent",
			"# Summary\nContent",
		},
		{
			"incomplete frontmatter",
			"---\ntags: go\nno closing",
			"---\ntags: go\nno closing",
		},
		{
			"empty content after frontmatter",
			"---\ntags: go\n---\n",
			"",
		},
	}
	for _, tt := range tests {
		got := stripFrontmatter(tt.input)
		if got != tt.expected {
			t.Errorf("stripFrontmatter(%s) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{5 * time.Minute, "5m0s"},
		{0, "0s"},
		{61 * time.Second, "1m1s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestPad(t *testing.T) {
	// pad should add spaces between header and time string
	result := pad(10, "5m30s")
	if !strings.HasSuffix(result, "5m30s") {
		t.Error("pad should end with time string")
	}
	if !strings.HasPrefix(result, " ") {
		t.Error("pad should start with spaces")
	}

	// empty time string
	if pad(10, "") != "" {
		t.Error("pad with empty time should return empty")
	}

	// very long header should still have minimum 2 spaces
	result = pad(100, "1s")
	if !strings.Contains(result, "  1s") {
		t.Error("pad should have at least 2 spaces")
	}
}

func TestNoFileComp(t *testing.T) {
	completions, directive := noFileComp(nil, nil, "")
	if completions != nil {
		t.Error("noFileComp should return nil completions")
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("noFileComp directive = %d, want %d", directive, cobra.ShellCompDirectiveNoFileComp)
	}
}

func TestAtomicWriteFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")

	err := atomicWriteFile(path, "# Hello\nWorld")
	if err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(data) != "# Hello\nWorld" {
		t.Errorf("file content = %q, want '# Hello\\nWorld'", string(data))
	}

	// Overwrite existing file
	err = atomicWriteFile(path, "updated")
	if err != nil {
		t.Fatalf("atomicWriteFile overwrite failed: %v", err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "updated" {
		t.Errorf("overwritten content = %q, want 'updated'", string(data))
	}
}

func TestDefaultReportPath(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Config{
		Output: config.OutputConfig{Dir: tmp},
	}

	// First path should not have suffix
	path1 := defaultReportPath(cfg, 1)
	if !strings.HasSuffix(path1, ".md") {
		t.Errorf("path should end with .md, got %s", path1)
	}
	if !strings.Contains(path1, "reports") {
		t.Errorf("path should contain 'reports', got %s", path1)
	}

	// Create the file so next call gets incremented name
	os.MkdirAll(filepath.Dir(path1), 0755)
	os.WriteFile(path1, []byte("exists"), 0644)

	path2 := defaultReportPath(cfg, 1)
	if path2 == path1 {
		t.Error("second path should differ from first when first exists")
	}
	if !strings.Contains(path2, "-1.md") {
		t.Errorf("second path should have -1 suffix, got %s", path2)
	}
}

func TestDefaultExportPath(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Config{
		Output: config.OutputConfig{Dir: tmp},
	}

	path1 := defaultExportPath(cfg, "test-slug")
	if !strings.HasSuffix(path1, "test-slug.md") {
		t.Errorf("path should end with test-slug.md, got %s", path1)
	}

	// Create file for collision test
	os.MkdirAll(filepath.Dir(path1), 0755)
	os.WriteFile(path1, []byte("exists"), 0644)

	path2 := defaultExportPath(cfg, "test-slug")
	if !strings.Contains(path2, "test-slug-1.md") {
		t.Errorf("collided path should have -1 suffix, got %s", path2)
	}
}

func TestCopyFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")

	os.WriteFile(src, []byte("hello copy"), 0644)

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "hello copy" {
		t.Errorf("copied content = %q, want 'hello copy'", string(data))
	}
}

func TestCopyDir(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	dstDir := filepath.Join(tmp, "dst")

	// Create source structure
	os.MkdirAll(filepath.Join(srcDir, "sub"), 0755)
	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("file a"), 0644)
	os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("file b"), 0644)

	err := copyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dstDir, "a.txt"))
	if string(data) != "file a" {
		t.Errorf("a.txt = %q, want 'file a'", string(data))
	}
	data, _ = os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
	if string(data) != "file b" {
		t.Errorf("sub/b.txt = %q, want 'file b'", string(data))
	}
}

