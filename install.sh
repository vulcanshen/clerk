#!/usr/bin/env bash
set -euo pipefail

# ── Check dependencies ──────────────────────────────────────────
if ! command -v jq &>/dev/null; then
  echo "Error: jq is required but not installed."
  echo "  brew install jq    # macOS"
  echo "  apt install jq     # Debian/Ubuntu"
  exit 1
fi

if ! command -v claude &>/dev/null; then
  echo "Error: claude CLI is required but not installed."
  echo "  See https://docs.anthropic.com/en/docs/claude-code"
  exit 1
fi

# ── Resolve wrap-up.sh path ──────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
HOOK_SCRIPT="$SCRIPT_DIR/wrap-up.sh"

if [[ ! -x "$HOOK_SCRIPT" ]]; then
  echo "Error: $HOOK_SCRIPT not found or not executable."
  exit 1
fi

# ── Ask for CLAUDE_AUTO_DIGEST_ROOT ──────────────────────────────────────────
DEFAULT_ROOT="$HOME/.claude/claude-auto-digest"
read -rp "Digest output directory [$DEFAULT_ROOT]: " CLAUDE_AUTO_DIGEST_ROOT
CLAUDE_AUTO_DIGEST_ROOT="${CLAUDE_AUTO_DIGEST_ROOT:-$DEFAULT_ROOT}"
CLAUDE_AUTO_DIGEST_ROOT="${CLAUDE_AUTO_DIGEST_ROOT/#\~/$HOME}"

mkdir -p "$CLAUDE_AUTO_DIGEST_ROOT"

# ── Ask for settings scope ───────────────────────────────────────
echo ""
echo "Where to install the hook?"
echo "  1) User-level   (~/.claude/settings.json)"
echo "  2) Project-level (.claude/settings.json)"
read -rp "Choice [1]: " SCOPE_CHOICE
SCOPE_CHOICE="${SCOPE_CHOICE:-1}"

if [[ "$SCOPE_CHOICE" == "2" ]]; then
  SETTINGS_FILE="$PWD/.claude/settings.json"
else
  SETTINGS_FILE="$HOME/.claude/settings.json"
fi

# Ensure settings file exists
mkdir -p "$(dirname "$SETTINGS_FILE")"
if [[ ! -f "$SETTINGS_FILE" ]]; then
  echo '{}' > "$SETTINGS_FILE"
fi

# Check if hook is already installed
if jq -e ".hooks.SessionEnd[]?.hooks[]? | select(.command == \"$HOOK_SCRIPT\")" "$SETTINGS_FILE" &>/dev/null; then
  echo ""
  echo "Hook is already installed in $SETTINGS_FILE, skipping."
  exit 0
fi

# Build new config to merge
jq -n --arg cmd "$HOOK_SCRIPT" --arg root "$CLAUDE_AUTO_DIGEST_ROOT" '{
  env: {
    CLAUDE_AUTO_DIGEST_ROOT: $root
  },
  hooks: {
    SessionEnd: [
      {
        hooks: [
          {
            type: "command",
            command: $cmd,
            timeout: 120
          }
        ]
      }
    ]
  }
}' > /tmp/wrapup-config.json

# Merge: set env.CLAUDE_AUTO_DIGEST_ROOT, append to SessionEnd array (skip if already present)
jq --argjson new "$(cat /tmp/wrapup-config.json)" '
  .env //= {} |
  .env.CLAUDE_AUTO_DIGEST_ROOT = $new.env.CLAUDE_AUTO_DIGEST_ROOT |
  .hooks //= {} |
  .hooks.SessionEnd //= [] |
  (if (.hooks.SessionEnd | map(.hooks[]?.command) | index($new.hooks.SessionEnd[0].hooks[0].command))
   then .
   else .hooks.SessionEnd += $new.hooks.SessionEnd
   end)
' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp" && mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
rm -f /tmp/wrapup-config.json

echo ""
echo "✓ Installed in $SETTINGS_FILE"
echo "  Hook:  $HOOK_SCRIPT"
echo "  Root:  $CLAUDE_AUTO_DIGEST_ROOT"
