#!/bin/sh
# clerk uninstaller for macOS / Linux / Git Bash
# Usage: curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/uninstall.sh | sh

set -e

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  mingw*|msys*|cygwin*) OS="windows" ;;
esac

# Determine install locations to check
if [ "$OS" = "windows" ]; then
  CANDIDATES="$HOME/bin/clerk.exe"
else
  CANDIDATES="$HOME/.local/bin/clerk /usr/local/bin/clerk"
fi

FOUND=""
for path in $CANDIDATES; do
  if [ -f "$path" ]; then
    FOUND="$path"
    break
  fi
done

if [ -z "$FOUND" ]; then
  echo "clerk not found in expected locations."
  echo "Checked: $CANDIDATES"
  exit 1
fi

rm "$FOUND"
echo "removed $FOUND"

# Remove summaries if present
SAVE_DIR="$HOME/.clerk"
if [ -d "$SAVE_DIR" ]; then
  printf "Remove saved summaries in %s? [y/N]: " "$SAVE_DIR"
  read -r answer
  case "$answer" in
    y|Y|yes|YES)
      rm -rf "$SAVE_DIR"
      echo "removed $SAVE_DIR"
      ;;
    *)
      echo "kept $SAVE_DIR"
      ;;
  esac
fi

# Remove config if present
CONFIG_DIR="$HOME/.config/clerk"
if [ -d "$CONFIG_DIR" ]; then
  printf "Remove config in %s? [y/N]: " "$CONFIG_DIR"
  read -r answer
  case "$answer" in
    y|Y|yes|YES)
      rm -rf "$CONFIG_DIR"
      echo "removed $CONFIG_DIR"
      ;;
    *)
      echo "kept $CONFIG_DIR"
      ;;
  esac
fi

echo ""
echo "clerk uninstalled."
