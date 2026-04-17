package main

import (
	"os"

	"github.com/vulcanshen/clerk/cmd"
	"github.com/vulcanshen/clerk/internal/platform"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "feed" || os.Args[1] == "punch") {
		platform.HideConsoleWindow()
	}
	cmd.Execute()
}
