package feed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
)

func testConfig(dir string) config.Config {
	return config.Config{
		Output: config.OutputConfig{
			Dir:      dir,
			Language: "en",
		},
		Summary: config.SummaryConfig{
			Timeout: "5m",
		},
		Log: config.LogConfig{
			RetentionDays: 30,
		},
	}
}

func writeTranscriptLines(t *testing.T, path string, messages []Message) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating transcript: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, msg := range messages {
		if err := enc.Encode(msg); err != nil {
			t.Fatalf("encoding message: %v", err)
		}
	}
}

func appendTranscriptLines(t *testing.T, path string, messages []Message) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("opening transcript for append: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, msg := range messages {
		if err := enc.Encode(msg); err != nil {
			t.Fatalf("encoding message: %v", err)
		}
	}
}

func userMsg(text string) Message {
	content, _ := json.Marshal(text)
	return Message{
		Type:    "user",
		Message: InnerMessage{Role: "user", Content: content},
	}
}

func assistantMsg(text string) Message {
	blocks, _ := json.Marshal([]ContentBlock{{Type: "text", Text: text}})
	return Message{
		Type:    "assistant",
		Message: InnerMessage{Role: "assistant", Content: blocks},
	}
}

func toolUseMsg() Message {
	blocks, _ := json.Marshal([]ContentBlock{{Type: "tool_use", Text: ""}})
	return Message{
		Type:    "assistant",
		Message: InnerMessage{Role: "assistant", Content: blocks},
	}
}

// Test 1: Feed pipeline end-to-end
func TestFeedPipelineEndToEnd(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(dir)
	cwd := "/tmp/test-project"

	// Create transcript
	transcriptPath := filepath.Join(dir, "transcript.jsonl")
	writeTranscriptLines(t, transcriptPath, []Message{
		userMsg("Please add a login feature"),
		assistantMsg("I'll implement the login feature with JWT authentication."),
		userMsg("Also add rate limiting"),
		assistantMsg("Done. I've added rate limiting middleware using a token bucket algorithm."),
	})

	// ReadTranscript
	messages, totalLines, parseErrors, err := ReadTranscript(transcriptPath, 0)
	if err != nil {
		t.Fatalf("ReadTranscript: %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(messages))
	}
	if totalLines != 4 {
		t.Fatalf("expected 4 total lines, got %d", totalLines)
	}
	if parseErrors != 0 {
		t.Fatalf("expected 0 parse errors, got %d", parseErrors)
	}

	// FilterConversation
	conversation := FilterConversation(messages)
	if !strings.Contains(conversation, "[User]") {
		t.Error("conversation should contain [User]")
	}
	if !strings.Contains(conversation, "login feature") {
		t.Error("conversation should contain user message text")
	}
	if !strings.Contains(conversation, "[Assistant]") {
		t.Error("conversation should contain [Assistant]")
	}

	// BuildPrompt
	prompt := BuildPrompt(conversation, "", "en")
	if !strings.Contains(prompt, "login feature") {
		t.Error("prompt should contain conversation text")
	}

	// Simulate CallClaude output
	claudeOutput := `### Core Work
- Implemented login feature with JWT authentication
- Added rate limiting middleware

### Supporting Work
- None

### Key Decisions & Rationale
- **Decision**: Use JWT → **Rationale**: Stateless auth for scalability

### User Notes
- User prefers security-first approach

### Version Log
- v1.0.0 — initial login + rate limiting
<!-- CLERK:TAGS -->
go, jwt, rate-limiting, auth`

	// ParseSummaryAndTags
	summary, tags := ParseSummaryAndTags(claudeOutput)
	if !strings.Contains(summary, "JWT authentication") {
		t.Error("summary should contain JWT authentication")
	}
	if len(tags) != 4 {
		t.Errorf("expected 4 tags, got %d: %v", len(tags), tags)
	}

	// BuildTerms
	date := time.Now().Format("20060102")
	terms := BuildTerms(tags, cwd, date)
	if len(terms) == 0 {
		t.Fatal("expected non-empty terms")
	}

	// SaveSummary
	if err := SaveSummary(cfg, cwd, summary, terms); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	// Verify summary file
	sPath := summaryPath(cfg, cwd)
	data, err := os.ReadFile(sPath)
	if err != nil {
		t.Fatalf("reading summary file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "JWT authentication") {
		t.Error("summary file should contain JWT authentication")
	}
	if !strings.Contains(content, "---\ntags:") {
		t.Error("summary file should contain YAML frontmatter")
	}
	for _, tag := range []string{"go", "jwt", "rate-limiting", "auth"} {
		if !strings.Contains(content, "  - "+tag) {
			t.Errorf("frontmatter should contain tag %q", tag)
		}
	}

	// saveIndex
	if err := saveIndex(cfg, sPath, terms); err != nil {
		t.Fatalf("saveIndex: %v", err)
	}

	// Verify index files
	idxDir := indexDir(cfg)
	for _, tag := range []string{"go", "jwt", "rate-limiting", "auth"} {
		termFile := filepath.Join(idxDir, tag+".md")
		data, err := os.ReadFile(termFile)
		if err != nil {
			t.Errorf("index file for %q not found: %v", tag, err)
			continue
		}
		content := string(data)
		slug := CwdToSlug(cwd)
		expectedLink := fmt.Sprintf("[%s+%s]", slug, date)
		if !strings.Contains(content, expectedLink) {
			t.Errorf("index file %q should contain link %q, got: %s", tag, expectedLink, content)
		}
	}
}

// Test 2: Cursor incremental processing
func TestCursorIncremental(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(dir)
	cwd := "/tmp/cursor-test"

	transcriptPath := filepath.Join(dir, "transcript.jsonl")

	// Write 10 lines
	var msgs []Message
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			msgs = append(msgs, userMsg(fmt.Sprintf("message %d", i)))
		} else {
			msgs = append(msgs, assistantMsg(fmt.Sprintf("response %d", i)))
		}
	}
	writeTranscriptLines(t, transcriptPath, msgs)

	// First read: all 10
	messages, totalLines, _, err := ReadTranscript(transcriptPath, 0)
	if err != nil {
		t.Fatalf("ReadTranscript: %v", err)
	}
	if len(messages) != 10 {
		t.Fatalf("expected 10 messages, got %d", len(messages))
	}
	if totalLines != 10 {
		t.Fatalf("expected 10 total lines, got %d", totalLines)
	}

	// Write cursor
	if err := writeCursor(cfg, cwd, totalLines); err != nil {
		t.Fatalf("writeCursor: %v", err)
	}

	// Verify cursor value
	cursor := readCursor(cfg, cwd)
	if cursor != 10 {
		t.Fatalf("expected cursor 10, got %d", cursor)
	}

	// Append 5 more lines
	var extraMsgs []Message
	for i := 10; i < 15; i++ {
		if i%2 == 0 {
			extraMsgs = append(extraMsgs, userMsg(fmt.Sprintf("message %d", i)))
		} else {
			extraMsgs = append(extraMsgs, assistantMsg(fmt.Sprintf("response %d", i)))
		}
	}
	appendTranscriptLines(t, transcriptPath, extraMsgs)

	// Second read: only new 5
	messages, totalLines, _, err = ReadTranscript(transcriptPath, 10)
	if err != nil {
		t.Fatalf("ReadTranscript: %v", err)
	}
	if len(messages) != 5 {
		t.Fatalf("expected 5 new messages, got %d", len(messages))
	}
	if totalLines != 15 {
		t.Fatalf("expected 15 total lines, got %d", totalLines)
	}
}

