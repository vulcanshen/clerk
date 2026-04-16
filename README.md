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
- **Incremental merge** — each session merges into a single daily summary per project, no duplicates
- **Conversation filtering** — strips tool calls, keeps only user/assistant text
- **Date-organized** — summaries saved to `~/.clerk/YYYYMMDD/<project-slug>.md`
- **Cursor tracking** — only processes new messages since the last run, saving tokens and time
- **Process management** — monitor active feeds, kill stuck processes, retry interrupted ones
- **Configurable** — output directory, language, model, and log retention are all customizable
- **One-command setup** — `clerk hook install` wires everything up
- **Recursion guard** — prevents infinite loops when clerk calls `claude -p`
- Cross-platform: macOS, Linux, Windows
- Shell completion (bash, zsh, fish, powershell)

## How It Works

When a Claude Code session ends, the `SessionEnd` hook triggers `clerk feed`, which:

1. Forks to background (so the hook returns immediately)
2. Reads only new messages from the transcript (JSONL) since the last run
3. Loads the existing daily summary for the project (if any)
4. Calls `claude -p` to produce a merged summary
5. Overwrites the daily summary file with the updated version

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
| `config` | Show current configuration (alias for `config show`) |
| `config show` | Show current configuration and config file path |
| `config set <key> <value>` | Set a configuration value (tab-completable keys) |
| `hook install` | Install clerk as a Claude Code SessionEnd hook |
| `hook uninstall` | Remove clerk from Claude Code SessionEnd hooks |
| `status` | Show active feed processes and interrupted sessions |
| `status --watch` | Live-refresh status every second |
| `retry <slug>` | Retry a specific interrupted session |
| `retry --all` | Retry all interrupted sessions |
| `kill <slug>` | Kill a specific active feed process |
| `kill --all` | Kill all active feed processes |

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
  },
  "log": {
    "retention_days": 30
  }
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `output.dir` | `~/.clerk/` | Root directory for summaries |
| `output.language` | `zh-TW` | Summary output language |
| `summary.model` | `""` (claude default) | Model to use for `claude -p` |
| `log.retention_days` | `30` | Days to keep log and cursor files |

Set values with `clerk config set`:

```bash
clerk config set output.language en
clerk config set summary.model haiku
clerk config set log.retention_days 14
```

The config file is optional — clerk uses sensible defaults when it doesn't exist.

## Summary Format

Each project gets one summary file per day, incrementally merged:

```markdown
# projects-my-app

> Last updated: 14:30:25

### Core Work
- Implemented user authentication with JWT tokens
- Fixed race condition in WebSocket handler

### Supporting Work
- Added CI pipeline with GitHub Actions
- Updated README with API documentation

### Key Decisions & Rationale
- **Decision**: Use JWT over sessions → **Rationale**: Stateless scaling for multi-region deploy

### User Notes
- Prefers minimal abstractions, direct code over frameworks

### Version Log
- v1.0.0 — Initial release with auth and WebSocket support
```

## Troubleshooting

Logs are stored at `~/.clerk/.log/YYYYMMDD-clerk.log`. Check them if summaries aren't appearing:

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

Common issues:

- **No summary generated** — Check if `claude` is in your PATH
- **Hook cancelled** — clerk forks to background to avoid this; update to latest version
- **Duplicate content** — Old behavior; current version uses incremental merge

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
