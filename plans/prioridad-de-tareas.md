# Feature: Prioridad de las tareas (baja / media / alta)

## Encargo y criterios de aceptación

Encargo literal: «add priority for each task, the user can set the priority of a task when create it and the user can update that. the existing todo task should have medium priority».

Requisitos explícitos:
1. Cada tarea tiene una prioridad.
2. El usuario elige la prioridad al crear la tarea.
3. El usuario puede cambiarla después.
4. Las tareas existentes quedan con prioridad **media**.

Inferencias (a confirmar en revisión):
- Niveles: `low` (Baja), `medium` (Media), `high` (Alta). La petición solo nombra «medium»; tres niveles es el mínimo natural. Los valores internos en inglés (`low|medium|high`) y las etiquetas en español, como el resto de la UI.
- Si el usuario no indica prioridad al crear, se asigna **media** (coherente con el valor por defecto de las existentes).
- La prioridad se muestra en la lista de tareas; **no** se pide ordenar ni filtrar por prioridad (queda fuera).
- Editar se hace en la pantalla de edición existente (`/tasks/{id}/edit`), no inline.

Criterios de aceptación:
1. Tras migrar, toda fila preexistente de `tasks` tiene `priority = 'medium'`; ninguna fila tiene prioridad nula o vacía.
2. El formulario de alta de `/` incluye un selector de prioridad con Baja / Media / Alta, con «Media» preseleccionada.
3. `POST /tasks` con `priority=high` guarda `high`; sin `priority` (o vacío) guarda `medium`; con un valor no válido (`urgente`) responde 422 con mensaje de error y no crea la tarea.
4. La lista muestra la prioridad de cada tarea de forma legible sin depender solo del color (texto visible o accesible).
5. `/tasks/{id}/edit` muestra el selector con la prioridad actual seleccionada; `PUT /tasks/{id}` con `priority=low` la actualiza y redirige a `/` (303) conservando título, estado y `created_at`.
6. `PUT /tasks/{id}` sin parámetro `priority` conserva la prioridad actual (no la resetea); con valor inválido responde 422 sin modificar la tarea.
7. Un error de validación del título al crear/editar conserva la prioridad elegida en el formulario re-renderizado.
8. Los tests existentes siguen pasando (salvo el ajuste del texto de ayuda de edición, ver riesgos).

## Comportamiento actual y propuesta

Actual (verificado):
- `tasks` = `id uuid`, `title string(200)`, `completed bool default false`, timestamps (`migrations/20260901000000_create_tasks.up.fizz`). Hay una migración de seed `…000001_seed_tasks.up.sql` con 3 filas que inserta sin columna de prioridad.
- `models.Task` valida solo el título en `Validate` y usa tags `form:"-"` para campos no editables.
- `TasksCreate` hace `c.Bind(task)` + `ValidateAndCreate`; `TasksUpdate` enlaza solo un struct `input{Title}` y actualiza solo el título (a propósito: «los demás datos no cambian»). Errores → 422 re-renderizando la plantilla.
- Plantillas: `templates/tasks/index.plush.html` (lista + formulario de alta en `.new-task-row`), `templates/tasks/edit.plush.html` (solo título; texto «Los demás datos de la tarea no cambian.»).
- Las migraciones se aplican al arrancar (`models.Migrate()` en `cmd/app/main.go`) mediante `pop.NewMigrationBox` sobre `migrations.FS()`.
- Fixtures: `fixtures/tasks.toml`, escenario «tres tareas», sin prioridad.

Propuesta: añadir la columna `priority` (string, `NOT NULL`, default `'medium'`), el campo `Priority` en el modelo con constantes y validación, un `<select>` en alta y edición, y una etiqueta de prioridad en cada tarea de la lista. Cambio vertical mínimo siguiendo los patrones actuales; sin JS.

## Alcance

Incluido:
- Migración fizz que añade `priority` con default `medium` (rellena las filas existentes).
- Modelo: constantes, validación, normalización del valor vacío a `medium`.
- Acciones `TasksCreate`/`TasksUpdate` y plantillas index/edit con selector.
- Etiqueta de prioridad en la lista y estilos.
- Tests de modelo y de acciones; fixture con prioridades variadas.

