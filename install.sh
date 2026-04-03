#!/bin/sh
# dops installer — downloads the latest release binary for your platform.
# Usage: curl -fsSL https://raw.githubusercontent.com/rundops/dops/main/install.sh | sh
set -eu

REPO="rundops/dops"
INSTALL_DIR="${DOPS_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture.
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Fetch latest release tag.
LATEST="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)"
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release." >&2
  exit 1
fi
VERSION="${LATEST#v}"

ARCHIVE="dops_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

echo "Installing dops ${LATEST} (${OS}/${ARCH})..."

# Download and extract.
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"
tar xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install binary.
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/dops" "${INSTALL_DIR}/dops"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo mv "${TMPDIR}/dops" "${INSTALL_DIR}/dops"
fi

chmod +x "${INSTALL_DIR}/dops"

echo ""
echo "  dops ${LATEST} installed successfully!"
echo ""

# Check if install dir is in PATH.
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo "  Add dops to your PATH by adding this to your shell profile:"
    echo ""
    SHELL_NAME="$(basename "${SHELL:-/bin/sh}")"
    case "$SHELL_NAME" in
      zsh)
        echo "    echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc"
        echo "    source ~/.zshrc"
        ;;
      bash)
        echo "    echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
        echo "    source ~/.bashrc"
        ;;
      fish)
        echo "    fish_add_path ${INSTALL_DIR}"
        ;;
      *)
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
    esac
    echo ""
    ;;
esac

echo "  Get started:"
echo ""
echo "    dops init    # set up ~/.dops with a sample runbook"
echo "    dops         # launch the TUI"
echo ""
