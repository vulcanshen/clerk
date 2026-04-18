package mcpserver

import (
	"testing"
)

func TestParseList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"comma separated", "go, cobra, mcp", []string{"go", "cobra", "mcp"}},
		{"space separated", "go cobra mcp", []string{"go", "cobra", "mcp"}},
		{"mixed", "go, cobra mcp", []string{"go", "cobra", "mcp"}},
		{"with extra spaces", "  go ,  cobra  , mcp  ", []string{"go", "cobra", "mcp"}},
		{"single item", "go", []string{"go"}},
		{"empty", "", nil},
		{"uppercase", "Go, COBRA, Mcp", []string{"go", "cobra", "mcp"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseList(tt.input)
			if len(got) != len(tt.expect) {
				t.Errorf("parseList(%q) = %v, want %v", tt.input, got, tt.expect)
				return
			}
			for i, v := range got {
				if v != tt.expect[i] {
					t.Errorf("parseList(%q)[%d] = %q, want %q", tt.input, i, v, tt.expect[i])
				}
			}
		})
	}
}