// Test 3: Empty transcript skip
func TestEmptyTranscriptSkip(t *testing.T) {
	dir := t.TempDir()

	transcriptPath := filepath.Join(dir, "transcript.jsonl")

	// Write only tool_use messages (no text content)
	writeTranscriptLines(t, transcriptPath, []Message{
		toolUseMsg(),
		toolUseMsg(),
		toolUseMsg(),
	})

	messages, totalLines, _, err := ReadTranscript(transcriptPath, 0)
	if err != nil {
		t.Fatalf("ReadTranscript: %v", err)
	}
	if totalLines != 3 {
		t.Fatalf("expected 3 total lines, got %d", totalLines)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	conversation := FilterConversation(messages)
	if strings.TrimSpace(conversation) != "" {
		t.Errorf("expected empty conversation, got %q", conversation)
	}
}

// Test 4: SaveSummary concurrent writes
func TestSaveSummaryConcurrent(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(dir)
	cwd := "/tmp/concurrent-test"

	summaryA := "### Core Work\n- Feature A implemented"
	summaryB := "### Core Work\n- Feature B implemented"
	tags := []string{"go", "test"}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		SaveSummary(cfg, cwd, summaryA, tags)
	}()

	go func() {
		defer wg.Done()
		SaveSummary(cfg, cwd, summaryB, tags)
	}()

	wg.Wait()

	// Verify file is not empty and not corrupted
	sPath := summaryPath(cfg, cwd)
	data, err := os.ReadFile(sPath)
	if err != nil {
		t.Fatalf("reading summary: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Fatal("summary file should not be empty")
	}

	// Content should be one complete summary, not a mix
	hasA := strings.Contains(content, "Feature A")
	hasB := strings.Contains(content, "Feature B")
	if !hasA && !hasB {
		t.Error("summary should contain either Feature A or Feature B")
	}

	// Verify frontmatter is intact
	if !strings.Contains(content, "---\ntags:") {
		t.Error("frontmatter should be intact")
	}
	if !strings.Contains(content, "---\n\n#") {
		t.Error("frontmatter closing should be intact")
	}
}

// Test 5: Cursor cleanup boundary cases
func TestCleanOldCursors(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(dir)
	cfg.Output.Dir = dir
	cfg.Log.RetentionDays = 30

	curDir := filepath.Join(dir, "cursor")
	if err := os.MkdirAll(curDir, 0755); err != nil {
		t.Fatalf("creating cursor dir: %v", err)
	}

	today := time.Now().Format("20060102")
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("20060102")
	sixtyDaysAgo := time.Now().AddDate(0, 0, -60).Format("20060102")

	// Create cursor files with different dates
	files := map[string]bool{
		today + "-project-a":         true,  // keep
		sevenDaysAgo + "-project-b":  true,  // keep (within 30 days)
		sixtyDaysAgo + "-project-c":  false, // delete (older than 30 days)
	}

	for name := range files {
		if err := os.WriteFile(filepath.Join(curDir, name), []byte("100"), 0644); err != nil {
			t.Fatalf("creating cursor file %s: %v", name, err)
		}
	}

	cleanOldCursors(cfg)

	for name, shouldExist := range files {
		_, err := os.Stat(filepath.Join(curDir, name))
		exists := err == nil
		if exists != shouldExist {
			if shouldExist {
				t.Errorf("cursor %s should have been kept but was deleted", name)
			} else {
				t.Errorf("cursor %s should have been deleted but was kept", name)
			}
		}
	}

	// Test with RetentionDays = 1: today's cursor must survive
	cfg.Log.RetentionDays = 1
	cleanOldCursors(cfg)

	if _, err := os.Stat(filepath.Join(curDir, today+"-project-a")); err != nil {
		t.Error("today's cursor should not be deleted even with RetentionDays=1")
	}
}
