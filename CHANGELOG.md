# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v2.0.0] - 2026-04-16

- Initial release: Claude Code session summarizer CLI
- `feed` command — process SessionEnd hook, filter conversation, generate summary via `claude -p`
- `config show` — display current configuration
- `hook install / uninstall` — manage Claude Code SessionEnd hook in settings.json
- Recursion guard via `CLERK_INTERNAL` environment variable
- Nested config structure (`output.dir`, `output.language`, `summary.model`)
- Homebrew, Scoop, deb, rpm packaging
- Man pages (auto-generated from cobra commands)
- Shell completion (bash, zsh, fish, powershell)
