#!/usr/bin/env bash
set -euo pipefail

REPO="iqthink/setup"
INSTALL_DIR="${IQDEV_BIN_DIR:-$HOME/.local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" != "darwin" ]; then
  echo "iqdev only supports macOS for now. Detected: $OS"
  exit 1
fi

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if ! command -v brew >/dev/null 2>&1 \
  && [ ! -x /opt/homebrew/bin/brew ] \
  && [ ! -x /usr/local/bin/brew ]; then
  cat >&2 <<'EOF'
iqdev requires Homebrew, but it is not installed.

Install it first (in this terminal, so it can ask for your password):

  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

Then re-run this installer.
EOF
  exit 1
fi

echo "Looking up the latest iqdev release..."
VERSION=$(curl -sI "https://github.com/$REPO/releases/latest" \
  | grep -i '^location:' \
  | sed 's/.*tag\///' \
  | tr -d '\r\n' \
  || true)

if [ -z "${VERSION:-}" ]; then
  echo "Could not determine the latest version. Is there a published release?"
  exit 1
fi
echo "Latest version: $VERSION"

BINARY_NAME="iqdev-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/${BINARY_NAME}"
CHECKSUMS_URL="https://github.com/$REPO/releases/download/${VERSION}/SHA256SUMS-${OS}-${ARCH}.txt"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading $BINARY_NAME..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/$BINARY_NAME"
curl -fsSL "$CHECKSUMS_URL" -o "$TMPDIR/checksums.txt"

echo "Verifying checksum..."
EXPECTED=$(awk '{print $1}' "$TMPDIR/checksums.txt")
if [ -z "$EXPECTED" ]; then
  echo "Checksum file is empty"; exit 1
fi
ACTUAL=$(shasum -a 256 "$TMPDIR/$BINARY_NAME" | awk '{print $1}')
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "ERROR: checksum mismatch"
  echo "  Expected: $EXPECTED"
  echo "  Got:      $ACTUAL"
  exit 1
fi
echo "Checksum verified."

mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/$BINARY_NAME" "$INSTALL_DIR/iqdev"
chmod +x "$INSTALL_DIR/iqdev"

echo
echo "iqdev ${VERSION} installed to $INSTALL_DIR/iqdev"
echo

if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  SHELL_NAME=$(basename "${SHELL:-bash}")
  case "$SHELL_NAME" in
    zsh)  RC="$HOME/.zshrc"  ;;
    bash) RC="$HOME/.bashrc" ;;
    *)    RC="$HOME/.zshrc"  ;;
  esac
  if ! grep -qF "$INSTALL_DIR" "$RC" 2>/dev/null; then
    printf '\n# iqdev: ensure ~/.local/bin is on PATH\nexport PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "$RC"
    echo "Added $INSTALL_DIR to your PATH in $RC."
  fi
  export PATH="$INSTALL_DIR:$PATH"
fi

# Re-attach stdin to the controlling terminal so the TUI works when this
# script is run via `curl ... | bash`.
if [ -e /dev/tty ]; then
  exec "$INSTALL_DIR/iqdev" < /dev/tty
fi
exec "$INSTALL_DIR/iqdev"
