# clerk

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/clerk)](https://github.com/vulcanshen/clerk/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/clerk)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/vulcanshen/clerk/ci.yml?label=CI)](https://github.com/vulcanshen/clerk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanshen/clerk)](https://goreportcard.com/report/github.com/vulcanshen/clerk)
[![License](https://img.shields.io/github/license/vulcanshen/clerk)](LICENSE)

[繁體中文](README.zh-TW.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Your Claude Code sessions disappear when you close the terminal. clerk makes sure you never lose track of what you did.

## The Problem

If you use Claude Code daily, you've hit these walls:

- **Lost context** — You forgot `-c` or `--resume`, and now you're starting from scratch. Your previous session had all the context, but good luck finding it in a pile of session IDs.
- **Session chaos** — Multiple projects, multiple sessions, all running in parallel. What did you do on the API server this morning? Which session had the auth fix? No idea.
- **Weekly report panic** — Friday afternoon, time for the weekly report. You're digging through `git log`, trying to reconstruct what you actually did all week.
- **Manual bookkeeping** — You told Claude to "save a summary" but forgot last time. Or the session crashed. Or you closed the terminal. The context is gone.

All of these boil down to one thing: **Claude Code doesn't remember across sessions, and you shouldn't have to.**

## The Solution

```bash
clerk install
```

That's it. clerk runs entirely on your machine — no remote services, no accounts, no data leaving your laptop. All you need is Claude Code.

It hooks into Claude Code and works silently in the background:

| Pain point | How clerk solves it |
|------------|-------------------|
| Lost context | `/clerk-resume` — instantly recover context from any previous session |
| Session chaos | Auto-generated daily summaries per project, organized by date |
| Weekly reports | `clerk report --days 7` — AI-generated report organized by date and project, ready to paste |
| Manual bookkeeping | Fully automatic — no commands to remember, no habits to build |

clerk is a **set-and-forget** tool. Install once, and every session is automatically summarized, tracked, tagged, and searchable. When you need context back, it's one slash command away. And when it's time for your weekly report, just ask:

```bash
clerk report --days 7
```

> **Note:** clerk is not an AI memory tool. Tools like claude-mem store context for AI to recall. clerk stores summaries for *you* to read — organized by date, searchable by keyword, ready for your weekly review.

## Features

- **Auto-summarize** — generates an incremental summary when your Claude Code session ends
- **Report generation** — `clerk report --days 7` produces a weekly report with summary, by-date, and by-project views
- **Context recovery** — `/clerk-resume` to rebuild context from previous sessions
- **Semantic search** — `/clerk-search` to find past work by tag with AI-powered semantic matching
- **Obsidian compatible** — output directory works as an Obsidian vault with tag graph view
- **Session tracking** — records every session start for history lookup
- **Tag system** — auto-extracts keywords from summaries for searchable indexing
- **Cursor tracking** — only processes new messages since the last run, saving tokens and time
- **Process management** — monitor active feeds, kill stuck processes, retry interrupted ones
- **Project-level config** — disable feed per-project, override global settings
- **One-command setup** — `clerk install` wires up hooks, MCP server, and skills
- Cross-platform: macOS, Linux, Windows
- Shell completion (bash, zsh, fish, powershell)

## How It Works

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

### Resume flow

1. You type `/clerk-resume` in Claude Code
2. Claude calls the `clerk-resume` MCP tool with your project's working directory
3. clerk returns file paths: daily summaries + full transcript files
4. Claude reads the summaries first for a quick overview
5. If more detail is needed, Claude reads the transcript files
6. Claude summarizes what was previously done and confirms context is restored

### Search flow

1. You type `/clerk-search` in Claude Code
2. Claude asks what keyword you're looking for (or you provide it as an argument)
3. Claude calls `clerk-tags-list` to get all available tags
4. Claude uses semantic reasoning to identify relevant tags (e.g. "database" → picks `postgres`, `sql`, `migration`)
5. Claude calls `clerk-tags-read` with the relevant tags to get summary and transcript paths
6. Claude reads the files and presents the relevant context

```
~/.clerk/
├── summary/
│   └── 20260416/
│       ├── projects-my-app.md
│       └── work-frontend.md
├── sessions/
│   ├── projects-my-app.md
│   └── work-frontend.md
├── tags/
│   ├── mcp.md
│   ├── refactor.md
│   └── auth.md
├── log/
│   └── 20260416-clerk.log
├── running/
└── cursor/
```

## Report

Friday afternoon, time for the weekly report? One command:

```bash
clerk report --days 7
```

clerk reads all summaries from the past 7 days, sends them to Claude, and outputs a structured report with three views:

- **Summary** — high-level overview of the entire period, organized by project
- **By Date** — what was done each day, broken down by project
- **By Project** — what was done on each project, broken down by date

Output goes to stdout. Save it, paste it, or pipe it wherever you need:

```bash
clerk report --days 7 > weekly-report.md
```

Default is `--days 1` (today only) — useful as a daily standup summary.

Want to include sessions that haven't ended yet? Add `--realtime`:

```bash
clerk report --days 7 --realtime
```

> **Note:** `--realtime` processes active session transcripts on the spot, which uses additional Claude API calls. Without this flag, only completed sessions are included.

Example output:

```markdown
### Summary (2026-04-14 ~ 2026-04-18)

#### my-api-server
Implemented user authentication with JWT, added rate limiting middleware,
and fixed connection pool leak under high concurrency.

#### frontend-app
Migrated from Vue 2 to Vue 3, replaced Vuex with Pinia, updated all unit tests.

---

### By Date

#### 2026-04-14
- **my-api-server**: Set up JWT auth with refresh token rotation
- **frontend-app**: Started Vue 3 migration, updated build config

#### 2026-04-16
- **my-api-server**: Added rate limiting middleware, fixed connection pool leak
- **frontend-app**: Replaced Vuex with Pinia, migrated 12 store modules

---

### By Project

#### my-api-server
- **2026-04-14**: JWT auth with refresh token rotation
- **2026-04-16**: Rate limiting middleware, connection pool leak fix

#### frontend-app
- **2026-04-14**: Vue 3 migration kickoff, build config update
- **2026-04-16**: Vuex → Pinia migration, 12 store modules converted
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
| `report` | Generate a report from recent summaries (default: today) |
| `report --days 7` | Weekly report across all projects |
| `version` | Print the version of clerk |
| `moveto <path>` | Move clerk data to a new directory and update config |
| `migrate` | Migrate data directory structure to the latest format (run after upgrading from v3.0.0) |

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
| `output.language` | `en` | Summary output language |
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

Available when MCP server is installed (`clerk install mcp`). These are called by Claude Code through skills — you don't need to invoke them directly.

| Tool | Description |
|------|-------------|
| `clerk-resume` | Returns summary + transcript file paths for context recovery |
| `clerk-tags-list` | List all available session tags |
| `clerk-tags-read` | Read the content of one or more tags |

## Skills

Available when skills are installed (`clerk install skills`):

| Skill | Description |
|-------|-------------|
| `/clerk-resume` | Recover context from previous sessions — calls MCP tool, reads files, rebuilds context |
| `/clerk-search` | Search past sessions by keyword — calls MCP tool, reads matching files |

## Troubleshooting

Logs are stored at `~/.clerk/log/YYYYMMDD-clerk.log`:

```bash
cat ~/.clerk/log/$(date +%Y%m%d)-clerk.log
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
