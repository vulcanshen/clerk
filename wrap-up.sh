#!/usr/bin/env bash
set -euo pipefail

# ── Recursion guard ──────────────────────────────────────────────
# claude -p itself triggers a SessionEnd hook when it finishes.
# WRAPUP_RUNNING prevents infinite recursion.
if [[ "${WRAPUP_RUNNING:-}" == "1" ]]; then
  exit 0
fi

# ── Config ───────────────────────────────────────────────────────
if [[ -z "${CLAUDE_AUTO_DIGEST_ROOT:-}" ]]; then
  echo "[wrap-up] CLAUDE_AUTO_DIGEST_ROOT not set. Configure it in ~/.claude/settings.json env." >&2
  exit 0
fi

# ── Read hook stdin (must happen in foreground) ──────────────────
INPUT="$(cat)"

# ── Self-background: save stdin to tmpfile, write placeholder, fork, exit ─
if [[ "${_WRAPUP_BG:-}" != "1" ]]; then
  TMPFILE="$(mktemp /tmp/wrapup-stdin.XXXXXX)"
  printf '%s' "$INPUT" > "$TMPFILE"

  # Write placeholder in foreground so user sees it immediately
  CWD_FG="$(printf '%s' "$INPUT" | jq -r '.cwd // empty')"
  if [[ -n "$CWD_FG" ]]; then
    SLUG_FG="$(echo "$CWD_FG" | awk -F/ '{print $(NF-1) "-" $NF}')"
    DATE_FG="$(date +%Y%m%d)"
    OUT_DIR_FG="$CLAUDE_AUTO_DIGEST_ROOT/$DATE_FG"
    mkdir -p "$OUT_DIR_FG"
    echo "=== CLAUDE DIGESTING ===" >> "$OUT_DIR_FG/$SLUG_FG.md"
  fi

  _WRAPUP_BG=1 _WRAPUP_STDIN="$TMPFILE" "$0" </dev/null >/dev/null 2>&1 &
  disown
  exit 0
fi

# ═══════════════════════════════════════════════════════════════════
# Everything below runs in background
# ═══════════════════════════════════════════════════════════════════

# ── Redirect stderr to log file ──────────────────────────────────
LOG_DIR="$CLAUDE_AUTO_DIGEST_ROOT/$(date +%Y%m%d)"
mkdir -p "$LOG_DIR"
exec 2>>"$LOG_DIR/wrap-up.err.log"
log_err() { echo "[wrap-up $(date +%Y-%m-%dT%H:%M:%S)] $*" >&2; }
trap 'log_err "Error on line $LINENO, exit code $?"' ERR

# ── Read input from tmpfile (written by foreground) ──────────────
INPUT="$(cat "$_WRAPUP_STDIN")"
rm -f "$_WRAPUP_STDIN"

TRANSCRIPT_PATH="$(printf '%s' "$INPUT" | jq -r '.transcript_path // empty')"
CWD="$(printf '%s' "$INPUT" | jq -r '.cwd // empty')"
SESSION_ID="$(printf '%s' "$INPUT" | jq -r '.session_id // empty')"
REASON="$(printf '%s' "$INPUT" | jq -r '.reason // empty')"

if [[ -z "$TRANSCRIPT_PATH" || -z "$CWD" ]]; then
  log_err "Missing transcript_path or cwd, skipping."
  exit 0
fi

if [[ ! -f "$TRANSCRIPT_PATH" ]]; then
  log_err "Transcript not found: $TRANSCRIPT_PATH"
  exit 0
fi

# ── Derive slug from cwd (last two path components) ─────────────
SLUG="$(echo "$CWD" | awk -F/ '{print $(NF-1) "-" $NF}')"

# ── Derive date and time ────────────────────────────────────────
DATE_DIR="$(date +%Y%m%d)"
TIME_HEADER="$(date +%H:%M)"

# ── Prepare output directory & placeholder ───────────────────────
OUTPUT_DIR="$CLAUDE_AUTO_DIGEST_ROOT/$DATE_DIR"
mkdir -p "$OUTPUT_DIR"

OUTPUT_FILE="$OUTPUT_DIR/$SLUG.md"

log_err "Starting digest for $SLUG"

# ── Extract metadata from transcript ────────────────────────────
SESSION_NAME="$(jq -r 'select(.type == "custom-title") | .customTitle // empty' "$TRANSCRIPT_PATH" | head -1 || true)"
[[ -z "$SESSION_NAME" ]] && SESSION_NAME="$SESSION_ID"

