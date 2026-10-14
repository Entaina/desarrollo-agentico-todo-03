#!/usr/bin/env bash
# Compila la aplicación fuera del repositorio y la arranca en modo desarrollo.
# Aplica las migraciones pendientes al arrancar (models.Migrate en cmd/app/main.go).
# Lanzar en segundo plano; no mata procesos existentes.
# PORT elige el puerto (3000 por defecto); con otro puerto el binario se llama todo-<PORT>
# para poder detener solo esa instancia.
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_ROOT"
export GO_ENV=development

if curl -fsS "$APP_URL/healthz" >/dev/null 2>&1 || ss -ltn 2>/dev/null | grep -q ":$PORT "; then
  echo "El puerto $PORT ya está en uso. Detén la instancia anterior antes de relanzar." >&2
  ss -ltnp 2>/dev/null | grep ":$PORT " >&2
  exit 1
fi

pg_isready -h 127.0.0.1 || { echo "PostgreSQL no responde." >&2; exit 1; }

bin_dir="$HOME/todo-run"
bin="$bin_dir/todo"
[ "$PORT" = 3000 ] || bin="$bin_dir/todo-$PORT"
mkdir -p "$bin_dir"
go build -o "$bin" ./cmd/app || exit 1
echo "Arrancando $bin en $APP_URL (pid $$)"
exec "$bin"
