# Changelog

All notable changes to this project will be documented in this file.

## [v5.1.1] - 2026-04-22

### New Features
- `report.instruction` config — custom instructions for report prompt (via `--append-system-prompt`)
- Report output language now passed via `--append-system-prompt` (system-level) instead of in user prompt

### Bug Fixes
- Tab completion no longer suggests files/directories for commands that don't accept file arguments

## [v5.1.0] - 2026-04-21

### New Features
- `clerk export` — list available slugs and dates for export
- `clerk export --summary <slug>` — export merged summaries for a project (across all dates)
- `clerk export --date <YYYYMMDD>` — export merged summaries for a date (across all projects)
- Tab completion for `--summary` (slugs) and `--date` (dates)
- `clerk report` now auto-saves to `<output.dir>/reports/clerk-report-YYYYMMDD-Nd.md` when run in terminal
- When piped (`clerk report | pbcopy`), outputs to stdout as before
- `-o` flag overrides the default path; duplicate filenames get auto-incrementing suffix (`-1`, `-2`, ...)
- `config show` non-JSON mode now displays key-value format instead of raw JSON
- `unregister` output moved to stderr
- Command table in all READMEs now shows API token usage per command
- `summary.instruction` config — custom instructions appended to feed prompt via `--append-system-prompt`
- `output.language` now passed via `--append-system-prompt` (system-level) instead of in user prompt

### Bug Fixes
- `config set output.dir` now converts relative paths to absolute (based on cwd)
- `config set output.dir` preserves `~` prefix for portability
- `config set output.dir` rejects empty string
- `config set summary.model` validates against known aliases (sonnet, opus, haiku) or full names containing one

### Internal
- `atomicWriteFile` shared helper for report and export output
- `applyKeyValue` handles validation only; `Set` handles normalization — no redundant logic

## [v5.0.2] - 2026-04-21

### New Features
- `clerk status --json` — structured JSON output for scripting
- `clerk config show --json` — pure JSON config output for machine consumption
- `clerk report -o` now supports tab completion for `.md`/`.txt` files
- `clerk data moveto` now supports tab completion for directories
- `clerk logs` shows spinner while redacting personal information via Claude API

### Bug Fixes
- `report -o ~/file.md`: tilde (`~`) is now expanded correctly (was writing to literal `~` directory)
- `register` now exits non-zero when issues remain unresolved (was always exit 0)
- `register` output moved to stderr (was polluting stdout)
- `punch` sessions file now uses file locking to prevent corruption from concurrent writes
- `writeCursor` now uses atomic write (temp+rename) to prevent cursor corruption on crash
- `status --json` outputs `[]` instead of `null` when no sessions exist
- `report --active` flush limited to 5 concurrent Claude API calls (was unbounded)
- `writeRunningState` now uses atomic write (temp+rename)
- `saveIndex` term files now use atomic write (temp+rename) instead of write+truncate
- `hook.go` and `mcp/install.go` output moved to stderr (was stdout)

### Documentation
- All READMEs: added token usage disclosure (API calls, cursor tracking, haiku option)

## [v5.0.1] - 2026-04-20

### Performance
- `clerk report --active`: flush active sessions in parallel (N sessions: N×5s → ~5s)
- `saveIndex`: write term files in parallel with goroutines, shared `sync.Map` stat cache
- `clerk register`: MCP check reduced from 2 `claude mcp list` calls to 1
- `clerk register`: `settings.json` read once and cached for all hook checks

### Bug Fixes
- `register`: hook FIXED message now shows updated path (was showing stale cached path)
- `register`: MCP FIXING message shows old path when pointing to different executable
- `register`: Claude API test shows spinner instead of appearing frozen
- `register`: migration output unified (FIXING/FIXED/FAILED format, no duplicate prints)

## [v5.0.0] - 2026-04-20

### New Features
- `clerk report` now shows step-by-step progress with spinner (only in terminal, silent when piped)
- Project config (`.clerk.json`) lookup walks up directories — works from subdirectories

