//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra/doc"
	"github.com/vulcanshen/clerk/cmd"
)

func main() {
	dir := "docs/man"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	header := &doc.GenManHeader{
		Title:   "CLERK",
		Section: "1",
		Date:    &time.Time{},
		Source:  "clerk",
		Manual:  "User Commands",
	}

	root := cmd.RootCmd()
	root.DisableAutoGenTag = true

	if err := doc.GenManTree(root, header, dir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("man pages generated in %s/\n", dir)
}
