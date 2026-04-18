package feed

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/logger"
	"github.com/vulcanshen/clerk/internal/platform"
)

type HookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type InnerMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type Message struct {
	Type    string       `json:"type"`
	Role    string       `json:"role"`
	Message InnerMessage `json:"message"`
	Content []ContentBlock `json:"content"`
}

func ParseHookInput(data []byte) (HookInput, error) {
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return input, fmt.Errorf("parsing hook input: %w", err)
	}
	if input.TranscriptPath == "" {
		return input, fmt.Errorf("transcript_path is empty")
	}
	if input.Cwd == "" {
		return input, fmt.Errorf("cwd is empty")
	}
	return input, nil
}

func ReadTranscript(path string, skipLines int) ([]Message, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("opening transcript: %w", err)
	}
	defer f.Close()

	var messages []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	lineNum := 0
	parseErrors := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= skipLines {
			continue
		}
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			parseErrors++
			continue
		}
		messages = append(messages, msg)
	}
	return messages, lineNum, parseErrors, scanner.Err()
}

func cursorDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), "cursor")
}

func cursorPath(cfg config.Config, cwd string) string {
	slug := CwdToSlug(cwd)
	return filepath.Join(cursorDir(cfg), time.Now().Format("20060102")+"-"+slug)
}

func readCursor(cfg config.Config, cwd string) int {
	data, err := os.ReadFile(cursorPath(cfg, cwd))
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return n
}

func writeCursor(cfg config.Config, cwd string, lineCount int) error {
	dir := cursorDir(cfg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cursor directory: %w", err)
	}
	if err := os.WriteFile(cursorPath(cfg, cwd), []byte(strconv.Itoa(lineCount)), 0644); err != nil {
		return fmt.Errorf("writing cursor: %w", err)
	}
	cleanOldCursors(cfg)
	return nil
}

func cleanOldCursors(cfg config.Config) {
	dir := cursorDir(cfg)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -cfg.Log.RetentionDays)
	for _, e := range entries {
		name := e.Name()
		if len(name) < 8 {
			continue
		}
		t, err := time.Parse("20060102", name[:8])
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func FilterConversation(messages []Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		// determine role: check outer type first, then message.role
		role := ""
		switch msg.Type {
		case "user":
			role = "User"
		case "assistant":
			role = "Assistant"
		default:
			if msg.Role == "user" {
				role = "User"
			} else if msg.Role == "assistant" {
				role = "Assistant"
			}
		}
		if role == "" {
			continue
		}

		// extract content blocks from inner message or outer content
		blocks := extractTextBlocks(msg)
		for _, text := range blocks {
			if text != "" {
				fmt.Fprintf(&sb, "[%s]\n%s\n\n", role, text)
			}
		}
	}
	return sb.String()
}

func extractTextBlocks(msg Message) []string {
	var texts []string

	// try inner message.content first (actual Claude Code transcript format)
	if len(msg.Message.Content) > 0 {
		// content could be a string or an array
		var contentStr string
		if err := json.Unmarshal(msg.Message.Content, &contentStr); err == nil {
			if contentStr != "" {
				texts = append(texts, contentStr)
			}
			return texts
		}

		var blocks []ContentBlock
		if err := json.Unmarshal(msg.Message.Content, &blocks); err == nil {
			for _, b := range blocks {
				if b.Type == "text" && b.Text != "" {
					texts = append(texts, b.Text)
				}
			}
			return texts
		}
	}

	// fallback: outer content (original expected format)
	for _, b := range msg.Content {
		if b.Type == "text" && b.Text != "" {
			texts = append(texts, b.Text)
		}
	}
	return texts
}

func CwdToSlug(cwd string) string {
	home, _ := os.UserHomeDir()
	rel := cwd
	// Case-insensitive prefix match for Windows
	if len(cwd) >= len(home) && strings.EqualFold(cwd[:len(home)], home) {
		rel = cwd[len(home):]
	}
	rel = strings.ToLower(rel)
	rel = strings.Trim(rel, "/\\")
	rel = strings.ReplaceAll(rel, "/", "-")
	rel = strings.ReplaceAll(rel, "\\", "-")
	if rel == "" {
		rel = "root"
	}
	return rel
}

