package punch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

type HookInput struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
}

func sessionsDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), ".sessions")
}

func Run(data []byte, cfg config.Config) error {
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parsing hook input: %w", err)
	}
	if input.SessionID == "" || input.Cwd == "" {
		return nil
	}

	slug := feed.CwdToSlug(input.Cwd)
	dir := sessionsDir(cfg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating sessions directory: %w", err)
	}

	filePath := filepath.Join(dir, slug+".md")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening sessions file: %w", err)
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(f, "- %s `%s`\n", ts, input.SessionID)
	return err
}
