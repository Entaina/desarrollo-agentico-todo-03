# Bug: Completar o reabrir una tarea desde la lista no se guarda

## Encargo y resultado esperado

Encargo literal: «seems when edit and save a task, not have persistence».

Interpretación (inferencia del triage, verificada solo en parte): el cambio que no persiste es **marcar una tarea como completada o reabrirla desde la lista**. Editar el título o la prioridad desde `/tasks/{id}/edit` sí persiste (comprobado). Si el usuario se refería a otra acción, este plan no la cubre, pero el bug descrito aquí existe igualmente.

Resultado esperado:
1. Al pulsar el botón de completar/reabrir, la UI cambia al instante (feedback actual, introducido en `5fdd6c3`) **y** el cambio se guarda en la base de datos.
2. Tras recargar `/`, la tarea conserva el estado elegido.
3. Si el servidor rechaza o falla la petición, la UI vuelve al estado anterior para no mostrar algo que no se guardó.
4. Sin JavaScript, el formulario sigue funcionando como hoy (`PUT` por `_method` y redirección 303 a `/`).

## Evidencia y causa raíz

Hechos observados (reproducción con Playwright contra la app local, `GO_ENV=development`):
- Crear tarea → editar título → guardar → recargar: el título nuevo persiste.
- Pulsar completar: la `li.task` recibe `task-completed`; al recargar, la tarea vuelve a estar pendiente.
- La red muestra `GET /tasks/{id}/toggle` → **404**.

Causa raíz (leída en el código):
- `public/assets/app.js`: para formularios `[data-task-toggle]`, el handler hace `event.preventDefault()`, actualiza el DOM y termina con `fetch(form.action);`. Sin opciones, `fetch` hace un `GET` sin cuerpo.
- `actions/app.go`: la ruta es `app.PUT("/tasks/{task_id}/toggle", TasksToggle)`. No existe `GET` para esa ruta → 404, y `TasksToggle` nunca se ejecuta.
- `templates/tasks/index.plush.html`: el formulario es `method="POST"` con campos ocultos `_method=PUT` y `authenticity_token`. Buffalo convierte `POST` + `_method` en `PUT` (`MethodOverride`, por defecto en `buffalo v1.1.4`), y el middleware CSRF (`github.com/gobuffalo/middleware/csrf`) acepta el token del campo `authenticity_token` o de la cabecera `X-CSRF-Token`. El `fetch` actual no envía ninguno de los dos.
- `git log -S"fetch(form.action)"` → el fallo entra en `5fdd6c3 feat(ui): improve task completion feedback`.
- Por qué no lo detectaron los tests: `Test_TasksToggle_CompletesAndReopens` (`actions/tasks_test.go`) hace el `PUT` directamente en el servidor; ningún test ejecuta el JavaScript.

Incertidumbres:
- Que «edit and save» signifique completar/reabrir es una inferencia.
- No hay evidencia de otro camino sin persistencia; editar título y prioridad están cubiertos por tests Go y por el escenario e2e `prioridad-de-tareas`.

Reproducción mínima:
1. App en marcha, abrir `/`, crear una tarea pendiente.
2. Pulsar su botón de completar: se ve como completada.
3. Recargar: aparece pendiente. En la red, `GET …/toggle` → 404.

## Alcance

Incluido:
- Corregir la petición del handler de `[data-task-toggle]` en `public/assets/app.js` para que envíe el formulario real (`POST` con `_method=PUT` y `authenticity_token`).
- Revertir la UI si la respuesta no es correcta o la petición falla.
- Subir la versión de caché de `app.js` en `templates/application.plush.html` (`?v=2` → `?v=3`), para que los navegadores no sigan usando el JS roto.
- Test de regresión e2e en el navegador (escenario `persistencia-al-completar-tarea`).

Excluido:
- Cambios en `actions/tasks.go`, rutas, modelo o migraciones: `TasksToggle` funciona (lo prueba `Test_TasksToggle_CompletesAndReopens`).
- Responder JSON o `204` en el toggle, o cualquier otra API nueva.
- Rediseñar el feedback visual, el contador o los flashes.
- Cambios en el formulario de edición.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `public/assets/app.js` | Handler de `submit` de `[data-task-toggle]`: actualización optimista y petición al servidor (aquí está el fallo). |
| `templates/application.plush.html` | Carga `app.js?v=2`; subir la versión. **Ya tiene un cambio ajeno sin commitear** (`app.css?v=3` → `?v=4`, de la feature de prioridad). |
| `templates/tasks/index.plush.html` | Formulario de toggle (`POST`, `_method=PUT`, `authenticity_token`, `data-task-toggle`). Solo lectura. |
| `actions/app.go` | Ruta `PUT /tasks/{task_id}/toggle`. Solo lectura. |
| `actions/tasks.go` | `TasksToggle`: invierte `Completed` y redirige 303 a `/`. Solo lectura. |
| `actions/tasks_test.go` | `Test_TasksToggle_*` cubren el servidor. Sin cambios. |
| `.claude/skills/e2e/scenarios/persistencia-al-completar-tarea.mjs` | Escenario de regresión en el navegador (nuevo). |

