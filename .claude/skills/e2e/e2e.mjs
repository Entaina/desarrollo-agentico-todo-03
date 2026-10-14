// Runner e2e: ejecuta los escenarios de scenarios/*.mjs contra la app en marcha.
// Variables: BASE_URL, E2E_HEADED=1, E2E_SLOWMO (ms), E2E_HOLD (ms), E2E_BROWSER,
// E2E_SHOTS (directorio), E2E_SCENARIO (nombres de archivo sin extensión, separados por comas).
import { chromium } from "playwright";
import { existsSync, readdirSync } from "node:fs";
import { join, dirname } from "node:path";
import { homedir } from "node:os";
import { fileURLToPath, pathToFileURL } from "node:url";

const base = (process.env.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const headed = process.env.E2E_HEADED === "1";
const slowMo = Number(process.env.E2E_SLOWMO || (headed ? 700 : 0));
const hold = Number(process.env.E2E_HOLD || (headed ? 3000 : 0));
const shots = process.env.E2E_SHOTS;
const only = (process.env.E2E_SCENARIO || "").split(",").map((s) => s.trim()).filter(Boolean);
const scenariosDir = join(dirname(fileURLToPath(import.meta.url)), "scenarios");

// Usa el navegador cacheado más reciente si el de la versión instalada de Playwright no está.
function cachedBrowser() {
  if (process.env.E2E_BROWSER) return process.env.E2E_BROWSER;
  const cache = join(homedir(), ".cache", "ms-playwright");
  if (!existsSync(cache)) return undefined;
  const [prefix, bin] = headed
    ? ["chromium-", join("chrome-linux64", "chrome")]
    : ["chromium_headless_shell-", join("chrome-headless-shell-linux64", "chrome-headless-shell")];
  const dirs = readdirSync(cache)
    .filter((d) => d.startsWith(prefix) && /^\d+$/.test(d.slice(prefix.length)))
    .sort((a, b) => Number(b.slice(prefix.length)) - Number(a.slice(prefix.length)));
  for (const d of dirs) {
    const path = join(cache, d, bin);
    if (existsSync(path)) return path;
  }
  return undefined;
}

async function launch() {
  const opts = { headless: !headed, slowMo };
  try {
    return await chromium.launch(opts);
  } catch (err) {
    const executablePath = cachedBrowser();
    if (!executablePath) throw err;
    console.log(`Navegador de Playwright no disponible; uso ${executablePath}`);
    return chromium.launch({ ...opts, executablePath });
  }
}

const files = existsSync(scenariosDir)
  ? readdirSync(scenariosDir).filter((f) => f.endsWith(".mjs")).sort()
  : [];
const selected = only.length ? files.filter((f) => only.includes(f.replace(/\.mjs$/, ""))) : files;
const missing = only.filter((n) => !files.includes(n + ".mjs"));
if (missing.length) {
  console.error(`Escenarios inexistentes: ${missing.join(", ")}`);
  process.exit(2);
}
if (!selected.length) {
  console.error(`No hay escenarios en ${scenariosDir}`);
  process.exit(2);
}

const browser = await launch();
const results = [];
for (const file of selected) {
  const name = file.replace(/\.mjs$/, "");
  const scenario = await import(pathToFileURL(join(scenariosDir, file)).href);
  const page = await browser.newPage();
  const cleanups = [];
  const ctx = {
    page,
    base,
    step: (msg) => console.log(`  - ${msg}`),
    shot: async (shotName) => shots && page.screenshot({ path: join(shots, `${name}-${shotName}.png`) }),
    // Registra una limpieza; se ejecuta al final del escenario aunque falle.
    cleanup: (fn) => cleanups.push(fn),
    unique: (prefix) => `${prefix} ${Date.now()}`,
  };

  console.log(`▶ ${name}: ${scenario.description || ""}`);
  try {
    await scenario.default(ctx);
    results.push({ name, ok: true });
    console.log(`✔ ${name}`);
  } catch (err) {
    results.push({ name, ok: false });
    console.log(`✘ ${name}: ${err.message}`);
    // La captura de fallo se guarda siempre, en E2E_SHOTS o en el directorio actual.
    await page.screenshot({ path: join(shots || process.cwd(), `${name}-fallo.png`) }).catch(() => {});
  } finally {
    for (const fn of cleanups.reverse()) {
      await fn().catch((err) => console.log(`  ! limpieza fallida: ${err.message}`));
    }
    if (hold) await page.waitForTimeout(hold);
    await page.close();
  }
}
await browser.close();

const failed = results.filter((r) => !r.ok);
console.log(`\n${results.length - failed.length}/${results.length} escenarios OK`);
console.log(failed.length ? "E2E FALLÓ" : "E2E OK");
process.exit(failed.length ? 1 : 0);
