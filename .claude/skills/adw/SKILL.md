---
name: adw
description: Workflow completo de un encargo — triage, plan (feature/bug/chore), implementación, e2e y, si el e2e pasa, commit y PR — con un resumen final de los pasos y resultados
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Skill
---

# Workflow: triage → plan → implementación → e2e → commit y PR

Recibe en `$ARGUMENTS` un encargo: texto libre o la URL de una issue de GitHub. Opciones:

- `--type feature|bug|chore`: fuerza la clasificación del triage.
- `--plan-only`: se detiene tras la fase de plan.
- `--no-e2e`: omite la fase e2e (indícalo como pendiente en el resumen).
- `--headed`: ejecuta el e2e con el navegador visible.
- `--remote <nombre>`: remoto al que hacer push (por defecto `origin`).
- `--base <rama>`: rama base del PR (por defecto `main`).
- `--no-pr`: se detiene tras el e2e, sin commit ni PR.

El trigger de `adws/` invoca este comando como `/adw <issue_url> --remote origin --base main`.

Coordinas skills existentes; no repitas su trabajo ni relajes sus reglas. Ejecuta las fases en orden y **detente en la primera que falle o se bloquee**: no la sortees rediseñando. Lleva un registro de cada fase (estado, evidencia, archivos) para el resumen.

Solo la fase 5 hace commit, push y PR, y solo si se cumplen sus condiciones. Nunca hagas merge, force-push, ni reescribas commits existentes, y no modifiques la issue (ni etiquetas, ni comentarios, ni cierre).

## Fase 1 — Triage

1. Invoca la skill `triage` con el encargo (y la URL si es una issue).
2. Toma del bloque `TRIAGE` el `tipo`, el `slug`, `ui` y el `encargo`. Si se pasó `--type`, úsalo en lugar del tipo y anótalo.
3. Si el tipo es `detener`, salta al resumen.
4. Anota los cambios previos ajenos que haya detectado: no deben acabar en el diff del encargo.

## Fase 2 — Plan

1. Fija la ruta `plans/<slug>.md`. Si ya existe, añade un sufijo (`-2`, `-3`…) en lugar de sobrescribirlo.
2. Invoca la skill del tipo (`feature`, `bug` o `chore`) con: `Ruta del plan: plans/<slug>.md.`, el encargo reformulado, los requisitos explícitos y las inferencias marcadas del triage.
3. Lee el plan y comprueba que existe, que tiene las secciones que exige su skill y que no contiene incertidumbres bloqueantes. Una incertidumbre es bloqueante si `/implement` tendría que elegir una solución nueva para resolverla.
4. Si hay bloqueantes o se pidió `--plan-only`, salta al resumen.

## Fase 3 — Implementación

1. Invoca la skill `implement` con `plans/<slug>.md`.
2. Si se detiene por un bloqueo, recoge el motivo y salta al resumen.
3. Revisa `git status --short` y `git diff --stat`. Solo deben aparecer archivos del alcance del plan, además de los cambios previos anotados en la fase 1.

Condición conocida del entorno: los tests de aplicación necesitan PostgreSQL y la base `todo_test` migrada, y se ejecutan con `GO_ENV=test`. Sin esa variable, las peticiones `POST`/`PUT` devuelven 403 por CSRF. Si ves ese síntoma, repite los tests con `GO_ENV=test` antes de darlos por fallidos.

## Fase 4 — E2E

Omítela con `--no-e2e`, o si la implementación no terminó con la validación en verde.

1. Si `ui: sí`, escribe o actualiza `.claude/skills/e2e/scenarios/<slug>.mjs` siguiendo el contrato y las reglas de escenarios de `.claude/skills/e2e/SKILL.md`:
   - deriva los pasos de los criterios de aceptación del plan, no de la implementación;
   - para un `bug`, reproduce el caso del plan y comprueba el comportamiento corregido;
   - usa los selectores del HTML que ha quedado tras la implementación.
   Si `ui: no`, no escribas escenario: el e2e sirve solo de regresión.
2. Invoca la skill `e2e` con `--rebuild` (y `--headed` si se pidió), para probar el código actual con las migraciones nuevas aplicadas. Ejecuta todos los escenarios: el nuevo y los existentes, como regresión.
3. Si falla un escenario:
   - si el escenario está mal (selector o expectativa que no sale del plan), corrígelo y repite una sola vez;
   - si falla el comportamiento de la aplicación, no cambies el código de la aplicación: anota el fallo, el paso y la captura, y salta al resumen como pendiente.
