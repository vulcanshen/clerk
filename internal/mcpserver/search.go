package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

// parseList splits a string by commas and/or spaces into trimmed, lowercased tokens.
func parseList(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	var out []string
	for _, tok := range strings.Fields(s) {
		tok = strings.TrimSpace(strings.ToLower(tok))
		if tok != "" {
			out = append(out, tok)
		}
	}
	return out
}

// ListTags returns all available tag names.
func ListTags(cfg config.Config) (string, error) {
	dir := feed.TagsDir(cfg)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "No tags found. Run some sessions with clerk feed first.", nil
		}
		return "", err
	}

	var tags []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		tags = append(tags, strings.TrimSuffix(e.Name(), ".md"))
	}

	if len(tags) == 0 {
		return "No tags found. Run some sessions with clerk feed first.", nil
	}

	return strings.Join(tags, ", "), nil
}

// ReadTags returns the content of one or more tag files.
func ReadTags(tags string, cfg config.Config) (string, error) {
	dir := feed.TagsDir(cfg)

	parsed := parseList(tags)
	if len(parsed) == 0 {
		return "", fmt.Errorf("tags is empty")
	}

	var sb strings.Builder
	for _, tag := range parsed {
		// Sanitize: reject tags with path traversal
		if strings.Contains(tag, "..") || strings.Contains(tag, "/") || strings.Contains(tag, "\\") {
			fmt.Fprintf(&sb, "## Tag: %s\n\n(invalid tag name)\n\n", tag)
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, tag+".md"))
		if err != nil {
			fmt.Fprintf(&sb, "## Tag: %s\n\n(not found)\n\n", tag)
			continue
		}
		fmt.Fprintf(&sb, "## Tag: %s\n\n%s\n\n", tag, strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(sb.String()), nil
}
