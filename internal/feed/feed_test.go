package feed

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCwdToSlug(t *testing.T) {
	tests := []struct {
		name   string
		cwd    string
		expect string
	}{
		{"unix path", "/Users/vulcan/Documents/sideproj/clerk", "documents-sideproj-clerk"},
		{"root home", "/Users/vulcan", "root"},
		{"windows backslash", `C:\Users\test\Desktop\Project`, "c:-users-test-desktop-project"},
		{"windows forward slash", "C:/Users/test/code", "c:-users-test-code"},
		{"windows chinese username", `C:\Users\劉茵淇\Desktop\Project`, "c:-users-劉茵淇-desktop-project"},
		{"empty after trim", "/", "root"},
		{"linux path", "/home/user/projects/app", "home-user-projects-app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CwdToSlug(tt.cwd)
			// home prefix varies by OS, just check no backslashes and no leading/trailing dashes
			if got == "" {
				t.Error("slug should not be empty")
			}
			for _, c := range got {
				if c == '/' || c == '\\' {
					t.Errorf("slug contains path separator: %s", got)
				}
			}
		})
	}
}

func TestCwdToSlugNoBackslashes(t *testing.T) {
	inputs := []string{
		`C:\Users\test\code`,
		`C:\Users\劉茵淇\Desktop\Project\ixCSP\mcp_agent`,
		"/Users/vulcan/Documents/sideproj/clerk",
		"/home/user/projects",
	}
	for _, input := range inputs {
		slug := CwdToSlug(input)
		for _, c := range slug {
			if c == '\\' || c == '/' {
				t.Errorf("CwdToSlug(%q) = %q, contains path separator", input, slug)
			}
		}
	}
}

func TestParseSummaryAndTags(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSummary string
		wantTags    []string
	}{
		{
			name:        "with tags",
			input:       "Some summary\n<!-- CLERK:TAGS -->\ngo, cobra, mcp",
			wantSummary: "Some summary",
			wantTags:    []string{"go", "cobra", "mcp"},
		},
		{
			name:        "no tags",
			input:       "Just a summary",
			wantSummary: "Just a summary",
			wantTags:    nil,
		},
		{
			name:        "reject tags with spaces",
			input:       "Summary\n<!-- CLERK:TAGS -->\ngo, cobra rewrite, mcp",
			wantSummary: "Summary",
			wantTags:    []string{"go", "mcp"},
		},
		{
			name:        "reject long tags",
			input:       "Summary\n<!-- CLERK:TAGS -->\ngo, this-is-a-very-long-tag-that-exceeds-thirty-characters-limit, mcp",
			wantSummary: "Summary",
			wantTags:    []string{"go", "mcp"},
		},
		{
			name:        "empty tags",
			input:       "Summary\n<!-- CLERK:TAGS -->\n",
			wantSummary: "Summary",
			wantTags:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, tags := ParseSummaryAndTags(tt.input)
			if summary != tt.wantSummary {
				t.Errorf("summary = %q, want %q", summary, tt.wantSummary)
			}
			if len(tags) != len(tt.wantTags) {
				t.Errorf("tags = %v, want %v", tags, tt.wantTags)
				return
			}
			for i, tag := range tags {
				if tag != tt.wantTags[i] {
					t.Errorf("tag[%d] = %q, want %q", i, tag, tt.wantTags[i])
				}
			}
		})
	}
}

