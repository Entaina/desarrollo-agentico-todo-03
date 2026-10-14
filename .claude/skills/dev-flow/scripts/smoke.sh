#!/usr/bin/env bash
# Prueba de humo contra la aplicación levantada: /healthz, GET de una página y,
# opcionalmente, un envío de formulario real con cookie de sesión y token CSRF.
#
# Uso:
#   smoke.sh                                  # solo /healthz y GET /
#   smoke.sh <ruta-post> <datos> <texto>      # además POST y comprueba <texto> en GET /
#
# <datos> va ya codificado como application/x-www-form-urlencoded en UTF-8,
# p. ej. 'title=Prueba&assignee=Ana+Garc%C3%ADa'. No uses --data-urlencode con
# texto no ASCII desde Git Bash en Windows: lo envía en Latin-1 y PostgreSQL lo rechaza.
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

post_path="${1:-}"
post_data="${2:-}"
expect="${3:-}"

section "Esperando $APP_URL/healthz"
for _ in $(seq 1 40); do
  curl -fsS "$APP_URL/healthz" >/dev/null 2>&1 && break
  sleep 3
done
curl -fsS "$APP_URL/healthz" || { echo "FALLO: /healthz no responde"; exit 1; }
echo

section "GET /"
code=$(curl -s -o /dev/null -w '%{http_code}' "$APP_URL/")
echo "GET / -> $code"
[ "$code" = 200 ] || exit 1

[ -n "$post_path" ] || { echo; echo "SMOKE OK"; exit 0; }

jar=$(mktemp)
trap 'rm -f "$jar"' EXIT

token=$(curl -s -c "$jar" -b "$jar" "$APP_URL/" \
  | grep -oE 'name="authenticity_token" value="[^"]+"' | head -1 \
  | sed -E 's/.*value="([^"]+)"/\1/')
[ -n "$token" ] || { echo "FALLO: no se encontró authenticity_token"; exit 1; }

section "POST $post_path"
code=$(curl -s -c "$jar" -b "$jar" -o /dev/null -w '%{http_code}' \
  --data-urlencode "authenticity_token=$token" --data "$post_data" "$APP_URL$post_path")
echo "POST $post_path -> $code"
case "$code" in
  2??|3??) ;;
  *) echo "FALLO: revisa el log del servidor"; exit 1 ;;
esac

if [ -n "$expect" ]; then
  section "Comprobando \"$expect\" en GET /"
  # Se lee el cuerpo una sola vez: con pipefail, `curl | grep -q` falla por SIGPIPE.
  body=$(curl -s -b "$jar" "$APP_URL/")
  matches=$(grep -F -- "$expect" <<<"$body")
  [ -n "$matches" ] || { echo "FALLO: texto no encontrado"; exit 1; }
  sed 's/^ *//' <<<"$matches" | head -5
fi

echo
echo "SMOKE OK"
