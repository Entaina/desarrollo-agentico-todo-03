// Escenario: completar y reabrir una tarea desde la lista se guarda en el servidor.
import assert from "node:assert";

export const description = "completar y reabrir persisten tras recargar";

export default async function ({ page, base, step, shot, cleanup, unique }) {
  const title = unique("E2E completar");
  const row = () => page.locator(".task", { hasText: title });
  const isCompleted = () => row().evaluate((el) => el.classList.contains("task-completed"));

  const toggleErrors = [];
  page.on("response", (res) => {
    if (res.url().includes("/toggle") && res.status() >= 400) {
      toggleErrors.push(`${res.request().method()} ${res.url()} -> ${res.status()}`);
    }
  });

  // Espera a que termine la petición del toggle y a que el botón vuelva a estar activo.
  const toggle = async () => {
    const done = page.waitForResponse((res) => res.url().endsWith("/"), { timeout: 5000 }).catch(() => {});
    await row().locator(".task-toggle-button").click();
    await done;
    await page.waitForFunction(
      (t) => {
        const li = [...document.querySelectorAll(".task")].find((el) => el.textContent.includes(t));
        return li && !li.querySelector(".task-toggle-button").disabled;
      },
      title,
      { timeout: 5000 }
    );
  };

  step("Crear una tarea pendiente");
  await page.goto(base + "/");
  await page.fill("#task-title", title);
  await page.click("text=Añadir");
  await page.waitForURL(base + "/");
  cleanup(async () => {
    await page.goto(base + "/");
    if (await row().count()) {
      page.once("dialog", (d) => d.accept());
      await row().locator('button[title="Eliminar tarea"]').click();
      await page.waitForLoadState();
    }
  });
  assert.equal(await isCompleted(), false);

  step("Completar y recargar: sigue completada");
  await toggle();
  await page.reload();
  assert.equal(await isCompleted(), true, "la tarea completada vuelve a pendiente al recargar");
  assert.equal(await row().locator(".task-toggle-button").getAttribute("aria-pressed"), "true");
  await shot("completada");

  step("Reabrir y recargar: sigue pendiente");
  await toggle();
  await page.reload();
  assert.equal(await isCompleted(), false, "la tarea reabierta vuelve a completada al recargar");

  step("El toggle no recibe errores del servidor");
  assert.deepEqual(toggleErrors, []);
}
