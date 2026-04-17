# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v3.1.0] - 2026-04-17

### Search & Tags Overhaul
- MCP tools split: `clerk-search` replaced by `clerk-tags-list` + `clerk-tags-read`
- Semantic search flow: AI gets full tag list, uses reasoning to pick relevant tags, reads in one call
- `/clerk-search` skill updated to orchestrate the two-step tag discovery flow
- Tag prompt strengthened: strict lowercase keyword format, no spaces/sentences allowed
- Tag validation in parser: rejects tags with spaces or over 30 characters

### Obsidian Compatibility
- Summary files now include YAML frontmatter with `tags:` for Obsidian tag pane and graph view
- Tag files use standard markdown links for Obsidian graph connections
- Tag file format: `[timestamp] [type] [cwd] [link]` per line

### Directory Structure
- Hidden directories renamed to non-hidden: `.tags/` → `tags/`, `.cursor/` → `cursor/`, `.running/` → `running/`, `.sessions/` → `sessions/`, `.log/` → `log/`
- Summaries moved under `summary/` subdirectory: `summary/YYYYMMDD/<slug>.md`

### New Commands
- `clerk version` — print version (alternative to `--version`)
- `clerk moveto <path>` — move clerk data to a new directory and update `output.dir` config
- `clerk migrate` — migrate data directory structure to latest format (currently: moves `YYYYMMDD/` dirs into `summary/`)

## [v3.0.0] - 2026-04-16

### MCP Server
- `clerk mcp` — MCP stdio server for Claude Code integration
- `clerk-resume` tool — returns summary + transcript file paths for context recovery
- `clerk-search` tool — search previous sessions by keyword/tag
- `clerk install mcp` / `clerk uninstall mcp` — manage MCP server registration

### Skills
- `/clerk-resume` — skill to recover context from previous sessions via MCP tool
- `/clerk-search` — skill to search past sessions by keyword via MCP tool
- `clerk install skills` / `clerk uninstall skills` — manage skill files

### Tag System
- Auto-extract tags from summaries during feed (via `<!-- CLERK:TAGS -->` separator)
- Tags stored in `~/.clerk/.tags/<tag>.md` with summary + transcript paths
- File locking on tag writes, stale entry cleanup on save
- Partial keyword matching in search

### Install Restructure
- `clerk install` — install all components (hook + mcp + skills) in one command
- `clerk install hook` / `clerk install mcp` / `clerk install skills` — individual install
- `clerk uninstall` — remove all components in one command
- Removed `clerk hook install/uninstall` and `clerk mcp install/uninstall` (moved to `clerk install/uninstall`)

### Config Restructure
- Config path changed to `~/.config/clerk/.clerk.json` (global) + `<cwd>/.clerk.json` (project)
- Merge order: defaults → global → project
- `config set` defaults to project config, `-g` flag for global
- Project config only stores explicitly set fields
- New `feed.enabled` setting (default true) — disable feed per-project

### Session Tracking
- `clerk punch` — record session ID on SessionStart hook
- Session history stored in `~/.clerk/.sessions/<slug>.md`
- `clerk install hook` now registers both SessionStart (punch) and SessionEnd (feed)

### Summary Engine
- Incremental merge — each session merges into a single daily summary per project
- Cursor tracking — only processes new transcript lines since last run
- Structured prompt with 5 sections: Core Work, Supporting Work, Key Decisions, User Notes, Version Log
- Fork-to-background mechanism — hook returns immediately, feed runs as detached process

### Process Management
- `clerk status` / `status --watch` — show active feed processes and interrupted sessions
- `clerk retry <slug>` / `retry --all` — re-run interrupted feed processes
- `clerk kill <slug>` / `kill --all` — force-stop active feed processes
- Running state files with conversation backup for crash recovery
- `config set` with tab completion for keys

### Reliability
- File locking (flock/LockFileEx) on summary and tag writes
- Platform abstraction layer for Windows support (file locking, process detaching, process alive checks)
- Recursion guard via `CLERK_INTERNAL` environment variable
- Daily log files with automatic cleanup based on `log.retention_days`
- Cursor file cleanup matching log retention

### Release Infrastructure
- GoReleaser config for multi-platform builds (macOS, Linux, Windows × amd64, arm64)
- GitHub Actions CI/CD (test on push/PR, release on tag)
- Homebrew tap, Scoop bucket, deb/rpm packaging
- Install/uninstall scripts for shell and PowerShell
- Man page auto-generation from Cobra commands
- Multi-language README (English, 繁體中文, 日本語, 한국어)
- Shell completion (bash, zsh, fish, powershell)
