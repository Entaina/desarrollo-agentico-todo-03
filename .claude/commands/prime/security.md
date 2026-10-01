---
description: Build security-focused context for the TODO application
allowed-tools: Bash, Read, Glob, Grep
---

# Prepara el contexto desde la seguridad

Carga el contexto inicial del repositorio desde una perspectiva de seguridad. No evalúes ni propongas un cambio concreto: identifica las partes, límites y mecanismos que alguien deberá conocer antes de trabajar. No modifiques archivos, no ejecutes acciones destructivas y no afirmes que el proyecto está libre de vulnerabilidades.

## Exploración

1. Ejecuta `git status --short --branch` y `git ls-files`.
2. Lee `go.mod`, `compose.yaml`, `database.yml`, `.buffalo.dev.yml` y la configuración de CI disponible.
3. Localiza rutas y acciones HTTP, manejo de sesiones o identidad, autorización, formularios, entradas de usuario, persistencia, templates y respuestas de error.
4. Revisa cómo se gestionan configuración, credenciales, variables de entorno, acceso a base de datos y dependencias.
5. Examina los tests relevantes para controles de acceso, validación de entradas y manejo de errores. Señala explícitamente los controles que no puedas comprobar con el código disponible.

## Resultado

Resume de forma concisa:

- superficie de ataque observada;
- controles de seguridad confirmados por el código;
- límites, supuestos y dudas, cada uno con la evidencia que lo motiva;
- zonas que convendría comprender o comprobar al recibir un encargo sensible;
- controles que no puedan verificarse con el código disponible.

Distingue hechos observados, inferencias y desconocidos. El resultado es contexto inicial, no una auditoría ni una garantía de seguridad.
