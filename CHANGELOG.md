# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v2.1.0] - 2026-04-16

### Core
- `feed` command — process SessionEnd hook, filter conversation, generate summary via `claude -p`
- `config show` — display current configuration
- `config set <key> <value>` — set configuration values with tab completion
- `hook install / uninstall` — manage Claude Code SessionEnd hook in settings.json
- `status` — show active feed processes and interrupted sessions
- `status --watch` — live-refresh status every second
- `retry` — re-run interrupted feed processes with slug selection and `--all` flag
- `kill` — force-stop active feed processes with slug selection and `--all` flag

### Summary Engine
- Incremental merge — each session merges into a single daily summary per project
- Cursor tracking — only processes new transcript lines since last run, saving tokens
- Structured prompt with 5 sections: Core Work, Supporting Work, Key Decisions, User Notes, Version Log
- Prior summary passed to claude -p for context-aware merging

### Reliability
- Fork-to-background mechanism — hook returns immediately, feed runs as detached process
- Running state files — track active processes with slug, cwd, conversation, and start time
- Orphan detection — interrupted processes are detected and recoverable via `retry`
- File locking (flock/LockFileEx) — prevents race conditions on concurrent writes
- Recursion guard via `CLERK_INTERNAL` environment variable

### Platform
- Cross-platform support: macOS, Linux, Windows
- Platform abstraction layer for file locking, process detaching, and process alive checks

### Configuration
- Nested config structure: `output.dir`, `output.language`, `summary.model`, `log.retention_days`
- Config file at `~/.config/clerk/config.json` (optional, sensible defaults)

### Logging & Cleanup
- Daily log files at `~/.clerk/.log/YYYYMMDD-clerk.log`
- Automatic cleanup of old log and cursor files based on `log.retention_days`

### Release Infrastructure
- GoReleaser config for multi-platform builds (macOS, Linux, Windows × amd64, arm64)
- GitHub Actions CI/CD (test on push/PR, release on tag)
- Homebrew tap, Scoop bucket, deb/rpm packaging
- Install/uninstall scripts for shell and PowerShell
- Man page auto-generation from Cobra commands
- Multi-language README (English, 繁體中文, 日本語, 한국어)
- Shell completion (bash, zsh, fish, powershell)