## Implementación paso a paso

1. Comprobar `git status --short --branch` y confirmar que `public/assets/app.js` sigue terminando el handler con `fetch(form.action);`.
2. En `public/assets/app.js`, extraer la actualización del DOM del toggle a una función local (p. ej. `setCompleted(task, form, completed)`) que aplique clase, `aria-pressed`, `aria-label`, `title`, la etiqueta oculta y el `span.task-status`. Así sirve también para revertir. Mantener el estilo actual del archivo (ES5, `var`, `function`).
3. Sustituir `fetch(form.action);` por el envío del formulario real:
   ```js
   fetch(form.action, {
     method: "POST",
     body: new FormData(form),
     credentials: "same-origin",
     headers: { Accept: "text/html" }
   }).then(function (response) {
     if (!response.ok) {
       throw new Error("HTTP " + response.status);
     }
   }).catch(function () {
     setCompleted(task, form, !completed);
   });
   ```
   - `FormData(form)` incluye `_method=PUT` y `authenticity_token`, que Buffalo ya entiende.
   - `fetch` sigue la redirección 303 a `/` y obtiene un 200: eso cuenta como éxito.
   - 403 (CSRF), 404 o 500, o un error de red, revierten la UI.
4. Desactivar el botón mientras la petición está en curso (`button.disabled = true` y reactivarlo en `finally`/al final de ambas ramas). Esto evita peticiones superpuestas por doble clic que dejarían UI y servidor desincronizados.
5. En `templates/application.plush.html`, cambiar `app.js?v=2` por `app.js?v=3`.
6. Crear el escenario e2e `.claude/skills/e2e/scenarios/persistencia-al-completar-tarea.mjs` (ver tests).
7. Ejecutar la validación.

## Tests y validación

Test de regresión (debe fallar antes del arreglo y pasar después). Es un escenario e2e porque el fallo vive en el JavaScript del navegador y ningún test Go lo ejecuta:
- Crear una tarea con título único (con `cleanup`).
- Pulsar completar → recargar → la `li.task` tiene `task-completed` y el botón `aria-pressed="true"`.
- Pulsar reabrir → recargar → ya no tiene `task-completed`.
- Comprobar que la petición al toggle no devuelve 4xx/5xx: escuchar `page.on("response")` para URLs que contengan `/toggle`.
- Antes del arreglo falla en «recargar → completada» (vuelve a pendiente, y la petición es `GET` → 404).

Tests Go: sin cambios. `Test_TasksToggle_CompletesAndReopens`, `Test_TasksToggle_NotFound` y `Test_TasksToggle_InvalidID` deben seguir pasando.

Validación:
- `gofmt -l .` (sin salida) y `go vet ./...`, aunque no cambie código Go, por si hubiera regresiones del árbol.
- `GO_ENV=test go test -p 1 ./...`: requiere PostgreSQL y la base `todo_test` migrada. CI usa `GO_ENV=test` y `soda migrate -e test` (`.github/workflows/ci.yaml`). Sin `GO_ENV=test`, el CSRF devuelve 403 en los tests de acciones.
- `go build ./...`.
- `/e2e --rebuild`: todos los escenarios, incluidos `prioridad-de-tareas` como regresión y el nuevo.
- Manual opcional: en el navegador, completar una tarea, recargar y comprobar que sigue completada.

## Riesgos y recuperación

- **Interpretación del encargo:** si el usuario se refería a otra acción, este arreglo no la resuelve. El resumen debe pedirle confirmación.
- **Caché del navegador:** sin subir `?v=` se seguiría sirviendo el JS roto. Se sube a `v=3`.
- **Solapamiento con cambios ajenos:** `templates/application.plush.html` ya tiene modificado el `?v=` del CSS (feature de prioridad sin commitear). El cambio de este bug cae en el mismo archivo, así que no puede commitearse por separado sin `git add -p`. Esto impedirá a `/adw` hacer commit y PR automáticamente. Se resuelve commiteando antes la feature de prioridad.
- **Revertir tras redirección:** `fetch` sigue el 303 y el `GET /` posterior puede fallar por motivos ajenos al toggle. En ese caso la UI se revertiría aunque el cambio sí se guardó. Riesgo bajo y visible al recargar; se acepta.
- **Doble clic:** mitigado desactivando el botón durante la petición.
- **Recuperación:** el cambio solo afecta a `app.js` y a un parámetro de caché; revertir el commit restaura el comportamiento anterior. Sin datos ni migraciones implicados.
