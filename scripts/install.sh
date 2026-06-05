#!/usr/bin/env sh
set -eu

GO_BIN="${GO_BIN:-go}"
PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"

if ! command -v "$GO_BIN" >/dev/null 2>&1; then
  echo "go was not found. Set GO_BIN=/path/to/go or install Go first." >&2
  exit 1
fi

mkdir -p "$BIN_DIR"
"$GO_BIN" build -ldflags "-X github.com/ryanrodrigues25200525-svg/Apple-music-cli/cmd.Version=dev" -o "$BIN_DIR/mu" main.go

echo "Installed mu to $BIN_DIR/mu"
echo "Run: mu doctor"
