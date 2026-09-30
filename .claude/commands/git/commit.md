---
description: Review, stage and commit the requested changes safely
allowed-tools: Bash, Read, Glob, Grep
---

# Crea un commit

Revisa los cambios actuales, selecciona los archivos correspondientes y crea un commit local. No hagas push ni modifiques commits existentes.

`$ARGUMENTS` puede contener una descripción de los archivos que deben incluirse y, opcionalmente, `-m "mensaje"`.

## Flujo

1. Ejecuta `git status --short --branch` y revisa `git diff` y `git diff --cached`.
2. Determina el alcance solicitado:
   - si se nombran archivos o una parte del proyecto, incluye solo esos cambios;
   - si no se especifica alcance, propone el conjunto coherente que corresponda al trabajo actual;
   - no incluyas cambios ajenos solo porque estén presentes.
3. Antes de añadir nada, excluye y avisa sobre posibles secretos, credenciales, archivos `.env`, datos personales, binarios o artefactos generados.
4. Ejecuta `git add -- <archivos>` con una lista explícita. No uses `git add .` ni `git add -A`.
5. Revisa `git diff --cached --check` y `git diff --cached`.
6. Si el usuario proporcionó `-m`, conserva literalmente ese mensaje. En caso contrario genera un mensaje Conventional Commits con:
   - tipo y scope apropiados;
   - descripción imperativa, en minúsculas y sin punto final;
   - primera línea de 50 caracteres o menos.
7. Ejecuta `git commit` y muestra hash, mensaje y archivos incluidos.
8. Muestra los cambios que hayan quedado fuera del commit.

Si no hay cambios coherentes que incluir, no crees un commit vacío.
