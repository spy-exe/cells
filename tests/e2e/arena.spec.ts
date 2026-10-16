import { test, expect, type Page } from "@playwright/test";
import { decodeSnapshot, type Snapshot } from "../../packages/protocol/game";
function observe(page: Page, onSnapshot?: (snapshot: Snapshot) => void) {
  let current: Snapshot | null = null;
  page.on("websocket", (ws) =>
    ws.on("framereceived", (event) => {
      const p = event.payload;
      if (typeof p !== "string" && p[2] === 2) {
        current = decodeSnapshot(
          p.buffer.slice(
            p.byteOffset,
            p.byteOffset + p.byteLength,
          ) as ArrayBuffer,
        );
        onSnapshot?.(current);
      }
    }),
  );
  return () => current;
}
test("AI Arena oferece cinco perfis e bots que se movem no mundo real", async ({
  page,
  browser,
}, info) => {
  const current = observe(page),
    room = `ai-${Date.now().toString(36)}-${info.project.name}`;
  await page.goto(`/?room=${room}`);
  await page.getByLabel("Modo", { exact: true }).selectOption("arena");
  const profiles = page.getByLabel("Como as IAs vão jogar?");
  await expect(profiles.locator("option")).toHaveCount(5);
  await profiles.selectOption("aggressive");
  await expect(
    page.getByText("IAs táticas, com regras programadas.", { exact: false }),
  ).toBeVisible();
  await page.getByLabel("Como você quer aparecer?").fill("AI explorer");
  await page.getByRole("button", { name: "Entrar na arena" }).click();
  await expect
    .poll(() => current()?.leaderboard.filter((p) => p.bot).length)
    .toBeGreaterThanOrEqual(9);
  await expect(page.getByRole("img", { name: /Arena Cells/ })).toBeVisible();
  // A legitimate spectator follows the leader. Never assume that enemies are
  // inside a newly spawned human's view on a large production map.
  const context = await browser.newContext(),
    watch = await context.newPage();
  const origins = new Map<number, { x: number; y: number }>();
  let moved = false;
  const watched = observe(watch, (snapshot) => {
    for (const e of snapshot.entities) {
      if (!e.bot || e.kind !== 2) continue;
      const old = origins.get(e.id);
      if (old && Math.hypot(e.x - old.x, e.y - old.y) > 20) moved = true;
      if (!old) origins.set(e.id, { x: e.x, y: e.y });
    }
  });
  try {
    await watch.goto(`/?room=${room}`);
    await watch.getByLabel("Modo", { exact: true }).selectOption("arena");
    await watch.getByLabel("Como as IAs vão jogar?").selectOption("aggressive");
    await watch.getByRole("button", { name: "Só quero assistir" }).click();
    await expect
      .poll(() => watched()?.entities.filter((e) => e.bot).length ?? 0)
      .toBeGreaterThan(0);
    // Leaders can change while feeding, so accumulate actual movement in all
    // received visible frames rather than requiring the first leader to remain.
    await expect.poll(() => moved, { timeout: 10000 }).toBe(true);
    await page.screenshot({
      path: `artifacts/${process.env.E2E_BASE_URL?.startsWith("https://cells.malha.app") ? "public" : "qa"}-arena-${info.project.name}.png`,
    });
  } finally {
    await context.close();
  }
  await page.getByRole("button", { name: "Sair da partida" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Comece pequeno",
  );
});
