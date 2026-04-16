#!/bin/sh
# clerk installer for macOS / Linux / Git Bash
# Usage: curl -fsSL https://raw.githubusercontent.com/vulcanshen/clerk/main/install.sh | sh

set -e

REPO="vulcanshen/clerk"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Error: unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
echo "Fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')
echo "Latest version: $VERSION"

# Set file extension and install dir
if [ "$OS" = "windows" ]; then
  EXT="zip"
  INSTALL_DIR="$HOME/bin"
else
  EXT="tar.gz"
  if [ "$(id -u)" = "0" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

FILENAME="clerk_${VERSION}_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/v${VERSION}/$FILENAME"

# Download
TMPDIR=$(mktemp -d)
echo "Downloading $FILENAME..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/$FILENAME"

# Extract
echo "Extracting..."
if [ "$EXT" = "zip" ]; then
  unzip -o "$TMPDIR/$FILENAME" -d "$TMPDIR" > /dev/null
else
  tar xzf "$TMPDIR/$FILENAME" -C "$TMPDIR"
fi

# Install
mkdir -p "$INSTALL_DIR"
if [ "$OS" = "windows" ]; then
  cp "$TMPDIR/clerk.exe" "$INSTALL_DIR/clerk.exe"
else
  cp "$TMPDIR/clerk" "$INSTALL_DIR/clerk"
  chmod +x "$INSTALL_DIR/clerk"
fi

# Install man pages (Unix only, skip on Git Bash)
if [ "$OS" != "windows" ]; then
  if [ -d "$TMPDIR/docs/man" ]; then
    if [ "$(id -u)" = "0" ]; then
      MAN_DIR="/usr/local/share/man/man1"
    else
      MAN_DIR="$HOME/.local/share/man/man1"
    fi
    mkdir -p "$MAN_DIR"
    cp "$TMPDIR"/docs/man/*.1 "$MAN_DIR/" 2>/dev/null || true
  fi
fi

# Cleanup
rm -rf "$TMPDIR"

echo ""
echo "clerk $VERSION installed to $INSTALL_DIR"

# Check if install dir is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo ""
    echo "WARNING: $INSTALL_DIR is not in your PATH."
    echo "Add it by running:"
    echo ""
    if [ "$OS" = "windows" ]; then
      echo "  echo 'export PATH=\"\$HOME/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
    else
      echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.$(basename "$SHELL")rc && source ~/.$(basename "$SHELL")rc"
    fi
    ;;
esac

echo ""
echo "Run 'clerk --help' to get started."
