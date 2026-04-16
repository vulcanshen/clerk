# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

The Claude Code Clerk — auto-summarize your sessions.

clerk is a CLI tool that hooks into Claude Code's `SessionEnd` event, automatically generating conversation summaries and saving them as organized markdown files.

## Features

- **Auto-summarize** — generates a summary when your Claude Code session ends
- **Conversation filtering** — strips tool calls, keeps only user/assistant text
- **Date-organized** — summaries saved to `~/.clerk/YYYYMMDD/<project-slug>.md`
- **Append mode** — multiple sessions on the same project append to the same daily file
- **Configurable** — output directory, language, and model are all customizable
- **One-command setup** — `clerk hook install` wires everything up
- **Recursion guard** — prevents infinite loops when clerk calls `claude -p`
- Cross-platform: macOS, Linux, Windows
- Shell completion (bash, zsh, fish, powershell)

## How It Works

When a Claude Code session ends, the `SessionEnd` hook triggers `clerk feed`, which:

1. Reads the session transcript (JSONL)
2. Filters out tool calls, keeping only conversation text
3. Calls `claude -p` to generate a summary
4. Appends the summary to a date-organized markdown file

```
~/.clerk/
└── 20260416/
    ├── projects-my-app.md
    ├── projects-api-server.md
    └── work-frontend.md
```

## Installation

### Quick Install

macOS / Linux / Git Bash:

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/install.ps1 | iex
```

To update, run the same command again. To uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.ps1 | iex
```

### Package Managers

| Platform | Command |
|----------|---------|
| Homebrew (macOS / Linux) | `brew install vulcanshen/tap/clerk` |
| Scoop (Windows) | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

`.deb` and `.rpm` packages can be downloaded from the [Releases page](https://github.com/vulcanshen/clerk/releases).

### Build from Source

```bash
go install github.com/vulcanshen/clerk@latest
```

## Quick Start

```bash
# Install the SessionEnd hook
clerk hook install
```

That's it. Once the hook is installed, clerk runs completely in the background — no manual steps, no extra commands. Every time you exit a Claude Code session, a summary is automatically generated and saved. You can forget about it.

## Commands

| Command | Description |
|---------|-------------|
| `feed` | Process a session transcript and generate a summary (called by hook) |
| `config show` | Show current configuration and config file path |
| `hook install` | Install clerk as a Claude Code SessionEnd hook |
| `hook uninstall` | Remove clerk from Claude Code SessionEnd hooks |

## Configuration

Config file: `~/.config/clerk/config.json`

```json
{
  "output": {
    "dir": "~/.clerk/",
    "language": "zh-TW"
  },
  "summary": {
    "model": ""
  }
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `output.dir` | `~/.clerk/` | Root directory for summaries |
| `output.language` | `zh-TW` | Summary output language |
| `summary.model` | `""` (claude default) | Model to use for `claude -p` |

The config file is optional — clerk uses sensible defaults when it doesn't exist.

## Summary Format

Each summary is appended with a timestamp separator:

```markdown
---
### 14:30:25

## 使用者輸入摘要
- Asked to fix the authentication bug in auth.ts
- Requested a new login page component

## AI 回應摘要
- Found and fixed token validation issue
- Created LoginPage component with form handling
```

## Shell Completion

```bash
# Zsh
mkdir -p ~/.zsh/completions
clerk completion zsh > ~/.zsh/completions/_clerk
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
source ~/.zshrc

# Bash
clerk completion bash > /etc/bash_completion.d/clerk

# Fish
clerk completion fish > ~/.config/fish/completions/clerk.fish

# PowerShell
clerk completion powershell > clerk.ps1
```

## License

[GPL-3.0](LICENSE)