### Breaking Changes
- `install` / `uninstall` → `register` / `unregister`
- `diagnosis` merged into `register` (register now checks, installs, and verifies everything)
- `diagnosis claude` merged into `register` (Claude API test runs automatically)
- `diagnosis error` / `diagnosis log` → `clerk logs --error` / `clerk logs`
- `install hook`, `install mcp`, `install skills` subcommands removed
- `uninstall hook`, `uninstall mcp`, `uninstall skills` subcommands removed
- `data purge` removed — use `rm -rf ~/.clerk/` instead
- Old commands (`install`, `uninstall`, `diagnosis`) show deprecation message and exit with error
- `logs` masks personal info by default (`--no-mask` to show raw)

### CI
- Upgrade GitHub Actions: checkout v4→v6, setup-go v5→v6, goreleaser-action v6→v7

## [v4.0.1] - 2026-04-19

### Bug Fixes
- Atomic summary writes (temp file + rename) prevent data loss on crash
- Concurrent feed serialization via per-slug file lock
- `punch` now uses project-level config (`LoadWithCwd`) — fixes session/summary path mismatch for projects with custom `output.dir`
- `moveto` checks destination non-empty before moving, defers source removal until all copies succeed
- `saveIndex` write-then-truncate ordering prevents partial file corruption
- `purge` only counts actually removed directories
- `toolNames` reset on each `NewServer()` call — no more duplicate tool names
- `fetchLatestVersion` checks HTTP status code before parsing response
- `readFrontmatterTags` / `updateSummaryFrontmatter` handle CRLF line endings
- `readSessionEntries` splits on last space for old-format entries with spaces in cwd
- `flushActiveSessions` respects per-project `feed.enabled` config
- `flushActiveSessions` reports flush errors to stderr
- Feed lock failure now aborts instead of proceeding unlocked
- `moveto` removes empty source directory after successful move

### Input Validation
- `log.retention_days` must be >= 1
- `summary.timeout` must be positive
- `--days` flag capped at 180 for `report`, `diagnosis error`, `diagnosis log`
- stdin limited to 1MB for `feed` and `punch` (LimitReader)
- `ExpandPath` handles `~` without trailing slash
- `version` command skips comparison for dev builds

### Diagnosis
- `diagnosis` auto-fixes MCP server and Skills when not installed
- `diagnosis` shows detailed info: hook binary path, MCP command, skills directory
- `diagnosis` shows config file paths (global + project if exists)
- `diagnosis` shows all config values with source (default / global / project)
- Home directory fail-fast check at startup (`$HOME` must be set)

### Platform
- Windows: `runtime.KeepAlive` prevents GC-related file descriptor invalidation in flock

### Testing
- Integration tests: feed pipeline end-to-end, cursor incremental processing, empty transcript skip, concurrent SaveSummary, cursor cleanup boundaries
- Session entry parsing tests: tab format with spaces, old space format, mixed formats

## [v4.0.0] - 2026-04-18

### Breaking Changes
- `tags/` directory replaced by unified `index/` directory
- MCP tools renamed: `clerk-tags-list` → `clerk-index-list`, `clerk-tags-read` → `clerk-index-read`
- Commands restructured: `retry`/`kill` → `status retry`/`status kill`, `moveto`/`purge` → `data moveto`/`data purge`
- `update` command renamed to `version`
- `migrate` command removed (merged into `diagnosis`)
- `--realtime` flag renamed to `--active`

### Unified Inverted Index
- All terms (tags, dates, slugs, keywords) indexed in `index/` directory
- Overlapping terms merge into single graph nodes in Obsidian
- Summary frontmatter includes all terms for Obsidian tag pane and graph view
- Index files use markdown links: `[slug+YYYYMMDD](../summary/YYYYMMDD/slug.md)`
- Migration from `tags/` rebuilds index from existing summaries

