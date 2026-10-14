---
name: e2e
description: Ejecuta con Playwright los escenarios e2e de la app TODO contra la aplicación en marcha, en modo headless o visible (--headed)
allowed-tools: Bash, Read
---

# Ejecuta las pruebas e2e

`e2e.mjs` (en el directorio de esta skill) es un runner que ejecuta cada escenario de `scenarios/*.mjs` contra la app real, en una página nueva por escenario, y ejecuta sus limpiezas aunque el escenario falle.

`$ARGUMENTS` admite:

- `--headed`: abre una ventana de Chromium visible (por defecto, headless).
- `--slowmo <ms>`: pausa entre acciones (por defecto 700 con `--headed`, 0 sin él).
- `--hold <ms>`: tiempo que cada página queda abierta al terminar su escenario (por defecto 3000 con `--headed`).
- `--shots`: guarda capturas en `tmp/e2e/`. Con un fallo siempre se intenta una captura `<escenario>-fallo.png`.
- `--scenario <a,b>`: ejecuta solo esos escenarios (nombre de archivo sin `.mjs`). Por defecto, todos.
- `--rebuild`: recompila y reinicia la app para probar el código actual del árbol de trabajo.
- `--url <url>`: URL de la app (por defecto `http://localhost:3000`).

No modifiques el código de la aplicación, los tests Go ni las dependencias del repositorio. Todo lo que instala o genera esta skill va a `tmp/e2e/`, que está en `.gitignore`.

## Escenarios

Cada escenario es un módulo ES en `scenarios/<nombre>.mjs`, con el nombre kebab-case de la funcionalidad o del bug que cubre:

```js
import assert from "node:assert";

export const description = "qué comprueba, en una línea";

export default async function ({ page, base, step, shot, cleanup, unique }) {
  const title = unique("E2E algo");        // datos únicos por ejecución
  step("Paso legible");                    // traza numerada en la salida
  await page.goto(base + "/");
  // ... acciones y aserciones con node:assert ...
  cleanup(async () => { /* deshacer lo creado */ });
  await shot("estado");                    // solo guarda con --shots
}
```

Reglas:

- Crea los datos que necesites con `unique(...)` y regístrales una limpieza con `cleanup` justo después de crearlos. No dependas de datos creados por otro escenario.
- No borres ni modifiques datos que el escenario no haya creado. Puedes leer los del seed.
- Usa selectores estables del HTML real (ids, clases, `title`, `aria-*`) y comprueba texto visible para el usuario.
- Importa solo `node:*`; Playwright llega por `page`.

## 1. App disponible

1. Sin `--rebuild`, si `curl -s -o /dev/null -w "%{http_code}" <url>/healthz` devuelve `200`, reutiliza la app y no la reinicies. Advierte en el resultado que puede no reflejar el código actual.
2. Con `--rebuild` y URL local:
   - Si el puerto `3000` lo ocupa un proceso cuyo comando es `tmp/e2e/todo` (compruébalo con `ss -ltnp` y `ps`), finalízalo. Si lo ocupa cualquier otro proceso, informa del conflicto con su PID y comando y detente: no lo finalices.
   - Continúa con el arranque del punto 3.
3. Si no responde, o tras el paso anterior, y la URL es local:
   - Ejecuta `docker compose up -d db` y espera a que `db` figure como saludable en `docker compose ps`.
   - Compila con `go build -o tmp/e2e/todo ./cmd/app` y arráncala en segundo plano con `ADDR=127.0.0.1 PORT=3000`, guardando la salida en `tmp/e2e/app.log`. Al arrancar aplica las migraciones pendientes en `todo_development`.
   - Espera hasta que `/healthz` devuelva `200`. Si no llega, muestra el final de `tmp/e2e/app.log` y detente.
4. Recuerda si arrancaste tú la app para indicarlo en el resultado.

## 2. Playwright

1. Crea `tmp/e2e/` si no existe. Si no contiene `node_modules/playwright`, ejecuta allí `npm init -y` y `npm install playwright@1.64.1`.
2. En cada ejecución, sincroniza la versión de la skill: copia `e2e.mjs` a `tmp/e2e/e2e.mjs` y reemplaza `tmp/e2e/scenarios/` por una copia de `scenarios/`.
3. No ejecutes `npx playwright install` por iniciativa propia. El runner usa el navegador de la versión instalada y, si no existe, el Chromium más reciente en `~/.cache/ms-playwright`. Solo si no hay ninguno, explica el bloqueo y propón ese comando.

## 3. Ejecutar

Desde `tmp/e2e/`, ejecuta `node e2e.mjs` con las variables según los argumentos:

| Argumento | Variable |
|---|---|
| `--headed` | `E2E_HEADED=1` |
| `--slowmo <ms>` | `E2E_SLOWMO=<ms>` |
| `--hold <ms>` | `E2E_HOLD=<ms>` |
| `--shots` | `E2E_SHOTS=<ruta absoluta de tmp/e2e>` |
| `--scenario <a,b>` | `E2E_SCENARIO=<a,b>` |
| `--url <url>` | `BASE_URL=<url>` |

Con `--headed`, comprueba antes que `DISPLAY` o `WAYLAND_DISPLAY` están definidos. Si no lo están, explica que no hay pantalla y ofrece ejecutar en headless.

Usa un timeout amplio del comando (al menos 3 minutos con `--headed`). El runner termina con código `0` si todos los escenarios pasan, `1` si alguno falla y `2` si no hay escenarios o se pidió uno inexistente.

## Resultado

Informa de:

- `E2E OK` o `E2E FALLÓ`, con cada escenario y, en los fallidos, el último paso impreso y el mensaje de la aserción;
- si la app se reutilizó, se reconstruyó o la arrancaste tú (y cómo pararla: `pkill -f tmp/e2e/todo`);
- el navegador utilizado si el runner avisó de un fallback;
- la ruta de las capturas, si existen (muéstralas con `Read` cuando el usuario quiera verlas o haya un fallo).
