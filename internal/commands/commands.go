package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func skillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

type skill struct {
	name    string
	content string
}

var clerkSkills = []skill{
	{
		name: "clerk-resume",
		content: `---
name: clerk-resume
description: Recover context from previous Claude Code sessions using clerk MCP tool
disable-model-invocation: true
---

First, check if the clerk MCP server is available by looking for "clerk" in your MCP tools. If the clerk-resume tool is not available, tell the user to run "clerk install mcp" and restart the session.

If the tool is available:
1. Call the clerk-resume MCP tool with cwd set to the current project's absolute working directory
2. Read all summary files first for a quick overview of past work
3. Read the transcript files to rebuild full context from previous conversations
4. Summarize what was previously done and confirm context is restored with the user
`,
	},
	{
		name: "clerk-search",
		content: `---
name: clerk-search
description: Search previous Claude Code sessions by keyword using clerk MCP tool
disable-model-invocation: true
---

First, check if the clerk MCP server is available by looking for "clerk" in your MCP tools. If the clerk-search tool is not available, tell the user to run "clerk install mcp" and restart the session.

If the tool is available:
1. Ask the user what keyword or topic they want to search for (if not already provided as an argument)
2. Call the clerk-search MCP tool with the keyword
3. Read the returned summary and transcript files to understand the context
4. Present the relevant context to the user
`,
	},
}

func Install() error {
	for _, s := range clerkSkills {
		dir := filepath.Join(skillsDir(), s.name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating skill directory %s: %w", s.name, err)
		}
		path := filepath.Join(dir, "SKILL.md")
		if err := os.WriteFile(path, []byte(s.content), 0644); err != nil {
			return fmt.Errorf("writing skill %s: %w", s.name, err)
		}
	}

	fmt.Printf("  skill:   installed (%d skills: %s)\n", len(clerkSkills), skillNames())
	return nil
}

func Uninstall() error {
	removed := 0

	for _, s := range clerkSkills {
		dir := filepath.Join(skillsDir(), s.name)
		if err := os.RemoveAll(dir); err == nil {
			removed++
		}
	}

	if removed == 0 {
		fmt.Println("  skill:   not installed")
	} else {
		fmt.Printf("  skill:   removed (%d skills)\n", removed)
	}
	return nil
}

func skillNames() string {
	names := ""
	for i, s := range clerkSkills {
		if i > 0 {
			names += ", "
		}
		names += "/" + s.name
	}
	return names
}
