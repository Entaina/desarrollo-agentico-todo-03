#!/usr/bin/env bash
# Informa del entorno disponible para validar y levantar la aplicación.
# No modifica nada. Ejecutar en Linux/WSL: wsl -d Ubuntu -- bash <ruta>/env-detect.sh
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

required_go=$(awk '/^go /{print $2}' "$REPO_ROOT/go.mod")
echo "repo:        $REPO_ROOT"
echo "go.mod:      go $required_go"

if command -v go >/dev/null 2>&1; then
  echo "go:          $(go version)"
else
  echo "go:          NO (ejecuta setup-go-wsl.sh)"
fi

if command -v soda >/dev/null 2>&1; then
  echo "soda:        sí"
else
  echo "soda:        NO (ejecuta setup-go-wsl.sh)"
fi

if command -v pg_isready >/dev/null 2>&1 && pg_isready -h 127.0.0.1 -q; then
  echo "postgres:    responde en 127.0.0.1:5432"
elif command -v pg_isready >/dev/null 2>&1; then
  echo "postgres:    instalado pero parado (sudo service postgresql start)"
else
  echo "postgres:    NO (la persona debe ejecutar: sudo bash setup-postgres.sh)"
fi

if docker info >/dev/null 2>&1; then
  echo "docker:      daemon disponible"
else
  echo "docker:      daemon no disponible"
fi

for port in 3000 5432; do
  if ss -ltn 2>/dev/null | grep -q ":$port "; then
    echo "puerto $port: ocupado"
  else
    echo "puerto $port: libre"
  fi
done

if curl -fsS "$APP_URL/healthz" >/dev/null 2>&1; then
  echo "app:         ya responde en $APP_URL"
fi