GIT_BRANCH="$(jq -r 'select(.type == "user") | .gitBranch // empty' "$TRANSCRIPT_PATH" | head -1 || true)"

# Session duration
FIRST_TS="$(jq -r 'select(.timestamp) | .timestamp' "$TRANSCRIPT_PATH" | head -1 || true)"
LAST_TS="$(jq -r 'select(.timestamp) | .timestamp' "$TRANSCRIPT_PATH" | tail -1 || true)"
DURATION=""
if [[ -n "$FIRST_TS" && -n "$LAST_TS" ]]; then
  FIRST_EPOCH="$(date -jf "%Y-%m-%dT%H:%M:%S" "${FIRST_TS%%.*}" +%s 2>/dev/null || date -d "${FIRST_TS}" +%s 2>/dev/null || echo "")"
  LAST_EPOCH="$(date -jf "%Y-%m-%dT%H:%M:%S" "${LAST_TS%%.*}" +%s 2>/dev/null || date -d "${LAST_TS}" +%s 2>/dev/null || echo "")"
  if [[ -n "$FIRST_EPOCH" && -n "$LAST_EPOCH" ]]; then
    DURATION="$(( (LAST_EPOCH - FIRST_EPOCH) / 60 ))m"
  fi
fi

# ── Extract conversation text ───────────────────────────────────
FILTERED="$(jq -r '
  select(.type == "user" or .type == "assistant") |
  .message // empty |
  if .content | type == "string" then
    .role + ": " + (.content | .[0:2000])
  elif .content | type == "array" then
    ([.content[] | select(.type == "text") | .text | .[0:2000]] | join(" ")) as $t |
    if $t == "" then empty else .role + ": " + $t end
  else empty end
' "$TRANSCRIPT_PATH" || true)"

if [[ -z "$FILTERED" ]]; then
  log_err "Empty filtered transcript, skipping."
  exit 0
fi

log_err "Starting claude -p for $SLUG (filtered: ${#FILTERED} bytes)"

# ── Build prompt & call claude ──────────────────────────────────
PROMPT='你是一個工作日誌摘要助手。請閱讀以下 Claude Code session 的 transcript（JSONL 格式），產生一段簡潔的中文摘要。

摘要要求：
- 用 3-8 個重點條列，說明這個 session 做了什麼
- 包含關鍵決策和原因
- 如果有產出或修改檔案，列出檔名
- 不要逐行翻譯對話，要提煉重點
- 不要加標題，直接從條列開始'

SUMMARY="$(printf '%s' "$FILTERED" | WRAPUP_RUNNING=1 claude -p "$PROMPT" 2>/dev/null)" || {
  log_err "claude -p failed with exit code $?, skipping."
  exit 0
}

if [[ -z "$SUMMARY" ]]; then
  log_err "Empty summary, skipping."
  exit 0
fi

# ── Build metadata line ─────────────────────────────────────────
META=""
[[ -n "$DURATION" ]] && META="$META | duration: $DURATION"
[[ -n "$GIT_BRANCH" && "$GIT_BRANCH" != "HEAD" ]] && META="$META | branch: $GIT_BRANCH"
[[ -n "$REASON" ]] && META="$META | reason: $REASON"

# Remove the last "=== CLAUDE DIGESTING ===" placeholder line (written by foreground)
if [[ -f "$OUTPUT_FILE" ]]; then
  LAST_LINE="$(tail -1 "$OUTPUT_FILE")"
  if [[ "$LAST_LINE" == "=== CLAUDE DIGESTING ===" ]]; then
    LINES=$(wc -l < "$OUTPUT_FILE")
    if [[ "$LINES" -le 1 ]]; then
      rm -f "$OUTPUT_FILE"
    else
      head -n $(( LINES - 1 )) "$OUTPUT_FILE" > "$OUTPUT_FILE.tmp"
      mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"
    fi
  fi
fi

{
  echo ""
  echo "## $TIME_HEADER $SESSION_NAME"
  echo ""
  if [[ -n "$META" ]]; then
    echo "_${META# | }_"
    echo ""
  fi
  echo "$SUMMARY"
} >> "$OUTPUT_FILE"

log_err "Digest saved to $OUTPUT_FILE"
