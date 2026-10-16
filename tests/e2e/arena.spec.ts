import { test, expect } from "@playwright/test";
import { decodeSnapshot, type Snapshot } from "../../packages/protocol/game";

test("AI Arena oferece cinco perfis e bots que se movem no mundo real", async ({
  page,
}, info) => {
  let current: Snapshot | null = null;
  page.on("websocket", (ws) =>
    ws.on("framereceived", (event) => {
      const p = event.payload;
      if (typeof p !== "string" && p[2] === 2)
        current = decodeSnapshot(
          p.buffer.slice(
            p.byteOffset,
            p.byteOffset + p.byteLength,
          ) as ArrayBuffer,
        );
    }),
  );
  await page.goto(`/?room=ai-${Date.now().toString(36)}-${info.project.name}`);
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
    .poll(() => current?.leaderboard.filter((p) => p.bot).length)
    .toBe(8);
  await expect(page.getByRole("img", { name: /Arena Cells/ })).toBeVisible();
  const initial = new Map(
    current!.entities
      .filter((e) => e.bot)
      .map((e) => [e.id, { x: e.x, y: e.y }]),
  );
  await expect
    .poll(
      () =>
        current?.entities.some((e) => {
          const old = initial.get(e.id);
          return old && Math.hypot(e.x - old.x, e.y - old.y) > 20;
        }),
      { timeout: 10000 },
    )
    .toBe(true);
  await page.screenshot({
    path: `artifacts/${process.env.E2E_BASE_URL?.startsWith("https://cells.malha.app") ? "public" : "qa"}-arena-${info.project.name}.png`,
  });
  await page.getByRole("button", { name: "Sair da partida" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Comece pequeno",
  );
});