func BuildPrompt(conversation, priorSummary, language string) string {
	return fmt.Sprintf(`You are a session summarizer for a Claude Code conversation.

Output language: %s

You receive:
1. A prior summary (may be empty on first run)
2. New conversation messages since that summary

Produce a single MERGED summary that integrates the prior summary with new activity. Do not duplicate — update existing items when they were refined or extended.

Format (use section titles in the output language):

### Core Work
(Major features, architecture changes, significant bug fixes)

### Supporting Work
(Docs, CI/CD, scripts, tests, formatting — more concise)

### Key Decisions & Rationale
- **Decision**: ... → **Rationale**: ...

### User Notes
(Reflections, design philosophy, preferences — only keep what is useful for future collaboration)

### Version Log
(One line each: version — key changes)

Rules:
- Prioritize by impact, not chronology
- Omit routine confirmations — summarize release cadence instead
- Be precise with facts (versions, counts)
- Always record WHY when user makes a deliberate choice
- Record "decided NOT to do X" with rationale — these are often more valuable than what was done
- If a prior item was extended in new messages, update it in-place rather than adding a new bullet
- In Version Log, count versions by listing, not by calculation
- Keep the summary concise — merge and condense older items as the summary grows
- Section titles and all content must be in the specified output language

After the summary, output a tag line in this exact format:
<!-- CLERK:TAGS -->
tag1, tag2, tag3

Tag rules:
- Each tag is a single lowercase keyword or hyphenated-word (e.g. go, cobra, mcp, bug-fix, ci-cd)
- NO spaces, NO sentences, NO descriptions — just the keyword itself
- Valid examples: go, cobra, mcp, refactor, bug-fix, ci-cd, docker, auth
- INVALID examples: "go cli + cobra rewrite", "config hierarchy (global + project)", "bug fixes (transcript format"
- Include: technologies, frameworks, tools, concepts, and actions
- Merge with prior tags — keep all relevant tags, remove obsolete ones
- Keep under 20 tags

---
Prior summary:
%s

New messages:
%s`, language, priorSummary, conversation)
}

func CallClaude(prompt string, model string, timeout string) (string, error) {
	args := []string{"-p"}
	if model != "" {
		args = append(args, "--model", model)
	}

	dur, err := time.ParseDuration(timeout)
	if err != nil {
		dur = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(), "CLERK_INTERNAL=1")

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude -p timed out after %s", timeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude exited with error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("running claude: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func summaryPath(cfg config.Config, cwd string) string {
	dir := config.ExpandPath(cfg.Output.Dir)
	dateDir := filepath.Join(dir, "summary", time.Now().Format("20060102"))
	slug := CwdToSlug(cwd)
	return filepath.Join(dateDir, slug+".md")
}

func ReadExistingSummary(cfg config.Config, cwd string) string {
	data, err := os.ReadFile(summaryPath(cfg, cwd))
	if err != nil {
		return ""
	}
	return string(data)
}

func ParseSummaryAndTags(output string) (string, []string) {
	parts := strings.SplitN(output, "<!-- CLERK:TAGS -->", 2)
	summary := strings.TrimSpace(parts[0])
	if len(parts) < 2 {
		return summary, nil
	}

	tagLine := strings.TrimSpace(parts[1])
	var tags []string
	for _, t := range strings.Split(tagLine, ",") {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		// reject tags with spaces, newlines, or that are too long
		if strings.ContainsAny(t, " \t\n") || len(t) > 30 {
			continue
		}
		tags = append(tags, t)
	}
	return summary, tags
}

func indexDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), "index")
}

func indexMarkdownLink(indexDir, summaryFilePath string) string {
	rel, err := filepath.Rel(indexDir, summaryFilePath)
	if err != nil {
		rel = summaryFilePath
	}
	rel = filepath.ToSlash(rel)
	slug := strings.TrimSuffix(filepath.Base(summaryFilePath), ".md")
	date := filepath.Base(filepath.Dir(summaryFilePath))
	return fmt.Sprintf("[%s+%s](%s)", slug, date, rel)
}

// BuildTerms combines AI-extracted tags with date, slug, and words from slug.
func BuildTerms(tags []string, cwd string, date string) []string {
	slug := CwdToSlug(cwd)
	words := strings.Split(slug, "-")

	seen := make(map[string]bool)
	var terms []string
	add := func(t string) {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" || seen[t] {
			return
		}
		seen[t] = true
		terms = append(terms, t)
	}

	for _, tag := range tags {
		add(tag)
	}
	add(date)
	add(slug)
	for _, w := range words {
		add(w)
	}

	return terms
}

