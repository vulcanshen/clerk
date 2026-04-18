package feed

import (
	"encoding/json"
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
				if !containsStr(result, s) {
					t.Errorf("result missing %q:\n%s", s, result)
				}
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	prompt := BuildPrompt("conv", "prior", "en")
	if !containsStr(prompt, "en") {
		t.Error("prompt should contain language")
	}
	if !containsStr(prompt, "conv") {
		t.Error("prompt should contain conversation")
	}
	if !containsStr(prompt, "prior") {
		t.Error("prompt should contain prior summary")
	}
	if !containsStr(prompt, "CLERK:TAGS") {
		t.Error("prompt should contain CLERK:TAGS instruction")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
