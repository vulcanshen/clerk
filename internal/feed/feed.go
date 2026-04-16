package feed

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
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

type Message struct {
	Type    string         `json:"type"`
	Role    string         `json:"role"`
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

func ReadTranscript(path string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening transcript: %w", err)
	}
	defer f.Close()

	var messages []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, scanner.Err()
}

func FilterConversation(messages []Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		for _, block := range msg.Content {
			if block.Type != "text" || block.Text == "" {
				continue
			}
			role := "User"
			if msg.Role == "assistant" {
				role = "Assistant"
			}
			fmt.Fprintf(&sb, "[%s]\n%s\n\n", role, block.Text)
		}
	}
	return sb.String()
}

func CwdToSlug(cwd string) string {
	home, _ := os.UserHomeDir()
	rel := cwd
	if strings.HasPrefix(cwd, home) {
		rel = cwd[len(home):]
	}
	rel = strings.ToLower(rel)
	rel = strings.Trim(rel, "/")
	rel = strings.ReplaceAll(rel, "/", "-")
	if rel == "" {
		rel = "root"
	}
	return rel
}

func BuildPrompt(conversation string, language string) string {
	return fmt.Sprintf(`You are a session summarizer. Summarize the following Claude Code conversation.

Output language: %s

Format your summary as:
## 使用者輸入摘要
(Summarize what the user asked or requested, in bullet points)

## AI 回應摘要
(Summarize what the AI did or responded, in bullet points)

---
Conversation:
%s`, language, conversation)
}

func CallClaude(prompt string, model string) (string, error) {
	args := []string{"-p"}
	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(), "CLERK_INTERNAL=1")

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude exited with error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("running claude: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func SaveSummary(cfg config.Config, cwd string, summary string) error {
	dir := config.ExpandPath(cfg.Output.Dir)
	dateDir := filepath.Join(dir, time.Now().Format("20060102"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	slug := CwdToSlug(cwd)
	filePath := filepath.Join(dateDir, slug+".md")

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}
	defer f.Close()

	timestamp := time.Now().Format("15:04:05")
	_, err = fmt.Fprintf(f, "\n---\n### %s\n\n%s\n", timestamp, summary)
	return err
}

func Run(inputData []byte, cfg config.Config) error {
	input, err := ParseHookInput(inputData)
	if err != nil {
		return err
	}

	messages, err := ReadTranscript(input.TranscriptPath)
	if err != nil {
		return err
	}

	conversation := FilterConversation(messages)
	if strings.TrimSpace(conversation) == "" {
		return nil
	}

	prompt := BuildPrompt(conversation, cfg.Output.Language)
	summary, err := CallClaude(prompt, cfg.Summary.Model)
	if err != nil {
		return err
	}

	return SaveSummary(cfg, input.Cwd, summary)
}
