package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

func Search(keyword string, cfg config.Config) (string, error) {
	keyword = strings.TrimSpace(strings.ToLower(keyword))
	if keyword == "" {
		return "", fmt.Errorf("keyword is empty")
	}

	dir := feed.TagsDir(cfg)

	// exact match first
	tagFile := filepath.Join(dir, keyword+".md")
	if data, err := os.ReadFile(tagFile); err == nil {
		return fmt.Sprintf("## Tag: %s\n\n%s", keyword, string(data)), nil
	}

	// partial match — scan all tag files
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "No tags found. Run some sessions with clerk feed first.", nil
		}
		return "", err
	}

	var matches []string
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if strings.Contains(name, keyword) {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		// list available tags
		var available []string
		for _, e := range entries {
			available = append(available, strings.TrimSuffix(e.Name(), ".md"))
		}
		if len(available) == 0 {
			return "No tags found. Run some sessions with clerk feed first.", nil
		}
		return fmt.Sprintf("No tags matching '%s'. Available tags: %s", keyword, strings.Join(available, ", ")), nil
	}

	var sb strings.Builder
	for _, tag := range matches {
		data, err := os.ReadFile(filepath.Join(dir, tag+".md"))
		if err != nil {
			continue
		}
		fmt.Fprintf(&sb, "## Tag: %s\n\n%s\n", tag, string(data))
	}

	return sb.String(), nil
}
