package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Show how to update clerk to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n\n", Version)

		exe, err := os.Executable()
		if err != nil {
			printFallback()
			return
		}
		exe, _ = filepath.EvalSymlinks(exe)
		path := filepath.ToSlash(strings.ToLower(exe))

		switch {
		case strings.Contains(path, "/cellar/") || strings.Contains(path, "/homebrew/"):
			fmt.Println("Installed via Homebrew. Run:")
			fmt.Println()
			fmt.Println("  brew upgrade clerk")

		case strings.Contains(path, "/scoop/apps/"):
			fmt.Println("Installed via Scoop. Run:")
			fmt.Println()
			fmt.Println("  scoop update clerk")

		case strings.Contains(path, "/go/bin/"):
			fmt.Println("Installed via go install. Run:")
			fmt.Println()
			fmt.Println("  go install github.com/vulcanshen/clerk@latest")

		case strings.Contains(path, "/.local/bin/"):
			fmt.Println("Installed via install script. Run:")
			fmt.Println()
			fmt.Println("  curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh")

		case strings.Contains(path, "/appdata/local/clerk/"):
			if os.Getenv("MSYSTEM") != "" {
				fmt.Println("Installed via install script (Git Bash). Run:")
				fmt.Println()
				fmt.Println("  curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh")
			} else {
				fmt.Println("Installed via install script (PowerShell). Run:")
				fmt.Println()
				fmt.Println("  irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex")
			}

		default:
			printFallback()
		}
	},
}

func printFallback() {
	fmt.Println("Download the latest release from:")
	fmt.Println()
	fmt.Println("  https://github.com/vulcanshen/clerk/releases/latest")
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
