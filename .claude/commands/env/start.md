---
description: Start the TODO development environment and verify that the application responds
allowed-tools: Bash, Read
---

# Arranca el entorno de desarrollo

Levanta la aplicación desde la raíz del repositorio. No mates procesos existentes ni elimines volúmenes.

## Modos

- Usa el modo **host** por defecto.
- Usa el modo **Docker completo** si `$ARGUMENTS` contiene `--docker` o si Go/Buffalo no están disponibles en el host.

## Comprobaciones previas

1. Confirma que `docker compose` está disponible y que el daemon responde.
2. Comprueba si los puertos `3000` o `5432` ya están ocupados.
   - Si pertenecen a este proyecto y el servicio responde, reutilízalos.
   - Si pertenecen a otro proceso, informa del conflicto y detente. No lo finalices.

## Modo host

1. Comprueba que `go` y `buffalo` están disponibles.
   - Si falta alguno, explica el bloqueo y continúa mediante Docker completo.
2. Ejecuta `docker compose up -d db`.
3. Espera a que PostgreSQL figure como saludable en `docker compose ps`.
4. Ejecuta `go mod download`.
5. Inicia `buffalo dev` en segundo plano desde la raíz y conserva su salida para poder diagnosticar el arranque.
6. Espera hasta que `http://localhost:3000/healthz` responda correctamente.

## Modo Docker completo

1. Ejecuta `docker compose up --build -d`.
2. Espera a que `db` y `app` figuren como saludables en `docker compose ps`.
3. Comprueba `http://localhost:3000/healthz`.

## Resultado

Informa de:

- modo utilizado;
- servicios iniciados;
- URL de la aplicación;
- respuesta de `/healthz`;
- cualquier fallback, advertencia o proceso que deba mantenerse activo.