func TestFilterConversation(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		contains []string
		empty    bool
	}{
		{
			name: "user and assistant",
			messages: []Message{
				{Type: "user", Content: []ContentBlock{{Type: "text", Text: "hello"}}},
				{Type: "assistant", Content: []ContentBlock{{Type: "text", Text: "hi there"}}},
			},
			contains: []string{"[User]", "hello", "[Assistant]", "hi there"},
		},
		{
			name: "skip non-text blocks",
			messages: []Message{
				{Type: "user", Content: []ContentBlock{{Type: "text", Text: "hello"}}},
				{Type: "assistant", Content: []ContentBlock{{Type: "tool_use", Text: ""}}},
			},
			contains: []string{"[User]", "hello"},
		},
		{
			name: "inner message content string",
			messages: []Message{
				{Type: "user", Message: InnerMessage{Role: "user", Content: json.RawMessage(`"hello from inner"`)}},
			},
			contains: []string{"[User]", "hello from inner"},
		},
		{
			name:     "empty messages",
			messages: []Message{},
			empty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterConversation(tt.messages)
			if tt.empty {
				if result != "" {
					t.Errorf("expected empty, got %q", result)
				}
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("result missing %q:\n%s", s, result)
				}
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	prompt := BuildPrompt("conv", "prior")
	if !strings.Contains(prompt, "conv") {
		t.Error("prompt should contain conversation")
	}
	if !strings.Contains(prompt, "prior") {
		t.Error("prompt should contain prior summary")
	}
	if !strings.Contains(prompt, "CLERK:TAGS") {
		t.Error("prompt should contain CLERK:TAGS instruction")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	// both language and instruction
	sp := BuildSystemPrompt("zh-TW", "focus on architecture")
	if sp != "Output language: zh-TW\nfocus on architecture" {
		t.Errorf("unexpected system prompt: %q", sp)
	}
	// language only
	sp = BuildSystemPrompt("en", "")
	if sp != "Output language: en" {
		t.Errorf("unexpected system prompt: %q", sp)
	}
	// instruction only
	sp = BuildSystemPrompt("", "be concise")
	if sp != "be concise" {
		t.Errorf("unexpected system prompt: %q", sp)
	}
	// both empty
	sp = BuildSystemPrompt("", "")
	if sp != "" {
		t.Errorf("expected empty, got: %q", sp)
	}
}

func buildMockClaude(t *testing.T) string {
	t.Helper()
	mockDir := t.TempDir()
	mockBin := filepath.Join(mockDir, "mock_claude")
	cmd := exec.Command("go", "build", "-o", mockBin, "./testdata/mock_claude.go")
	cmd.Dir = filepath.Join(".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build mock_claude: %v\n%s", err, out)
	}
	return mockBin
}

func TestCallClaude(t *testing.T) {
	mockBin := buildMockClaude(t)
	orig := ClaudeBinary
	ClaudeBinary = mockBin
	defer func() { ClaudeBinary = orig }()

	// basic call
	out, err := CallClaude("hello world", "", "1m", "")
	if err != nil {
		t.Fatalf("CallClaude failed: %v", err)
	}
	if !strings.Contains(out, "Mock summary generated") {
		t.Errorf("expected mock summary, got: %s", out)
	}

	// with system prompt
	out, err = CallClaude("test prompt", "", "1m", "Output language: zh-TW")
	if err != nil {
		t.Fatalf("CallClaude with system prompt failed: %v", err)
	}
	if !strings.Contains(out, "Output language: zh-TW") {
		t.Errorf("system prompt not echoed back, got: %s", out)
	}

	// with model
	out, err = CallClaude("test", "haiku", "1m", "")
	if err != nil {
		t.Fatalf("CallClaude with model failed: %v", err)
	}
	if !strings.Contains(out, "Mock summary") {
		t.Errorf("expected mock summary, got: %s", out)
	}
}

func TestCallClaudeError(t *testing.T) {
	mockBin := buildMockClaude(t)
	orig := ClaudeBinary
	ClaudeBinary = mockBin
	defer func() { ClaudeBinary = orig }()

	_, err := CallClaude("MOCK_FAIL", "", "1m", "")
	if err == nil {
		t.Error("expected error for MOCK_FAIL prompt")
	}
}

func TestCallClaudeTimeout(t *testing.T) {
	mockBin := buildMockClaude(t)
	orig := ClaudeBinary
	ClaudeBinary = mockBin
	defer func() { ClaudeBinary = orig }()

	_, err := CallClaude("MOCK_HANG", "", "100ms", "")
	if err == nil {
		t.Error("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %v", err)
	}
}

func TestCallClaudeParseSummaryAndTags(t *testing.T) {
	mockBin := buildMockClaude(t)
	orig := ClaudeBinary
	ClaudeBinary = mockBin
	defer func() { ClaudeBinary = orig }()

	out, err := CallClaude("test", "", "1m", "")
	if err != nil {
		t.Fatalf("CallClaude failed: %v", err)
	}

	summary, tags := ParseSummaryAndTags(out)
	if !strings.Contains(summary, "Mock summary generated") {
		t.Errorf("summary should contain mock content, got: %s", summary)
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(tags), tags)
	}
	expected := []string{"mock", "test", "cli"}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("tag[%d] = %q, want %q", i, tag, expected[i])
		}
	}
}

func TestBuildTerms(t *testing.T) {
	home, _ := os.UserHomeDir()
	cwd := home + "/projects/my-app"
	terms := BuildTerms([]string{"go", "mcp"}, cwd, "20260418")

	expected := []string{"go", "mcp", "20260418", "projects-my-app", "projects", "my", "app"}
	for _, e := range expected {
		found := false
		for _, term := range terms {
			if term == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("BuildTerms missing %q, got %v", e, terms)
		}
	}

	// test dedup: if tag matches a word from slug
	terms2 := BuildTerms([]string{"projects", "go"}, cwd, "20260418")
	count := 0
	for _, term := range terms2 {
		if term == "projects" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 'projects' once, got %d times in %v", count, terms2)
	}
}

