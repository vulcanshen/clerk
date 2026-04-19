package mcpserver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vulcanshen/clerk/internal/config"
)

func testConfig(dir string) config.Config {
	return config.Config{
		Output: config.OutputConfig{
			Dir: dir,
		},
	}
}

// Test 6: Session entry parsing (tab format + old format compatibility)
func TestReadSessionEntries(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(dir)

	sessDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatalf("creating sessions dir: %v", err)
	}

	slug := "test-project"
	sessFile := filepath.Join(sessDir, slug+".md")

	t.Run("tab format with spaces in cwd", func(t *testing.T) {
		// Tab-separated: cwd has spaces, tab separates cwd from transcript path
		content := "- 2026-04-18 14:30:05 `session-abc`\t/Users/Some User/My Project\t/Users/Some User/.claude/projects/abc.jsonl\n"
		if err := os.WriteFile(sessFile, []byte(content), 0644); err != nil {
			t.Fatalf("writing session file: %v", err)
		}

		entries, err := readSessionEntries(cfg, slug)
		if err != nil {
			t.Fatalf("readSessionEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		e := entries[0]
		if e.SessionID != "session-abc" {
			t.Errorf("SessionID = %q, want %q", e.SessionID, "session-abc")
		}
		if e.Cwd != "/Users/Some User/My Project" {
			t.Errorf("Cwd = %q, want %q", e.Cwd, "/Users/Some User/My Project")
		}
		if e.TranscriptPath != "/Users/Some User/.claude/projects/abc.jsonl" {
			t.Errorf("TranscriptPath = %q, want %q", e.TranscriptPath, "/Users/Some User/.claude/projects/abc.jsonl")
		}
	})

	t.Run("old space format", func(t *testing.T) {
		// Old format: space-separated, no tabs. Cwd has no spaces.
		content := "- 2026-04-17 10:00:00 `session-old` /Users/vulcan/project /Users/vulcan/.claude/projects/def.jsonl\n"
		if err := os.WriteFile(sessFile, []byte(content), 0644); err != nil {
			t.Fatalf("writing session file: %v", err)
		}

		entries, err := readSessionEntries(cfg, slug)
		if err != nil {
			t.Fatalf("readSessionEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		e := entries[0]
		if e.SessionID != "session-old" {
			t.Errorf("SessionID = %q, want %q", e.SessionID, "session-old")
		}
		if e.TranscriptPath != "/Users/vulcan/.claude/projects/def.jsonl" {
			t.Errorf("TranscriptPath = %q, want %q", e.TranscriptPath, "/Users/vulcan/.claude/projects/def.jsonl")
		}
	})

	t.Run("old format with spaces in cwd", func(t *testing.T) {
		// Old format with spaces in cwd: split on last space
		content := "- 2026-04-17 10:00:00 `session-sp` /Users/Some User/project /path/to/transcript.jsonl\n"
		if err := os.WriteFile(sessFile, []byte(content), 0644); err != nil {
			t.Fatalf("writing session file: %v", err)
		}

		entries, err := readSessionEntries(cfg, slug)
		if err != nil {
			t.Fatalf("readSessionEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		e := entries[0]
		if e.SessionID != "session-sp" {
			t.Errorf("SessionID = %q, want %q", e.SessionID, "session-sp")
		}
		// LastIndex splits on last space, so cwd gets the path-with-spaces part
		if e.Cwd != "/Users/Some User/project" {
			t.Errorf("Cwd = %q, want %q", e.Cwd, "/Users/Some User/project")
		}
		if e.TranscriptPath != "/path/to/transcript.jsonl" {
			t.Errorf("TranscriptPath = %q, want %q", e.TranscriptPath, "/path/to/transcript.jsonl")
		}
	})

	t.Run("minimal old format no cwd", func(t *testing.T) {
		// Oldest format: only session ID, no cwd or transcript
		content := "- 2026-04-16 09:00:00 `session-minimal`\n"
		if err := os.WriteFile(sessFile, []byte(content), 0644); err != nil {
			t.Fatalf("writing session file: %v", err)
		}

		entries, err := readSessionEntries(cfg, slug)
		if err != nil {
			t.Fatalf("readSessionEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		e := entries[0]
		if e.SessionID != "session-minimal" {
			t.Errorf("SessionID = %q, want %q", e.SessionID, "session-minimal")
		}
		if e.Cwd != "" {
			t.Errorf("Cwd should be empty, got %q", e.Cwd)
		}
		if e.TranscriptPath != "" {
			t.Errorf("TranscriptPath should be empty, got %q", e.TranscriptPath)
		}
	})

	t.Run("mixed formats", func(t *testing.T) {
		// Multiple entries with different formats
		content := "- 2026-04-16 09:00:00 `s1`\n" +
			"- 2026-04-17 10:00:00 `s2` /cwd /transcript.jsonl\n" +
			"- 2026-04-18 11:00:00 `s3`\t/cwd with spaces\t/transcript.jsonl\n"
		if err := os.WriteFile(sessFile, []byte(content), 0644); err != nil {
			t.Fatalf("writing session file: %v", err)
		}

		entries, err := readSessionEntries(cfg, slug)
		if err != nil {
			t.Fatalf("readSessionEntries: %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}
	})
}
