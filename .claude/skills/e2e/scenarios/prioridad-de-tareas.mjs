// Escenario: la prioridad se elige al crear, se muestra en la lista y se edita.
import assert from "node:assert";

export const description = "crear con prioridad Alta, seed en Media, editar a Baja";

export default async function ({ page, base, step, shot, cleanup, unique }) {
  const title = unique("E2E prioridad");
  const row = () => page.locator(".task", { hasText: title });
  const priorityOf = (text) => page.locator(".task", { hasText: text }).locator(".task-priority").innerText();

  step("El alta propone prioridad Media");
  await page.goto(base + "/");
  assert.equal(await page.inputValue("#task-priority"), "medium");

  step("Crear una tarea con prioridad Alta");
  await page.fill("#task-title", title);
  await page.selectOption("#task-priority", "high");
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
  assert.match(await priorityOf(title), /Prioridad: Alta/);

  step("Las tareas del seed tienen prioridad Media");
  for (const t of ["Preparar la reunión del lunes", "Enviar el informe mensual"]) {
    if (await page.locator(".task", { hasText: t }).count()) {
      assert.match(await priorityOf(t), /Prioridad: Media/);
    }
  }
  await shot("lista");

  step("Editar: preselecciona Alta y cambia a Baja");
  await row().locator('a[title="Editar tarea"]').click();
  assert.equal(await page.inputValue("#task-priority"), "high");
  await page.selectOption("#task-priority", "low");
  await page.click("text=Guardar cambios");
  await page.waitForURL(base + "/");
  assert.match(await priorityOf(title), /Prioridad: Baja/);
  await shot("lista-editada");
}
