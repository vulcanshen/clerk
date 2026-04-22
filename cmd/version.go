package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:               "version",
	Short:             "Show current version and check for updates",
	ValidArgsFunction: noFileComp,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", Version)

		latest, err := fetchLatestVersion()
		if err != nil || latest == "" {
			fmt.Printf("Latest version:  (unable to check — offline or rate limited)\n\n")
		} else {
			fmt.Printf("Latest version:  %s\n\n", latest)

			if Version != "dev" && Version == latest {
				fmt.Println("Already up to date.")
				return
			}
		}

		exe, err := os.Executable()
		if err != nil {
			printFallback()
			return
		}
		path := filepath.ToSlash(strings.ToLower(exe))

		switch {
		case strings.Contains(path, "/cellar/") || strings.Contains(path, "/homebrew/"):
			fmt.Println("Installed via Homebrew. Run:")
			fmt.Println()
			fmt.Println("  brew update && brew upgrade clerk")

		case strings.Contains(path, "/scoop/apps/"):
			fmt.Println("Installed via Scoop. Run:")
			fmt.Println()
			fmt.Println("  scoop update && scoop update clerk")

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

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/vulcanshen/clerk/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

func printFallback() {
	fmt.Println("Download the latest release from:")
	fmt.Println()
	fmt.Println("  https://github.com/vulcanshen/clerk/releases/latest")
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
