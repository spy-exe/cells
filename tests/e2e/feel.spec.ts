import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
const fixture = () => {
  const b = Buffer.from(
    readFileSync("packages/protocol/fixtures/snapshot-v1.hex", "utf8").trim(),
    "hex",
  );
  b.writeFloatLE(6000, 12);
  b.writeFloatLE(3000, 16);
  b.writeFloatLE(3000, 20);
  b.writeFloatLE(2200, 24);
  b.writeFloatLE(40000, 28);
  b.writeUInt16LE(0, 32);
  b.writeUInt16LE(0, 34);
  b.writeFloatLE(40000, 38);
  b.writeFloatLE(3000, 61);
  b.writeFloatLE(3000, 65);
  b.writeFloatLE(40000, 69);
  b.writeFloatLE(40000, 86);
  return b;
};
test("cenário visual: célula grande e W mantido sem esperar outro clique", async ({
  page,
}, info) => {
  // Controlled binary fixture tests rendering and input only; real multiplayer
  // and authoritative action outcomes are covered by Classic E2E and Go tests.
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  let ejectCommands = 0,
    highest = 0;
  await page.routeWebSocket("**/ws", (ws) => {
    ws.send(fixture());
    ws.onMessage((raw) => {
      const b = Buffer.from(raw);
      if (b[2] === 1) {
        const sequence = b.readUInt32LE(3);
        expect(sequence).toBeGreaterThan(highest);
        highest = sequence;
        if (b[15] & 2) ejectCommands++;
      }
    });
  });
  await page.goto(`/?room=feel-${Date.now().toString(36)}`);
  await page.getByLabel("Como você quer aparecer?").fill("Nitidez");
  await page.getByRole("button", { name: "Entrar na arena" }).click();
  await expect(page.getByRole("img", { name: /Arena Cells/ })).toBeVisible();
  await expect(page.getByTestId("mass")).toContainText("40.000");
  await page.keyboard.down("w");
  await expect
    .poll(() => ejectCommands, { timeout: 2000 })
    .toBeGreaterThanOrEqual(3);
  await page.keyboard.up("w");
  await page.waitForTimeout(300);
  const stopped = ejectCommands;
  await page.waitForTimeout(200);
  expect(ejectCommands).toBe(stopped);
  const button = page.getByRole("button", { name: "Ejetar massa" }),
    box = await button.boundingBox();
  expect(box).not.toBeNull();
  await page.mouse.move(box!.x + box!.width / 2, box!.y + box!.height / 2);
  await page.mouse.down();
  await expect.poll(() => ejectCommands - stopped).toBeGreaterThanOrEqual(3);
  await page.mouse.up();
  await page.screenshot({
    path: `artifacts/fixture-large-cell-${info.project.name}.png`,
  });
  expect(errors).toEqual([]);
});
