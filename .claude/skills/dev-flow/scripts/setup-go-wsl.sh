#!/usr/bin/env bash
# Instala sin root la versión de Go que exige go.mod en ~/.local/go y la CLI soda.
# Idempotente. Ejecutar en Linux/WSL.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

version=$(awk '/^go /{print $2}' "$REPO_ROOT/go.mod")
# Las descargas de Go usan siempre x.y.z.
if [[ "$version" =~ ^[0-9]+\.[0-9]+$ ]]; then
  version="$version.0"
fi

if command -v go >/dev/null 2>&1 && [ "$(go env GOVERSION)" = "go$version" ]; then
  echo "Go $version ya instalado."
else
  section "Descargando go$version"
  mkdir -p "$HOME/.local"
  tmp=$(mktemp)
  curl -fsSLo "$tmp" "https://go.dev/dl/go$version.linux-amd64.tar.gz"
  rm -rf "$HOME/.local/go"
  tar -C "$HOME/.local" -xzf "$tmp"
  rm -f "$tmp"
fi

line='export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"'
grep -qF "$line" "$HOME/.profile" || echo "$line" >> "$HOME/.profile"

go version

section "Instalando soda (misma versión que el CI)"
go install github.com/gobuffalo/pop/v6/soda@v6.4.1
command -v soda