Excluido:
- Ordenar o filtrar por prioridad; cambiar el orden actual (`created_at asc, id asc`).
- Cambiar la prioridad desde la lista sin entrar a editar; JavaScript nuevo.
- Prioridades personalizables, numéricas o por usuario; fechas límite.
- Tipo `ENUM` de PostgreSQL (ver decisión en capa de persistencia).

## Diseño por capas

**Persistencia y migraciones**
- Nueva migración `migrations/20260901000002_add_priority_to_tasks.up.fizz` / `.down.fizz` (siguiente timestamp tras `…000001`; mantener el patrón `AAAAMMDDHHMMSS_nombre`):
  - up: `add_column("tasks", "priority", "string", {"default": "medium", "size": 10})` (por defecto fizz crea la columna `NOT NULL`; verificar tras generar el SQL). En PostgreSQL `ADD COLUMN … NOT NULL DEFAULT 'medium'` rellena las filas existentes con `medium`, lo que cumple el requisito 4 sin `UPDATE` aparte.
  - down: `drop_column("tasks", "priority")`.
- Decisión: columna `string` + validación en el modelo (mismo enfoque que el título, sin tipos especiales ni `CHECK`). Alternativa descartada: `CHECK (priority IN (…))` por SQL: más robusta ante escrituras fuera de la app, pero introduce un patrón (SQL de constraints) ausente del repo. Se acepta el riesgo; se puede añadir después.
- No modificar migraciones existentes (ya aplicadas). La migración de seed sigue funcionando porque la columna nueva aparece después y tiene default.

**Modelo** (`models/task.go`)
- Constantes `PriorityLow = "low"`, `PriorityMedium = "medium"`, `PriorityHigh = "high"` y `DefaultPriority = PriorityMedium`; helper `PriorityLabel(p string) string` (Baja/Media/Alta) y lista ordenada `Priorities` para renderizar las opciones.
- Campo `Priority string` con tags `json:"priority" db:"priority" form:"priority"`.
- `Validate`: si `Priority` no está vacía y no pertenece al conjunto → `verrs.Add("priority", "La prioridad no es válida.")`.
- Vacío: permitido en `Validate` y normalizado a `medium` por un hook `BeforeSave(tx)` (así los tests existentes que crean `&models.Task{Title: …}` con `ValidateAndCreate` siguen funcionando sin tocar). **Verificar** en pop v6.4.1 que `ValidateAndCreate` ejecuta `Validate` antes de `BeforeSave`; si el orden fuera el inverso, mover la normalización a un helper llamado desde las acciones y desde `Validate`.
- Una tarea nueva en la vista (`&models.Task{}` en `TasksIndex`) debe pintar «Media» preseleccionada: el template usa `if (task.Priority == "" || task.Priority == p)` o la acción asigna `DefaultPriority`.

**Acción y ruta** (`actions/tasks.go`; sin rutas nuevas)
- `TasksCreate`: `Bind` ya recoge `priority` por el tag `form`. Tras `Bind`, `task.Priority = strings.TrimSpace(task.Priority)`. En el re-render por error, la tarea conserva la prioridad elegida.
- `TasksUpdate`: ampliar el struct `input` con `Priority string \`form:"priority" json:"priority"\``. Si `strings.TrimSpace(input.Priority) != ""`, asignarla a `task.Priority`; si viene vacía, no tocar la actual (requisito 6; mantiene el test `PersistsTitleOnly`, que solo envía `title`). Sigue sin enlazarse `completed`, `id` ni `created_at`.
- `ValidateAndUpdate` devuelve 422 con `verrs` si la prioridad no es válida; en ese caso hay que evitar persistir (ya es el comportamiento) y **re-renderizar `edit` con la tarea tal como la envió el usuario**.
- Nota: `c.Set("task", task)` en el 422 de update pasa la entidad modificada en memoria, mismo patrón que hoy para el título.