func saveIndex(cfg config.Config, summaryFilePath string, terms []string) error {
	dir := indexDir(cfg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating index directory: %w", err)
	}

	for _, term := range terms {
		if strings.Contains(term, "..") || strings.Contains(term, "/") || strings.Contains(term, "\\") {
			continue
		}
		termFile := filepath.Join(dir, term+".md")

		f, err := os.OpenFile(termFile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			continue
		}

		if err := platform.FlockExclusive(f); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to lock index file %s: %v\n", term, err)
			f.Close()
			continue
		}

		mdLink := indexMarkdownLink(dir, summaryFilePath)

		// read existing content through locked file descriptor
		f.Seek(0, 0)
		existing, _ := io.ReadAll(f)
		lines := strings.Split(string(existing), "\n")
		var cleaned []string
		found := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			// check for stale entries by extracting link path
			if start := strings.Index(trimmed, "]("); start != -1 {
				end := strings.Index(trimmed[start:], ")")
				if end != -1 {
					relPath := trimmed[start+2 : start+end]
					absPath := filepath.Join(dir, relPath)
					if _, err := os.Stat(absPath); err != nil {
						continue // stale entry
					}
				}
			}
			if strings.Contains(trimmed, mdLink) {
				found = true
			}
			cleaned = append(cleaned, line)
		}

		if !found {
			cleaned = append(cleaned, fmt.Sprintf("- %s", mdLink))
		}

		// overwrite with cleaned content
		content := strings.Join(cleaned, "\n") + "\n"
		f.Truncate(0)
		f.Seek(0, 0)
		if _, err := f.WriteString(content); err != nil {
			platform.FlockUnlock(f)
			f.Close()
			continue
		}

		platform.FlockUnlock(f)
		f.Close()
	}
	return nil
}

func SaveSummary(cfg config.Config, cwd string, summary string, tags []string) error {
	path := summaryPath(cfg, cwd)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}
	defer f.Close()

	if err := platform.FlockExclusive(f); err != nil {
		return fmt.Errorf("locking output file: %w", err)
	}
	locked := true
	defer func() {
		if locked {
			platform.FlockUnlock(f)
		}
	}()

	// YAML frontmatter with tags for Obsidian
	var sb strings.Builder
	if len(tags) > 0 {
		sb.WriteString("---\ntags:\n")
		for _, tag := range tags {
			fmt.Fprintf(&sb, "  - %s\n", tag)
		}
		sb.WriteString("---\n\n")
	}

	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(&sb, "# %s\n\n> Last updated: %s\n\n%s\n", CwdToSlug(cwd), timestamp, summary)

	_, err = f.WriteString(sb.String())
	return err
}

func runningDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), "running")
}

func writeRunningState(cfg config.Config, slug, cwd, conversation string) (string, error) {
	dir := runningDir(cfg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	state := RunningState{
		Slug:         slug,
		Cwd:          cwd,
		StartedAt:    time.Now().Format(time.RFC3339),
		Conversation: conversation,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	pid := fmt.Sprintf("%d", os.Getpid())
	path := filepath.Join(dir, pid)
	return path, os.WriteFile(path, data, 0644)
}

func removeRunningState(path string) {
	os.Remove(path)
}

func RunningStates(cfg config.Config) ([]RunningInfo, error) {
	dir := runningDir(cfg)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []RunningInfo
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if !platform.IsProcessAlive(pid) {
			os.Remove(filepath.Join(dir, e.Name()))
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var state RunningState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}
		startedAt, _ := time.Parse(time.RFC3339, state.StartedAt)
		result = append(result, RunningInfo{
			PID:       pid,
			Slug:      state.Slug,
			StartedAt: startedAt,
		})
	}
	return result, nil
}

type RunningState struct {
	Slug         string `json:"slug"`
	Cwd          string `json:"cwd"`
	StartedAt    string `json:"started_at"`
	Conversation string `json:"conversation"`
}

type RunningInfo struct {
	PID       int
	Slug      string
	StartedAt time.Time
}

// OrphanState represents a state file left behind by an interrupted feed process.
type OrphanState struct {
	Path  string
	State RunningState
}

// OrphanStates returns state files whose processes are no longer alive.
func OrphanStates(cfg config.Config) ([]OrphanState, error) {
	dir := runningDir(cfg)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []OrphanState
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if platform.IsProcessAlive(pid) {
			continue
		}

		filePath := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var state RunningState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}
		if state.Conversation == "" {
			os.Remove(filePath)
			continue
		}
		result = append(result, OrphanState{Path: filePath, State: state})
	}
	return result, nil
}

