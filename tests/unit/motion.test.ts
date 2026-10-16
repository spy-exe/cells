import { it, expect } from "vitest";
import { MotionTimeline, virusPoints } from "../../apps/web/src/motion";
import type { Snapshot } from "../../packages/protocol/game";
const sample = (tick: number, x: number, received: number) =>
  ({
    tick,
    self: 7,
    alive: true,
    teams: false,
    size: 6000,
    x,
    y: 1000,
    view: 900,
    mass: 36,
    splitIn: 0,
    ejectIn: 0,
    respawnIn: 0,
    teamMass: [0, 0, 0],
    received,
    leaderboard: [],
    entities: [
      {
        id: 8,
        owner: 7,
        kind: 2,
        x,
        y: 1000,
        mass: 36,
        name: "A",
        bot: false,
        team: 0,
      },
    ],
  }) satisfies Snapshot;
it("keeps a steady cell moving under jitter without reversing or unbounded prediction", () => {
  const timeline = new MotionTimeline();
  const packets = [
    sample(0, 1000, 1000),
    sample(3, 1030, 1105),
    sample(6, 1060, 1225),
    sample(9, 1090, 1302),
    sample(12, 1120, 1410),
  ];
  let next = 0,
    last = 1000;
  for (let now = 1000; now < 1600; now += 8) {
    while (next < packets.length && packets[next].received <= now)
      timeline.push(packets[next++]);
    const p = timeline.entity(8, now);
    if (!p) continue;
    expect(p.x).toBeGreaterThanOrEqual(last - 0.001);
    expect(p.x - last).toBeLessThan(12);
    last = p.x;
  }
  expect(last).toBeLessThanOrEqual(1140.01);
  expect(last).toBeGreaterThan(1120);
});
it("resets the display clock when the server changes life or restarts", () => {
  const t = new MotionTimeline();
  t.push(sample(60, 1500, 1000));
  t.entity(8, 1050);
  t.push(sample(1, 500, 1100));
  expect(t.entity(8, 1100)?.x).toBe(500);
});
it("draws 48 actual virus tips with alternating inner vertices", () => {
  const p = virusPoints();
  expect(p).toHaveLength(192);
  expect(p[0]).toBe(127);
  expect(p[1]).toBe(64);
});
