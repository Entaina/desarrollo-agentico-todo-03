---
description: Inspect the Go/Buffalo TODO application and establish working context
allowed-tools: Bash, Read, Glob, Grep
---

# Comprende el proyecto

Construye el contexto inicial necesario para trabajar en este repositorio. No modifiques ningún archivo.

## Exploración

1. Ejecuta `git status --short --branch` y `git ls-files`.
2. Lee, como mínimo:
   - `go.mod`;
   - `compose.yaml`;
   - `database.yml`;
   - `.buffalo.dev.yml`;
   - `cmd/app/main.go`;
   - `actions/app.go` y `actions/tasks.go`;
   - `models/task.go`;
   - `templates/tasks/index.plush.html`;
   - `.github/workflows/ci.yaml`.
3. Localiza los tests que cubren las capas anteriores y revisa los que resulten relevantes.
4. Adapta la exploración al modo indicado en `$ARGUMENTS`:
   - `exploración`: prioriza el mapa del sistema, sus límites, incógnitas y puntos de extensión;
   - `optimización`: localiza rutas críticas, señales medibles y posibles cuellos de botella sin asumir que ya están demostrados;
   - `diseño de producto`: reconstruye los flujos de usuario, el comportamiento observable y las decisiones de dominio;
   - cualquier otro argumento: úsalo como foco concreto y profundiza solo en los archivos relacionados.
5. Si no se indica un modo, construye el contexto general descrito arriba.

## Resultado

Resume de forma concisa:

- propósito y stack;
- estructura y responsabilidades principales;
- flujo de una petición desde la ruta hasta la persistencia y la vista;
- cómo arrancar, comprobar y probar la aplicación;
- convenciones que debería respetar un cambio;
- dudas o riesgos que todavía no puedan resolverse con el contexto leído.

Distingue hechos observados de inferencias. No propongas una implementación hasta recibir un encargo.