4. Si `/e2e --rebuild` se detiene porque otro proceso ocupa el puerto `3000`, no lo finalices: anótalo como bloqueo.

## Fase 5 — Commit y PR

Ejecútala solo si se cumplen **todas** estas condiciones; si falta alguna, no hagas commit y explica cuál en el resumen:

- no se pasó `--no-pr`, `--no-e2e` ni `--plan-only`;
- la validación de `/implement` terminó en verde;
- `/e2e --rebuild` terminó con `E2E OK` (todos los escenarios, incluido el de regresión);
- el diff del encargo se limita al alcance del plan, y ninguno de sus archivos tenía ya cambios previos ajenos anotados en la fase 1 (si se solapan, no se pueden separar con seguridad: detente);
- tras `git fetch <remote> <base>`, `git rev-list <remote>/<base>..HEAD` está vacío: si `HEAD` tiene commits que no están en la base, el PR los arrastraría; detente y explícalo.

Pasos:

1. **Rama.** Comprueba `git status --short --branch`. Crea una rama nueva `adw/<slug>` desde el `HEAD` actual con `git switch -c` (si ya existe local o en el remoto, añade un sufijo `-2`, `-3`…). Los cambios ajenos sin commitear viajan con el árbol de trabajo, pero no se incluyen en el commit.
2. **Commit.** Invoca la skill `git:commit` con la lista explícita de archivos del encargo y un mensaje Conventional Commits mediante `-m`:
   - archivos: los del diff del encargo, más `plans/<slug>.md` y, si existe, `.claude/skills/e2e/scenarios/<slug>.mjs`; nunca nada de `tmp/`, ni los cambios previos ajenos;
   - mensaje: tipo según el triage (`feat`, `fix` o `chore`), scope de la capa principal y descripción imperativa en el idioma de los commits recientes (`git log --oneline -10`); si el encargo es una issue, añade `Refs <url>` en el cuerpo;
   - termina el mensaje con las líneas de atribución que indique el entorno para los commits.
   Comprueba después con `git show --stat HEAD` que el commit contiene exactamente esos archivos.
3. **Push.** `git push -u <remote> adw/<slug>`. Si falla (permisos, remoto inexistente, rama protegida), no reintentes con otras opciones: anota el error y salta al resumen; el commit queda local.
4. **PR.** Crea el PR con `gh pr create --base <base> --head adw/<slug> --title "<primera línea del commit>" --body-file <archivo temporal en tmp/>`. El cuerpo contiene:
   - el encargo y su origen (si es una issue, `Refs <url>`; no uses `Closes` para no cerrarla al mergear);
   - el triage y la ruta del plan;
   - los cambios principales;
   - la validación ejecutada con sus resultados reales, incluido el e2e por escenario;
   - incertidumbres e inferencias pendientes de revisión;
   - al final, las líneas de atribución que indique el entorno para los PR.
   Si `gh` no está autenticado o la creación falla, anota el error: la rama ya está en el remoto y el PR se puede abrir a mano.
5. Muestra la URL del PR y comprueba con `gh pr view --json url,headRefName,baseRefName` que apunta a la rama y base correctas. No hagas merge ni esperes al CI.

## Resumen

Termina con un resumen breve de los pasos necesarios para resolver el encargo y lo que se hizo en cada uno:

1. **Encargo**: una frase y su origen (texto o issue).
2. **Triage**: tipo, justificación y si fue forzado.
3. **Plan**: ruta, decisiones principales y las inferencias señaladas.
4. **Implementación**: archivos cambiados y tests añadidos o actualizados.
5. **Validación**: cada comando ejecutado (`gofmt`, `go vet`, `go test`, `go build`, `/e2e`) y su resultado real, incluidos los fallos.
6. **E2E**: escenario creado o actualizado, resultado por escenario y si la app se reconstruyó.
7. **Commit y PR**: rama, hash del commit, archivos incluidos y URL del PR; o por qué no se hicieron.
8. **Estado**: `completado`, `detenido en <fase>` o `completado con pendientes`.
9. **Siguientes pasos**: revisar el PR y el CI, resolver incertidumbres, mergear; o lo que hace falta para desbloquear.

Indica qué fases no se ejecutaron y por qué. Separa hechos verificados de inferencias y no declares terminado nada que no hayas comprobado.
