# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

The Claude Code Clerk — auto-summarize your sessions, recover context, search by keyword.

clerk is a CLI tool that hooks into Claude Code, automatically generating conversation summaries, tracking sessions, and providing MCP tools to recover context and search past work.

## Features

- **Auto-summarize** — generates an incremental summary when your Claude Code session ends
- **Context recovery** — `/clerk-resume` to rebuild context from previous sessions
- **Keyword search** — `/clerk-search` to find past work by tag
- **Session tracking** — records every session start for history lookup
- **Tag system** — auto-extracts keywords from summaries for searchable indexing
- **Cursor tracking** — only processes new messages since the last run, saving tokens and time
- **Process management** — monitor active feeds, kill stuck processes, retry interrupted ones
- **Project-level config** — disable feed per-project, override global settings
- **One-command setup** — `clerk install` wires up hooks, MCP server, and skills
- Cross-platform: macOS, Linux, Windows
- Shell completion (bash, zsh, fish, powershell)

## How It Works

```
clerk install
```

That's it. Once installed, clerk runs completely in the background — no manual steps, no extra commands. Every time you exit a Claude Code session, a summary is automatically generated and saved. You can forget about it.

When you need context from a previous session, use `/clerk-resume` in Claude Code. When you need to find past work, use `/clerk-search`.

### What gets installed

| Component | What it does |
|-----------|-------------|
| **hook** | SessionStart records session ID, SessionEnd triggers summary generation |
| **mcp** | MCP stdio server providing `clerk-resume` and `clerk-search` tools |
| **skills** | `/clerk-resume` and `/clerk-search` slash commands for Claude Code |

### Summary flow

1. Session ends → hook triggers `clerk feed`
2. Feed forks to background (hook returns immediately)
3. Reads only new messages since last run (cursor tracking)
4. Loads existing daily summary, calls `claude -p` to merge
5. Saves updated summary + extracts tags for search indexing

```
~/.clerk/
├── 20260416/
│   ├── projects-my-app.md
│   └── work-frontend.md
├── .sessions/
│   ├── projects-my-app.md
│   └── work-frontend.md
├── .tags/
│   ├── mcp.md
│   ├── refactor.md
│   └── auth.md
├── .log/
│   └── 20260416-clerk.log
├── .running/
└── .cursor/
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

Then set up the hooks, MCP server, and skills:

```bash
clerk install
```

### Package Managers

| Platform | Command |
|----------|---------|
| Homebrew (macOS / Linux) | `brew install vulcanshen/tap/clerk` |
| Scoop (Windows) | `scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket && scoop install clerk` |
| Debian / Ubuntu | `sudo dpkg -i clerk_<version>_linux_amd64.deb` |
| RHEL / Fedora | `sudo rpm -i clerk_<version>_linux_amd64.rpm` |

### Build from Source

```bash
go install github.com/vulcanshen/clerk@latest
```

## Commands

| Command | Description |
|---------|-------------|
| `install` | Install all components (hook + mcp + skills) |
| `install hook` | Install SessionStart/SessionEnd hooks only |
| `install mcp` | Register MCP server only |
| `install skills` | Install slash command skills only |
| `uninstall` | Remove all components |
| `config` | Show current configuration (alias for `config show`) |
| `config show` | Show merged configuration and file paths |
| `config set <key> <value>` | Set project-level config value |
| `config set -g <key> <value>` | Set global config value |
| `status` | Show active feed processes and interrupted sessions |
| `status --watch` | Live-refresh status every second |
| `retry <slug>` | Retry a specific interrupted session |
| `retry --all` | Retry all interrupted sessions |
| `kill <slug>` | Kill a specific active feed process |
| `kill --all` | Kill all active feed processes |

Internal commands (called by hooks, not by users):

| Command | Description |
|---------|-------------|
| `feed` | Process session transcript and generate summary |
| `punch` | Record session ID on session start |
| `mcp` | Start MCP stdio server |

## Configuration

### Config files

- Global: `~/.config/clerk/.clerk.json`
- Project: `<cwd>/.clerk.json` (overrides global)

### Available settings

```json
{
  "output": {
    "dir": "~/.clerk/",
    "language": "en"
  },
  "summary": {
    "model": ""
  },
  "log": {
    "retention_days": 30
  },
  "feed": {
    "enabled": true
  }
}
```

| Key | Default | Description |
|-----|---------|-------------|
| `output.dir` | `~/.clerk/` | Root directory for summaries |
| `output.language` | `zh-TW` | Summary output language |
| `summary.model` | `""` (claude default) | Model to use for `claude -p` |
| `log.retention_days` | `30` | Days to keep log and cursor files |
| `feed.enabled` | `true` | Enable/disable feed for this project |

### Examples

```bash
# Disable feed for a specific project
cd /path/to/unimportant-project
clerk config set feed.enabled false

# Use a cheaper model globally
clerk config set -g summary.model haiku

# Change output language globally
clerk config set -g output.language en
```

## MCP Tools

Available when MCP server is installed (`clerk install mcp`):

| Tool | Description |
|------|-------------|
| `clerk-resume` | Returns summary + transcript file paths for context recovery |
| `clerk-search` | Search previous sessions by keyword/tag |

## Skills

Available when skills are installed (`clerk install skills`):

| Skill | Description |
|-------|-------------|
| `/clerk-resume` | Recover context from previous sessions — calls MCP tool, reads files, rebuilds context |
| `/clerk-search` | Search past sessions by keyword — calls MCP tool, reads matching files |

## Troubleshooting

Logs are stored at `~/.clerk/.log/YYYYMMDD-clerk.log`:

```bash
cat ~/.clerk/.log/$(date +%Y%m%d)-clerk.log
```

Common issues:

- **No summary generated** — Check if `claude` is in your PATH
- **Hook cancelled** — clerk forks to background to avoid this; update to latest version
- **MCP tool not found** — Run `clerk install mcp` and restart the session

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
