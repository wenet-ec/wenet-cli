#!/usr/bin/env sh
# install.sh
set -eu

REPO="${WENET_CLI_REPO:-wenet-ec/wenet-cli}"
VERSION="${WENET_CLI_VERSION:-latest}"
INSTALL_DIR="${WENET_CLI_INSTALL_DIR:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  linux|darwin) ;;
  *)
    echo "unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64)
    arch="amd64"
    ;;
  arm64|aarch64)
    arch="arm64"
    ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/wenet_${os}_${arch}.tar.gz"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/wenet_${os}_${arch}.tar.gz"
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

archive="$tmp_dir/wenet.tar.gz"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$archive"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$archive"
else
  echo "curl or wget is required" >&2
  exit 1
fi

tar -xzf "$archive" -C "$tmp_dir"
chmod +x "$tmp_dir/wenet"

if [ ! -d "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
  mv "$tmp_dir/wenet" "$INSTALL_DIR/wenet"
else
  sudo mv "$tmp_dir/wenet" "$INSTALL_DIR/wenet"
fi

echo "wenet installed to ${INSTALL_DIR}/wenet"