**Renderizado (Plush)**
- `templates/tasks/index.plush.html`:
  - Formulario de alta: dentro de `.new-task-row` añadir `<label for="task-priority" class="sr-only">Prioridad</label>` + `<select id="task-priority" name="priority">` con las 3 opciones, `selected` según `task.Priority` (Media por defecto), y marcar `aria-invalid`/mensaje de error de `priority` si `errors.Get("priority")` no está vacío (bloque análogo a `title-error`, con `id="priority-error"`).
  - Cada tarea: `<span class="task-priority task-priority-<%= t.Priority %>">Prioridad: Alta</span>` junto al título (texto visible; no solo color). Mantener la rejilla `.task` (tres columnas) sin romperla: colocar la etiqueta dentro de la celda del título (p. ej. envolviendo título + etiqueta en el segundo `grid` cell) o ajustar `grid-template-columns` solo si hace falta.
- `templates/tasks/edit.plush.html`: añadir `<label for="task-priority">Prioridad</label>` + `<select name="priority">` con la actual seleccionada; actualizar el texto de ayuda «Los demás datos de la tarea no cambian.» (pasa a ser falso) a algo como «Hasta 200 caracteres. El estado de la tarea no cambia.».
- Plush escapa los valores por defecto; las clases CSS se derivan de un valor validado, pero nunca del input crudo: al renderizar la clase usar un helper/`PriorityLabel` o sanear con el conjunto permitido (si `Priority` fuera inválida en un 422 se pinta etiqueta neutra).

**CSS** (`public/assets/app.css`)
- Estilos `select` coherentes con `input[type="text"]` (alto 44px, borde, fuente), `.task-priority` (pastilla pequeña) y variantes `-low|-medium|-high` usando tokens `:root` existentes (añadir 1–2 variables si hace falta; contraste AA, texto no solo color).
- Ajustar `.new-task-row` para el selector (en móvil y en `max-width: 380px` hay reglas que fijan `flex-basis` del input; revisar que el selector no rompa el wrap).
- Subir `?v=3` de `app.css` en `templates/application.plush.html` (cache-busting).

