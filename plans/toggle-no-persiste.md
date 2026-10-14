# Bug: completar o reabrir una tarea desde la lista no se guarda

## Encargo y resultado esperado

Encargo: «Completar o reabrir una tarea desde la lista no se guarda: al marcar una tarea como completada el cambio se ve en pantalla, pero al refrescar la página se pierde.»

Resultado esperado:

1. Al pulsar el círculo de una tarea, la fila cambia al instante (comportamiento actual, se conserva) **y** el cambio queda guardado: tras recargar, la tarea mantiene el estado elegido.
2. Si el servidor no confirma el cambio (error HTTP o de red), la fila vuelve a su estado anterior y se avisa al usuario, en lugar de mostrar un estado que no está guardado.
3. Sin JavaScript, el formulario sigue funcionando como hoy (POST con `_method=PUT` y redirección 303).

## Evidencia y causa raíz

Estado de partida: `origin/main` en `ff94db6`.

Hechos observados:

- `actions/app.go` registra solo `app.PUT("/tasks/{task_id}/toggle", TasksToggle)` y aplica `csrf.New` a todas las peticiones.
- El formulario de cada fila (`templates/tasks/index.plush.html`, atributo `data-task-toggle`) envía `POST` con los campos ocultos `_method=PUT` y `authenticity_token`. Buffalo convierte `_method` en el método real y el middleware CSRF valida el campo `authenticity_token`.
- `public/assets/app.js` intercepta el `submit`, hace `event.preventDefault()`, cambia la fila y termina con `fetch(form.action);`: un `GET` sin cuerpo, sin `_method` ni token. No existe ruta GET para `/tasks/{id}/toggle`, así que la petición no llega a `TasksToggle` y la base de datos no cambia. Se introdujo en `5fdd6c3 feat(ui): improve task completion feedback`.
- El `fetch` no comprueba la respuesta, por lo que el fallo pasa desapercibido y la interfaz muestra un estado falso.
- `TasksToggle` funciona: `Test_TasksToggle_CompletesAndReopens` lo cubre con `PUT`.

Causa raíz: el envío asíncrono no reproduce el contrato del formulario (método `POST` con `_method=PUT` y `authenticity_token`) y no trata los fallos.

Incertidumbres:

- No hay infraestructura de tests de JavaScript (ni `package.json` ni runner). El fallo es de cliente: ningún test Go puede fallar antes del arreglo, porque el servidor ya acepta correctamente el formulario. La regresión se demuestra con la reproducción manual; el test Go propuesto protege el contrato del servidor del que depende el JavaScript.
- En los tests Go el middleware CSRF está desactivado (`GO_ENV=test`), así que la validación del token solo se comprueba en el navegador.

Reproducción mínima (antes del arreglo):

1. Arrancar la app y abrir la lista.
2. Pulsar el círculo de una tarea pendiente: se marca como completada.
3. Recargar: vuelve a aparecer pendiente. En la pestaña de red, `GET /tasks/<id>/toggle` no devuelve éxito y la fila `completed` no cambia.

## Alcance

Incluido:

- Enviar el toggle asíncrono como el formulario original (POST con los campos del formulario, incluidos `_method` y `authenticity_token`, y cookies de la sesión).
- Revertir la fila y avisar si la respuesta no es correcta o la petición falla.
- Un test Go del contrato `POST` + `_method=PUT` en el que se apoya el JavaScript.

Excluido:

- Cambios en rutas, en `TasksToggle`, en el middleware CSRF o en el HTML del formulario.
- Introducir un runner de tests de JavaScript o nuevas dependencias.
- Rediseñar el aspecto de la fila o los mensajes de flash.

## Archivos relevantes

- `public/assets/app.js`: intercepta el formulario de toggle y actualiza la fila; contiene el `fetch` defectuoso.
- `templates/tasks/index.plush.html`: formulario `data-task-toggle` con `_method` y `authenticity_token` (no cambia; define el contrato).
- `actions/app.go`: ruta `PUT /tasks/{task_id}/toggle` y middleware CSRF (no cambia).
- `actions/tasks.go`: `TasksToggle` (no cambia).
- `actions/tasks_test.go`: tests de acciones; se añade el test del contrato.

## Implementación paso a paso

1. En `public/assets/app.js`, extrae a una función `setTaskCompleted(task, completed)` la actualización actual de la fila: clase `task-completed`, etiqueta `.toggle-label`, `aria-pressed`, `aria-label`, `title` y el indicador `.task-status`. Localiza el botón con `task.querySelector(".task-toggle-button")` para que la función solo dependa de la fila.
2. En el manejador de `submit`, calcula `completed = !task.classList.contains("task-completed")` y llama a `setTaskCompleted(task, completed)` (la respuesta inmediata se conserva).
3. Sustituye `fetch(form.action);` por:
   ```js
   fetch(form.action, {
     method: "POST",
     body: new URLSearchParams(new FormData(form)),
     credentials: "same-origin",
   })
   ```
   seguido de un `.then` que lance un error si `!response.ok` y un `.catch` que llame a `setTaskCompleted(task, !completed)` y muestre `window.alert("No se pudo guardar el cambio. Inténtalo de nuevo.")`. Añade un comentario breve que explique que se envía como el formulario original y que se deshace si el servidor no confirma. `fetch` sigue la redirección 303 a `/`, que responde 200, así que un toggle correcto da `response.ok`.
4. Conserva intacto el bloque de `data-confirm` al principio del manejador.
5. En `actions/tasks_test.go`, añade `Test_TasksToggle_FormMethodOverride`: crea una tarea, envía `as.HTML("/tasks/%s/toggle", task.ID).Post(url.Values{"_method": {"PUT"}})`, comprueba código 303 y que tras `as.DB.Reload(task)` la tarea está completada.

## Tests y validación

- `gofmt -l .` → sin salida.
- `go vet ./...`
- `go test -p 1 ./actions` (requiere PostgreSQL y migraciones de test según `database.yml` y `.github/workflows/ci.yaml`); si no hay base de datos disponible, indícalo en el reporte en lugar de darlo por pasado.
- `go build ./...`
- `node --check public/assets/app.js` si `node` está disponible, para verificar la sintaxis del JavaScript.
- Verificación manual en navegador (no automatizable en este repositorio): completar y reabrir una tarea, recargar y comprobar que el estado persiste; en la pestaña de red, la petición es `POST /tasks/<id>/toggle` y termina en 200 tras la redirección. Simular un fallo (por ejemplo, borrar el campo `authenticity_token` en las herramientas de desarrollo) debe revertir la fila y mostrar el aviso.

## Riesgos y recuperación

- Riesgo: si el token CSRF de la página caduca, el servidor responde 403; con el arreglo la fila se revierte y se avisa, que es el comportamiento deseado.
- Riesgo: navegadores sin `fetch`/`URLSearchParams` (no soportados por el resto del proyecto de todos modos).
- Recuperación: el cambio se limita a `public/assets/app.js` y a un test; revertir el commit restaura el comportamiento anterior sin migraciones ni cambios de datos.
