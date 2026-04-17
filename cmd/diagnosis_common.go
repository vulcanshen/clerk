package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/vulcanshen/clerk/internal/config"
	"github.com/vulcanshen/clerk/internal/feed"
)

func maskOutput(lines []string, cfg config.Config) string {
	raw := strings.Join(lines, "\n")

	prompt := fmt.Sprintf(`You are a log redaction tool. Replace any personally identifiable information in the following log lines with # symbols. This includes:
- Usernames in file paths (e.g. /Users/john/ → /Users/####/)
- Home directory names
- Any personal names, emails, or identifiers

Keep the log structure, timestamps, log levels, and error messages intact. Only mask the personal parts.
Output the redacted log lines only, no explanation.

%s`, raw)

	output, err := feed.CallClaude(prompt, cfg.Summary.Model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: --mask failed, showing raw output")
		return raw
	}
	return output
}
