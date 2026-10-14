#!/usr/bin/env bash
# Reproduce las comprobaciones de .github/workflows/ci.yaml y prueba que las
# migraciones nuevas se pueden revertir. Sale con código distinto de cero al primer fallo.
# Uso: validate.sh [--no-rollback]
set -uo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
cd "$REPO_ROOT"
export GO_ENV=test

rollback=1
[ "${1:-}" = "--no-rollback" ] && rollback=0

cleanup() {
  # soda vuelca el esquema en migrations/schema.sql: es un artefacto, no se versiona.
  if [ -f migrations/schema.sql ] && ! git ls-files --error-unmatch migrations/schema.sql >/dev/null 2>&1; then
    rm -f migrations/schema.sql
  fi
}
trap cleanup EXIT

fail() {
  echo "FALLO: $*"
  exit 1
}

section "gofmt -l ."
files=$(gofmt -l .)
[ -z "$files" ] || { echo "$files"; fail "archivos sin formato"; }
echo OK

section "go vet ./..."
go vet ./... || fail "go vet"
echo OK

section "PostgreSQL"
pg_isready -h 127.0.0.1 || fail "PostgreSQL no responde (sudo service postgresql start)"

section "soda migrate -e test"
soda migrate -e test || fail "migraciones"

if ! porcelain=$(git status --porcelain --untracked-files=all -- migrations 2>&1); then
  echo "$porcelain"
  fail "git status (en WSL sobre /mnt/c puede requerir: git config --global --add safe.directory $REPO_ROOT)"
fi
new_migrations=$(echo "$porcelain" | awk '$1=="??" && $2 ~ /\.up\./ {print $2}')
if [ "$rollback" = 1 ] && [ -n "$new_migrations" ]; then
  steps=$(echo "$new_migrations" | wc -l)
  section "Rollback de $steps migración(es) nueva(s) y reaplicación"
  soda migrate down -e test -s "$steps" || fail "migrate down"
  soda migrate -e test || fail "migrate up tras down"
fi

section "go test -p 1 -count=1 ./..."
out=$(go test -p 1 -count=1 -v ./... 2>&1)
status=$?
echo "$out" | grep -E '^(ok|FAIL|---|\?)|--- FAIL' | grep -v -- '--- PASS'
echo "PASS: $(echo "$out" | grep -cE -- '--- PASS')  FAIL: $(echo "$out" | grep -cE -- '--- FAIL')"
if [ "$status" -ne 0 ]; then
  echo "$out" | grep -B2 -A20 -- '--- FAIL' | head -120
  fail "tests"
fi

section "go build ./..."
go build ./... || fail "build"
echo OK

echo
echo "VALIDACIÓN OK"