**JavaScript:** sin cambios. El handler de `data-task-toggle` en `app.js` no toca la prioridad; la etiqueta no se re-renderiza en cliente.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20260901000002_add_priority_to_tasks.up.fizz` (nuevo) | Añade `priority` con default `medium`. |
| `migrations/20260901000002_add_priority_to_tasks.down.fizz` (nuevo) | Elimina la columna. |
| `migrations/embed.go` | Ya embebe `*.fizz`; sin cambios. |
| `models/task.go` | Campo, constantes, etiquetas, validación y normalización. |
| `models/task_test.go` | Tests de prioridad del modelo. |
| `actions/tasks.go` | Crear y actualizar con prioridad. |
| `actions/tasks_test.go` | Tests de acciones y renderizado. |
| `templates/tasks/index.plush.html` | Selector de alta y etiqueta en la lista. |
| `templates/tasks/edit.plush.html` | Selector de edición y texto de ayuda. |
| `templates/application.plush.html` | Bump de `?v=` del CSS. |
| `public/assets/app.css` | Estilos del `select` y de la etiqueta. |
| `fixtures/tasks.toml` | Prioridades variadas en el escenario «tres tareas». |

## Implementación paso a paso

1. Comprobar `git status --short --branch`; confirmar que la última migración es `20260901000001` y que no existe ya `priority` en el modelo.
2. Crear la migración `up`/`down` (paso de persistencia). Generar/inspeccionar el SQL resultante (p. ej. aplicándola sobre una base de desarrollo) y confirmar `NOT NULL DEFAULT 'medium'` y que las 3 filas del seed quedan en `medium`.
3. Editar `models/task.go`: constantes, `Priorities`, `PriorityLabel`, campo `Priority`, validación y `BeforeSave`. Confirmar el orden validate→BeforeSave en pop v6.4.1.
4. Editar `actions/tasks.go`: `TasksCreate` (trim) y `TasksUpdate` (campo `Priority` en `input`, asignación solo si no vacía).
5. Editar `templates/tasks/index.plush.html`: selector en el alta (con errores) y etiqueta por tarea.
6. Editar `templates/tasks/edit.plush.html`: selector y texto de ayuda actualizado.
7. Editar `public/assets/app.css` (select, etiqueta, responsive) y subir `?v=` en `application.plush.html`.
8. Actualizar `fixtures/tasks.toml` añadiendo `priority` a las 3 filas del escenario (p. ej. `high`, `medium`, `low`).
9. Escribir/actualizar los tests descritos abajo; ajustar los que dependan del texto de ayuda de edición si existen (hoy ninguno lo comprueba).
10. Ejecutar la validación y revisar a mano `/` y `/tasks/{id}/edit` en escritorio y móvil.

## Tests y validación

Modelo (`models/task_test.go`, `ModelSuite`):
- Crear sin prioridad (`&Task{Title:…}`) → válido y, tras recargar, `Priority == "medium"`.
- Crear con cada valor válido (`low`, `medium`, `high`) → se conserva.
- `Priority: "urgente"` → `verrs.Get("priority")` contiene el mensaje y no se crea la fila.

Acciones (`actions/tasks_test.go`, `ActionSuite`):
- `GET /` vacío: el selector de alta contiene las tres opciones y «Media» con `selected`.
- `GET /` con el escenario «tres tareas»: aparece la etiqueta de prioridad de cada tarea (`Prioridad: Alta`, etc.).
- `POST /tasks` con `priority=high` → 303 y fila con `high`; sin `priority` → `medium`; `priority=urgente` → 422, mensaje y 0 filas; título vacío con `priority=low` → 422 y el selector re-renderizado sigue en `low`.
- `GET /tasks/{id}/edit` preselecciona la prioridad actual de la tarea.
- `PUT /tasks/{id}` con `priority=low` → 303, prioridad `low`, título actualizado, `completed` y `created_at` intactos.
- `PUT /tasks/{id}` solo con `title` → la prioridad no cambia (el test existente `PersistsTitleOnly` debe seguir pasando; ampliarlo para fijar una prioridad `high` previa y comprobar que se conserva).
- `PUT /tasks/{id}` con `priority=urgente` → 422 y la tarea no cambia.
- Regresión: el resto de `Test_Tasks*` pasa sin cambios.

Migración: comprobar a mano (o con un test que ejecute la migración sobre filas previas, si el equipo lo quiere) que las filas del seed quedan en `medium`. No hay infraestructura de test de migraciones en el repo; la verificación manual es: aplicar hasta `…000001`, comprobar 3 filas, aplicar `…000002`, `SELECT priority FROM tasks` → todas `medium`; luego probar el `down`.

Condiciones de ejecución: PostgreSQL disponible (`docker compose up -d db`) y la base `todo_test` con las migraciones aplicadas, **incluida la nueva** (las migraciones las aplica `models.Migrate()` con `GO_ENV=test`; si el entorno de test usa un esquema preparado de antemano, hay que re-prepararlo, de lo contrario los tests fallarán con «column priority does not exist»).

Comandos:
- `gofmt -l .` (sin salida)
- `go vet ./...`
- `go test -p 1 ./...`
- `go build ./...`

## Riesgos y recuperación

- **Niveles asumidos (incertidumbre):** tres niveles y etiquetas en español son una inferencia. Cambiarlos después implica migración de datos.
- **Orden hook/validación en pop (por verificar):** si `BeforeSave` corre después de `Validate` es correcto con el diseño propuesto; si no, los tests existentes que crean tareas sin prioridad fallarían. El paso 3 lo comprueba.
- **Esquema de test desactualizado:** si la base de test no recibe la migración nueva, fallan casi todos los tests. Mitigación: ejecutar las migraciones antes de `go test`.
- **Texto de ayuda de edición:** «Los demás datos de la tarea no cambian.» deja de ser cierto; se actualiza. Ningún test actual lo comprueba (verificado por lectura de `actions/tasks_test.go`).
- **Sin `CHECK` en BD:** una escritura externa podría guardar un valor inválido; la lista debe tolerarlo mostrando una etiqueta neutra en vez de fallar o inyectar clases.
- **Maquetación:** el selector añadido a `.new-task-row` puede desbordar en móvil (reglas específicas a 760px y 380px); validar visualmente. No hay tests visuales.
- **Despliegue:** la migración se aplica al arrancar; `ADD COLUMN … DEFAULT` en PostgreSQL ≥ 11 es instantáneo, riesgo bajo con tablas pequeñas.
- **Recuperación:** revertir el commit y aplicar `down` (`drop_column`) si ya se migró; los datos de prioridad se perderían (no hay datos previos que preservar). Otros cambios son aditivos de código y plantillas.
