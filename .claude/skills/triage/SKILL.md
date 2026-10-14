---
name: triage
description: Analiza un encargo y lo clasifica como feature, bug o chore, o lo detiene; resume los pasos necesarios para resolverlo sin planificar ni cambiar código
allowed-tools: Bash, Read, Glob, Grep
---

# Triage de un encargo

Recibe en `$ARGUMENTS` un encargo: texto libre o la URL de una issue de GitHub. Analiza y clasifica; no escribas planes ni cambies archivos. Es la primera fase de `/adw`, y también se puede usar sola.

## 1. Entender el encargo

1. Comprueba `git status --short --branch` y anota los cambios previos ajenos al encargo.
2. Si el argumento es una URL de issue, lee título, cuerpo y comentarios con `gh issue view <url> --comments`. Si `gh` no está disponible o falla, detente y explica el bloqueo.
3. Reformula el encargo en una frase y separa:
   - requisitos explícitos;
   - inferencias tuyas (márcalas como tales);
   - lo que falta para poder actuar.
4. Recorre lo justo del repositorio (rutas en `actions/app.go`, acciones, modelos, migraciones, templates, assets, tests, configuración) para contrastar el encargo con el estado real e identificar las capas afectadas. Profundizar es trabajo de la skill de planificación.

## 2. Clasificar

Elige un único tipo:

| Tipo | Cuándo |
|---|---|
| `bug` | El comportamiento actual contradice uno esperado y demostrable: un test, un requisito existente, un error 500, datos corruptos, una regresión. Hay que poder describir esperado frente a observado. |
| `feature` | Añade o cambia comportamiento observable para el usuario o la API que hoy no existe. |
| `chore` | No cambia el comportamiento observable: dependencias, versiones, refactor, CI, Docker, configuración, documentación, tooling. |

Reglas de desempate:

- «No funciona como me gustaría», sin evidencia de que el comportamiento actual sea incorrecto, es `feature`, no `bug`.
- Un refactor necesario para una funcionalidad forma parte de esa `feature`.

Decide `detener` en lugar de un tipo si:

- el encargo mezcla trabajos independientes (propón cómo dividirlo);
- es una pregunta o no pide un cambio;
- falta información imprescindible y cualquier elección cambiaría el resultado de forma sustancial (indica qué preguntar);
- requiere secretos, accesos o acciones irreversibles fuera del repositorio.

## 3. Impacto en la interfaz

Indica si el cambio afecta a lo que ve o hace el usuario en el navegador (rutas HTML, acciones que renderizan o redirigen, templates Plush, `public/assets/`). Esta decisión determina si `/adw` debe escribir un escenario e2e.

## Resultado

Responde con este bloque, sin texto adicional antes:

```
TRIAGE
tipo: feature | bug | chore | detener
slug: <kebab-case descriptivo, para plans/<slug>.md y el escenario e2e>
ui: sí | no
encargo: <una frase>
```

Y después:

1. **Justificación**: por qué ese tipo, con la evidencia del repositorio (rutas y archivos).
2. **Requisitos explícitos** e **inferencias**, por separado.
3. **Capas afectadas** previstas.
4. **Pasos para resolverlo**: la secuencia concreta (por ejemplo, `/feature …` → `/implement plans/<slug>.md` → `/e2e --rebuild`), o qué hace falta para desbloquearlo si es `detener`.
5. **Cambios previos ajenos** detectados en el árbol de trabajo.