### New Features
- `clerk diagnosis claude` — test Claude API compatibility
- `clerk report -o <file>` — save report to file with progress indicator
- `summary.timeout` config (default 5m) — timeout for `claude -p` calls
- `clerk install` checks for Claude Code before proceeding
- `clerk uninstall` asks whether to remove data

### Security & Stability
- Path traversal protection on index terms
- Diagnosis verifies hook binary exists
- saveTags file locking read-through-lock fix
- saveTags truncate error handling
- CwdToSlug case-insensitive prefix match for Windows
- Session entries use tab separator (handles paths with spaces)
- Remove all filepath.EvalSymlinks (fixes brew upgrade path breakage)
- Diagnosis detects and auto-fixes Cellar versioned paths
- JSON-based settings.json parsing in diagnosis (replaces string splitting)
- HookInput struct deduplicated (punch uses feed.HookInput)
- cleanOldLogs runs once per process instead of every log write
- Path concatenation standardized to filepath.Join

### Documentation
- Mermaid flow diagram showing session lifecycle
- Obsidian integration section with summary/index format details
- Troubleshooting simplified: diagnosis first, then diagnosis error --mask
- summary.timeout documented in all READMEs

### Testing & CI
- Unit tests for CwdToSlug, ParseSummaryAndTags, FilterConversation, BuildPrompt, BuildTerms, ExpandPath, applyKeyValue, parseList
- CI runs tests on Ubuntu, Windows, and macOS

## [v3.6.2] - 2026-04-18

### Diagnosis
- Feed pipeline test: diagnosis now tests full feed flow (BuildPrompt → CallClaude → ParseSummaryAndTags)
- Clear error messages when API format changes: suggests updating clerk or reporting issue
- Export `ParseSummaryAndTags` for diagnosis use

## [v3.6.1] - 2026-04-18

### Testing
- Add unit tests for core functions: CwdToSlug, parseSummaryAndTags, FilterConversation, BuildPrompt, ExpandPath, applyKeyValue, parseList
- CI now runs tests on Ubuntu, Windows, and macOS before release

### Diagnosis
- `clerk diagnosis` now tests `claude -p` API call (with Y/n prompt, default yes)

## [v3.6.0] - 2026-04-18

### Command Restructure
- `retry` / `kill` moved under `status` (`status retry`, `status kill`)
- `moveto` / `purge` moved under `data` (`data moveto`, `data purge`)
- `update` renamed to `version` (shows current version + checks for updates)
- Old `version` subcommand removed (redundant with `--version` flag)
- `migrate` removed — merged into `diagnosis` (auto-fix on check)

### Diagnosis
- `diagnosis` now auto-fixes issues: hidden dirs, summary migration, hook format
- Fix failures logged for `diagnosis error` export
- Troubleshooting section in README updated: diagnosis first, then `diagnosis error --mask` for issues

### File Cleanup
- `cmd/doctor.go` → `cmd/diagnosis.go`
- `cmd/migrate.go` → `cmd/diagnosis_migrate.go` (functions only, no command)
- `cmd/version.go` removed
- `cmd/data.go` added (parent command)

## [v3.5.0] - 2026-04-18

### Windows
- Fix terminal window popup on hook trigger: use CREATE_NEW_CONSOLE + HideWindow to run feed/punch in a hidden console
- Revert cmd.exe wrapper approach (broke stdin piping)

### New Command
- `clerk purge` — delete all clerk data (summaries, tags, sessions, logs, cursors) with confirmation prompt (`-y` to skip)

### Bug Fix
- Remove cmd.exe /c start /b hook wrapping that broke feed on Windows

## [v3.4.0] - 2026-04-17

### Windows
- Fix terminal window flash on hook trigger: wrap hook commands with `cmd.exe /c start /b` on Windows
- No impact on macOS/Linux

### Commands
- Rename `doctor` → `diagnosis`, `doctor diagnosis` → `diagnosis error`
- Add `diagnosis log` — show all logs (not just errors) with `--days N`
- Add `--mask` flag to `diagnosis error` and `diagnosis log` — redact personal info via Claude before output
- MCP scope changed from `local` to `user` for global availability

