#!/usr/bin/env bash
# Utilidades comunes de los scripts de /dev-flow. Se carga con `source`.

# Raíz del repositorio: los scripts viven en .claude/skills/dev-flow/scripts.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
export REPO_ROOT

# Go instalado sin root por setup-go-wsl.sh, y binarios de `go install`.
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"

PORT="${PORT:-3000}"
APP_URL="${APP_URL:-http://localhost:$PORT}"
export PORT APP_URL

section() {
  echo
  echo "== $*"
}
