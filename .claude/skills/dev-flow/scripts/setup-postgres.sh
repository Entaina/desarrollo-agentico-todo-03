#!/usr/bin/env bash
# Instala y prepara PostgreSQL en Linux/WSL con las credenciales de database.yml.
# Lo ejecuta la PERSONA en su propia terminal, escribiendo su contraseña:
#   sudo bash .claude/skills/dev-flow/scripts/setup-postgres.sh
# Idempotente.
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "Ejecuta este script con sudo." >&2
  exit 1
fi

if ! command -v pg_isready >/dev/null 2>&1; then
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y postgresql
fi

service postgresql start

su postgres -c "psql -v ON_ERROR_STOP=1 -c \"ALTER USER postgres PASSWORD 'postgres';\""

for db in todo_test todo_development; do
  if su postgres -c "psql -tAc \"SELECT 1 FROM pg_database WHERE datname = '$db'\"" | grep -q 1; then
    echo "La base de datos $db ya existe."
  else
    su postgres -c "createdb $db"
    echo "Base de datos $db creada."
  fi
done

pg_isready -h 127.0.0.1
echo "PostgreSQL listo."
