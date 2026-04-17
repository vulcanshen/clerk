package mcpserver

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

func sessionsDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), "sessions")
}

func claudeProjectsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "projects")
}

func cwdToClaudeProjectSlug(cwd string) string {
	return strings.ReplaceAll(cwd, "/", "-")
}

func readSessionIDs(cfg config.Config, slug string) ([]string, error) {
	path := filepath.Join(sessionsDir(cfg), slug+".md")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var ids []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// format: "- 2026-04-16 14:30:05 `session-id`"
		if idx := strings.Index(line, "`"); idx >= 0 {
			rest := line[idx+1:]
			if end := strings.Index(rest, "`"); end >= 0 {
				ids = append(ids, rest[:end])
			}
		}
	}
	return ids, scanner.Err()
}

func findTranscriptPaths(cwd string, sessionIDs []string) []string {
	projectSlug := cwdToClaudeProjectSlug(cwd)
	projectDir := filepath.Join(claudeProjectsDir(), projectSlug)

	seen := make(map[string]bool)
	var paths []string
	for _, id := range sessionIDs {
		p := filepath.Join(projectDir, id+".jsonl")
		if seen[p] {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
			seen[p] = true
		}
	}
	return paths
}

func findSummaryPaths(cfg config.Config, slug string) []string {
	root := filepath.Join(config.ExpandPath(cfg.Output.Dir), "summary")
	var paths []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() || len(e.Name()) != 8 {
			continue
		}
		p := filepath.Join(root, e.Name(), slug+".md")
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}
	return paths
}

func Resume(cwd string, cfg config.Config) (string, error) {
	slug := feed.CwdToSlug(cwd)

	sessionIDs, err := readSessionIDs(cfg, slug)
	if err != nil {
		return "", fmt.Errorf("reading session history: %w", err)
	}

	transcripts := findTranscriptPaths(cwd, sessionIDs)
	summaries := findSummaryPaths(cfg, slug)

	if len(transcripts) == 0 && len(summaries) == 0 {
		return fmt.Sprintf("No previous sessions or summaries found for %s.", slug), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s\n\n", slug))

	if len(summaries) > 0 {
		sb.WriteString("## Summary files (read these first for a quick overview)\n\n")
		for _, p := range summaries {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	if len(transcripts) > 0 {
		sb.WriteString("## Transcript files (full conversation history)\n\n")
		for _, p := range transcripts {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Read the summary files first for context. If you need more detail, read the relevant transcript files.")

	return sb.String(), nil
}
