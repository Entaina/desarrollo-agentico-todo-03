# Feature: prioridad de tareas

## Encargo y criterios de aceptación

Encargo inicial: «agrega la opción para añadir a la tarea un responsable, prioridad (alta, media, baja, urgente)».

Revisión posterior del encargo:

- Se elimina el responsable.
- La prioridad no se aplica a las tareas ya creadas: esas tareas quedan en `media`.
- Aclaración confirmada por la persona: la prioridad **solo se elige al crear** la tarea; la edición no la muestra ni la cambia.

Criterios de aceptación:

1. El formulario de creación muestra un selector de prioridad (Urgente, Alta, Media, Baja) con `Media` preseleccionada.
2. Al crear una tarea, la prioridad elegida se persiste; si no se envía, se guarda `media`.
3. Una prioridad fuera de la lista devuelve 422 con «La prioridad no es válida.» y no persiste nada.
4. La lista muestra una etiqueta de prioridad en cada tarea.
5. La edición no muestra la prioridad. Un `PUT /tasks/{id}` con `priority` la ignora: solo cambia el título y conserva la prioridad, `completed`, `id` y `created_at`.
6. Las tareas existentes antes de la migración quedan con prioridad `media`.
7. No existe ningún campo de responsable.

## Comportamiento actual y propuesta

Actual (observado en `main`):

- `models/task.go`: `Task` tiene `ID`, `Title`, `Completed`, `CreatedAt` y `UpdatedAt`, y `Validate` solo comprueba el título.
- `actions/tasks.go`: `TasksCreate` enlaza el formulario completo. `TasksUpdate` enlaza un struct con solo `title`.
- Las plantillas solo muestran y envían el título.
- Las migraciones fizz están embebidas y se aplican al arrancar mediante `models.Migrate()`.

Propuesta: un cambio vertical mínimo de migración → modelo → acción de creación → índice → CSS. `TasksUpdate` y `edit.plush.html` no cambian; su struct de entrada ya excluye la prioridad.

## Alcance

Incluido:

- columna `priority`;
- validación, valor por defecto y etiqueta en el modelo;
- selector en el alta y etiqueta en la lista;
- estilos;
- tests.

Excluido:

- responsable;
- editar la prioridad;
- ordenar o filtrar por prioridad;
- cambios en `public/assets/app.js`;
- cambios en el seed `20260901000001_seed_tasks`.

## Diseño por capas

**Persistencia**: `migrations/20261014000000_add_priority_to_tasks.{up,down}.fizz`.

```
add_column("tasks", "priority", "string", {"size": 10, "default": "media"})
```

La columna es `NOT NULL`, y el `default` asigna `media` a las filas existentes. El down ejecuta `drop_column("tasks", "priority")`.

**Modelo** (`models/task.go`):

- constantes `PriorityUrgent`, `PriorityHigh`, `PriorityMedium` y `PriorityLow`;
- `Priorities []PriorityOption{Value, Label}`;
- campo `Priority`;
- `Validate`: la prioridad debe estar vacía o pertenecer a la lista;
- `BeforeSave`: si está vacía, asigna `media`;
- `PriorityLabel()`.

**Acciones** (`actions/tasks.go`):

- `TasksIndex` crea `task` con `Priority: media` y expone `priorities`;
- `TasksCreate` también expone `priorities` en la respuesta 422;
- `TasksEdit` y `TasksUpdate` no cambian, salvo un comentario que documenta que la prioridad no se edita.

**Vistas**: en `index.plush.html`, la etiqueta `.priority .priority-<valor>` dentro de `.task-body`, el selector en `.new-task-meta` y los errores por campo. `edit.plush.html` no cambia.

**CSS**: estilos de `.task-body`, `.task-meta`, `.priority-*`, `.new-task-meta` y `select`. La versión de la hoja pasa a `?v=4`.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261014000000_add_priority_to_tasks.{up,down}.fizz` | Columna `priority` |
| `models/task.go`, `models/task_test.go` | Prioridad, validación, valor por defecto y tests |
| `actions/tasks.go`, `actions/tasks_test.go` | Alta con prioridad y tests de alta, índice y edición |
| `templates/tasks/index.plush.html` | Selector y etiqueta |
| `templates/application.plush.html`, `public/assets/app.css` | Estilos |

## Implementación paso a paso

1. Crear la migración.
2. Ampliar el modelo.
3. Ampliar el índice y la creación en las acciones.
4. Actualizar `index.plush.html`.
5. Actualizar el CSS y la versión de la hoja.
6. Escribir los tests y validar.

## Tests y validación

Modelo:

- valor por defecto `media`;
- las cuatro prioridades válidas, con su etiqueta;
- prioridad inválida.

Acciones:

- creación con prioridad;
- creación sin prioridad → `media`;
- prioridad inválida → 422;
- el índice muestra la etiqueta y no hay campo de responsable;
- la edición no tiene selector;
- el `PUT` con `priority` la ignora, también cuando el valor es inválido.

Los tests existentes deben seguir pasando sin cambios.

```
gofmt -l .
go vet ./...
soda migrate -e test
go test -p 1 ./...
go build ./...
```

## Riesgos y recuperación

- **Migración**: en PostgreSQL ≥ 11, `ADD COLUMN ... NOT NULL DEFAULT` no reescribe la tabla. Para revertir, `soda migrate down` elimina la columna y la prioridad de todas las tareas.
- **Prioridad fija**: una tarea creada con la prioridad equivocada no se puede corregir desde la interfaz; hay que borrarla y crearla de nuevo. Es una consecuencia directa del requisito confirmado.
- **Historial local**: la migración `20261014000000` se aplicó primero con la columna `assignee` en las bases locales de desarrollo y de test. Se revirtió antes de reescribirla y nunca se publicó.
