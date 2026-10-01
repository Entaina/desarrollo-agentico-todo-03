---
description: Understand the TODO application's codebase before making a change
allowed-tools: Bash, Read, Glob, Grep
---

# Prepara el contexto desde el código

Carga el contexto inicial del repositorio desde la perspectiva de quien necesita comprender su código. No resuelvas ni diagnostiques un encargo: construye una base reutilizable para el trabajo posterior. No modifiques ningún archivo.

## Exploración

1. Ejecuta `git status --short --branch` y `git ls-files`.
2. Lee `go.mod`, `compose.yaml`, `database.yml` y `.buffalo.dev.yml`.
3. Localiza los puntos de entrada, rutas, acciones, modelos, migraciones, templates y recursos estáticos relacionados.
4. Sigue al menos un flujo completo desde una interacción o petición HTTP hasta la persistencia y la respuesta renderizada.
5. Revisa los tests que cubren las capas implicadas y las instrucciones de ejecución, comprobación y build disponibles en el repositorio.

## Resultado

Resume de forma concisa:

- propósito y stack;
- mapa de directorios y responsabilidades;
- flujo de datos y control observado;
- cómo arrancar, comprobar y probar la aplicación;
- convenciones, límites y puntos de extensión que condicionan el trabajo posterior;
- incógnitas que sigan abiertas.

Distingue hechos observados de inferencias. El resultado es contexto inicial, no un plan ni una propuesta de implementación.
