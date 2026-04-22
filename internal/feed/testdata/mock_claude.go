// mock_claude is a standalone program that mimics claude -p for testing.
// It reads stdin and returns a predefined summary with tags.
// Build: go build -o mock_claude ./testdata/mock_claude.go
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	// Read stdin (the prompt)
	input, _ := io.ReadAll(os.Stdin)
	prompt := string(input)

	// Check for --append-system-prompt in args
	var systemPrompt string
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--append-system-prompt" && i+1 < len(args) {
			systemPrompt = args[i+1]
		}
	}

	// If prompt contains "MOCK_HANG", sleep long enough to trigger timeout
	if strings.Contains(prompt, "MOCK_HANG") {
		time.Sleep(10 * time.Minute)
		os.Exit(0)
	}

	// If prompt contains "MOCK_FAIL", simulate an error
	if strings.Contains(prompt, "MOCK_FAIL") {
		fmt.Fprintln(os.Stderr, "mock error: simulated failure")
		os.Exit(1)
	}

	// Return a summary that echoes back some info for verification
	fmt.Printf(`### Core Work
Mock summary generated.

### Key Decisions & Rationale
- **Decision**: test → **Rationale**: mock

### User Notes
prompt_length: %d
system_prompt: %s

<!-- CLERK:TAGS -->
mock, test, cli`, len(prompt), systemPrompt)
}