func Run(inputData []byte, cfg config.Config) error {
	logger.Info(cfg, "feed started")

	input, err := ParseHookInput(inputData)
	if err != nil {
		logger.Errorf(cfg, "parse hook input: %v", err)
		return err
	}

	slug := CwdToSlug(input.Cwd)
	logger.Infof(cfg, "processing session for %s (cwd: %s)", slug, input.Cwd)

	cursor := readCursor(cfg, input.Cwd)
	logger.Infof(cfg, "cursor at line %d", cursor)

	messages, totalLines, parseErrors, err := ReadTranscript(input.TranscriptPath, cursor)
	if err != nil {
		logger.Errorf(cfg, "read transcript: %v", err)
		return err
	}
	if parseErrors > 0 {
		logger.Infof(cfg, "skipped %d unparseable lines in transcript", parseErrors)
	}
	logger.Infof(cfg, "read %d new messages (lines %d → %d)", len(messages), cursor, totalLines)

	conversation := FilterConversation(messages)
	if strings.TrimSpace(conversation) == "" {
		logger.Info(cfg, "no new conversation text, skipping")
		if err := writeCursor(cfg, input.Cwd, totalLines); err != nil {
			logger.Errorf(cfg, "write cursor: %v", err)
		}
		return nil
	}

	statePath, err := writeRunningState(cfg, slug, input.Cwd, conversation)
	if err == nil {
		defer removeRunningState(statePath)
	}

	priorSummary := ReadExistingSummary(cfg, input.Cwd)
	if priorSummary != "" {
		logger.Info(cfg, "found prior summary, will merge")
	}

	logger.Info(cfg, "calling claude -p for summary...")
	prompt := BuildPrompt(conversation, priorSummary, cfg.Output.Language)
	output, err := CallClaude(prompt, cfg.Summary.Model, cfg.Summary.Timeout)
	if err != nil {
		logger.Errorf(cfg, "claude -p failed: %v", err)
		return err
	}
	logger.Info(cfg, "summary generated successfully")

	summary, tags := ParseSummaryAndTags(output)
	date := time.Now().Format("20060102")
	terms := BuildTerms(tags, input.Cwd, date)

	if err := SaveSummary(cfg, input.Cwd, summary, terms); err != nil {
		logger.Errorf(cfg, "save summary: %v", err)
		return err
	}

	if len(terms) > 0 {
		sPath := summaryPath(cfg, input.Cwd)
		if err := saveIndex(cfg, sPath, terms); err != nil {
			logger.Errorf(cfg, "save index: %v", err)
		}
		logger.Infof(cfg, "saved %d index terms: %v", len(terms), terms)
	}

	if err := writeCursor(cfg, input.Cwd, totalLines); err != nil {
		logger.Errorf(cfg, "write cursor: %v", err)
	}
	logger.Infof(cfg, "summary saved for %s, cursor updated to line %d", slug, totalLines)
	return nil
}

// Retry re-runs a summary from an orphan state.
func Retry(orphan OrphanState, cfg config.Config) error {
	logger.Infof(cfg, "retrying summary for %s", orphan.State.Slug)

	priorSummary := ReadExistingSummary(cfg, orphan.State.Cwd)
	prompt := BuildPrompt(orphan.State.Conversation, priorSummary, cfg.Output.Language)
	output, err := CallClaude(prompt, cfg.Summary.Model, cfg.Summary.Timeout)
	if err != nil {
		logger.Errorf(cfg, "retry claude -p failed for %s: %v", orphan.State.Slug, err)
		return err
	}

	summary, tags := ParseSummaryAndTags(output)
	date := time.Now().Format("20060102")
	terms := BuildTerms(tags, orphan.State.Cwd, date)

	if err := SaveSummary(cfg, orphan.State.Cwd, summary, terms); err != nil {
		logger.Errorf(cfg, "retry save failed for %s: %v", orphan.State.Slug, err)
		return err
	}

	if len(terms) > 0 {
		sPath := summaryPath(cfg, orphan.State.Cwd)
		saveIndex(cfg, sPath, terms)
	}

	os.Remove(orphan.Path)
	logger.Infof(cfg, "retry succeeded for %s", orphan.State.Slug)
	return nil
}

// IndexDir exports the index directory path for use by MCP server.
func IndexDir(cfg config.Config) string {
	return indexDir(cfg)
}

// RebuildIndexForSummary builds index entries for a single summary file.
func RebuildIndexForSummary(cfg config.Config, summaryFilePath string, terms []string) error {
	return saveIndex(cfg, summaryFilePath, terms)
}
