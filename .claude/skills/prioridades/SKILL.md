---
name: prioridades
description: Muestra las tareas de la base de desarrollo con su prioridad (Alta, Media, Baja) marcada con banderas de color
allowed-tools: Bash
---

# Visualiza las tareas por prioridad

Muestra las tareas de `todo_development` con su prioridad. Es una skill de **solo lectura**: no modifiques tareas, prioridades ni esquema, ni levantes o pares servicios.

Argumentos opcionales en `$ARGUMENTS`:

- `pendientes`: muestra solo las tareas no completadas.
- `alta`, `media` o `baja`: muestra solo ese nivel.

## Codificación de la prioridad

La columna `tasks.priority` es un entero (ver `models/task.go`):

| Valor | Nivel | Marca |
| --- | --- | --- |
| `1` | Alta | 🔴 🚩 |
| `0` | Media | 🟡 ⚑ |
| `-1` | Baja | ⚪ ⚐ |

Cualquier otro valor es un dato inválido: muéstralo como `❓ Desconocida (<valor>)` y avísalo.

## Pasos

1. Desde la raíz del repositorio, comprueba que el servicio `db` está en marcha con `docker compose ps db`. Si no lo está, informa y sugiere `/env:start`; no lo arranques tú.
2. Consulta las tareas, de más a menos prioritaria y, dentro de cada nivel, por fecha de creación:

   ```sh
   docker compose exec -T db psql -U postgres -d todo_development -At -F '|' -c \
     "SELECT priority, completed, title FROM tasks ORDER BY priority DESC, created_at ASC, id ASC;"
   ```

   Si la consulta falla porque no existe la columna `priority`, informa de que falta aplicar la migración `20261014000000_add_priority_to_tasks` y detente.
3. Aplica los filtros de `$ARGUMENTS`, si los hay.
4. Presenta el resultado:
   - Un resumen en una línea con el recuento por nivel, p. ej. `🔴 Alta: 2 · 🟡 Media: 3 · ⚪ Baja: 1 · Total: 6 (2 completadas)`.
   - Una tabla Markdown con las columnas `Prioridad`, `Tarea` y `Estado`. En `Prioridad` usa la marca y el nombre del nivel (`🔴 🚩 Alta`). En `Estado` pon `✅ Completada` o `⏳ Pendiente`. Si la tarea está completada, tacha el título con `~~título~~`.
   - Si no hay tareas, o el filtro no deja ninguna, dilo de forma explícita.
5. Termina con una línea que señale la tarea pendiente más prioritaria (la primera pendiente de la lista), o indica que no hay tareas pendientes.

No inventes tareas ni completes datos que no devuelva la consulta.