### Docs
- Fix PowerShell completion instructions in all READMEs

## [v3.3.0] - 2026-04-17

### New Command
- `clerk doctor` — check if your environment is set up correctly (claude CLI, hook, MCP, skills, config, output dir, migration status)
- `clerk doctor diagnosis` — show error logs for troubleshooting (`--days N` to include multiple days)

## [v3.2.14] - 2026-04-17

### Bug Fix
- `clerk update`: show friendly message when GitHub API is unavailable or rate limited

## [v3.2.13] - 2026-04-17

### Bug Fix
- Windows install.sh (Git Bash): show friendly error when clerk.exe is locked by active Claude Code sessions

## [v3.2.12] - 2026-04-17

### Bug Fix
- Windows install script: show friendly error when clerk.exe is locked by active Claude Code sessions

## [v3.2.11] - 2026-04-17

### Bug Fix
- Windows: move HideConsoleWindow to main() before Cobra parsing, eliminating window flash entirely
- Fix version display consistency in `clerk update` (remove `v` prefix from latest version)

## [v3.2.10] - 2026-04-17

### Update Command Enhancement
- `clerk update` now checks GitHub API for latest version before showing upgrade instructions
- Shows "Already up to date" if current version matches latest
- Only shows upgrade command when an update is available
- Fix brew/scoop instructions to include index update (`brew update &&`, `scoop update &&`)

## [v3.2.9] - 2026-04-17

### New Command
- `clerk update` — detects install method and shows the appropriate update command
- Supports Homebrew, Scoop, go install, install script (Git Bash / PowerShell), and direct download

## [v3.2.8] - 2026-04-17

### Bug Fix
- Windows: hide console window for hook commands (feed, punch) using Windows API to eliminate terminal flash
- No impact on macOS/Linux or user-initiated commands

## [v3.2.7] - 2026-04-17

### Bug Fix
- Fix markdown links in tag files using backslashes on Windows
- MCP scope changed from `local` to `user` for global availability
- Windows: reduce terminal window flash by using CREATE_NO_WINDOW for background feed process

## [v3.2.6] - 2026-04-17

### Install
- Add `--force` flag to `clerk install` for force reinstall (uninstall then install)
- Works on all subcommands: `install --force`, `install hook --force`, etc.
- Useful for updating executable paths after upgrading or switching between dev/release builds

## [v3.2.5] - 2026-04-17

### Bug Fix
- Fix Windows hook/MCP install: use forward slashes in command paths so bash can execute them correctly
- Fixes issue where backslash paths with non-ASCII characters (e.g. Chinese usernames) caused hook failures

## [v3.2.4] - 2026-04-17

### Bug Fix
- Fix Windows path handling: backslashes in cwd now correctly converted to hyphens in slug
- Add defensive MkdirAll before writing session files
- Fix cwdToClaudeProjectSlug to handle Windows backslashes

## [v3.2.3] - 2026-04-17

### Config
- `config show` now displays executable path and version for easier debugging

## [v3.2.2] - 2026-04-17

### Report
- Report no longer flushes active sessions by default — saves Claude API calls
- Add `--realtime` flag to opt-in to including active sessions
- Realtime status message outputs to stderr to avoid polluting piped output

## [v3.2.1] - 2026-04-17

### Report Enhancement
- `clerk report` now includes active sessions — no need to close current session first
- Punch now stores `cwd` and `transcript_path` for active session processing
- Sessions file format extended: `- timestamp \`session_id\` cwd transcript_path`
- Backward compatible with old session entries (no migration needed)

## [v3.2.0] - 2026-04-17

### New Commands
- `clerk report` — generate a report from recent summaries (default: today, `--days 7` for weekly)
- Report outputs three views: summary (by time range), by-date, and by-project
- `clerk version` — print version (alternative to `--version`)
- `clerk moveto <path>` — move clerk data to a new directory and update `output.dir` config
- `clerk migrate` — migrate data directory structure to latest format, auto-reinstalls only installed components

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
