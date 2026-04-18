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

// ListIndex returns all available index term names.
func ListIndex(cfg config.Config) (string, error) {
	dir := feed.IndexDir(cfg)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "No index entries found. Run some sessions with clerk feed first.", nil
		}
		return "", err
	}

	var terms []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		terms = append(terms, strings.TrimSuffix(e.Name(), ".md"))
	}

	if len(terms) == 0 {
		return "No index entries found. Run some sessions with clerk feed first.", nil
	}

	return strings.Join(terms, ", "), nil
}

// ReadIndex returns the content of one or more index term files.
func ReadIndex(terms string, cfg config.Config) (string, error) {
	dir := feed.IndexDir(cfg)

	parsed := parseList(terms)
	if len(parsed) == 0 {
		return "", fmt.Errorf("terms is empty")
	}

	var sb strings.Builder
	for _, term := range parsed {
		if strings.Contains(term, "..") || strings.Contains(term, "/") || strings.Contains(term, "\\") {
			fmt.Fprintf(&sb, "## %s\n\n(invalid term name)\n\n", term)
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, term+".md"))
		if err != nil {
			fmt.Fprintf(&sb, "## %s\n\n(not found)\n\n", term)
			continue
		}
		fmt.Fprintf(&sb, "## %s\n\n%s\n\n", term, strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(sb.String()), nil
}
