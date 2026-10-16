import type { Entity, Snapshot } from "../../../packages/protocol/game";
interface Sample {
  snapshot: Snapshot;
  entities: Map<number, Entity>;
}
export class MotionTimeline {
  private samples: Sample[] = [];
  private offset = 0;
  private displayTime = -Infinity;
  private lastNow: number | null = null;
  private cachedAt = -Infinity;
  private cached: { a: Sample; b: Sample; t: number } | null = null;
  push(snapshot: Snapshot) {
    this.cachedAt = -Infinity;
    const latest = this.samples.at(-1)?.snapshot;
    const measured = snapshot.received - (snapshot.tick * 1000) / 30;
    if (
      !latest ||
      snapshot.tick < latest.tick ||
      snapshot.self !== latest.self ||
      snapshot.alive !== latest.alive
    ) {
      this.samples = [];
      this.offset = measured;
      this.displayTime = -Infinity;
      this.lastNow = null;
    } else this.offset += (measured - this.offset) * 0.08;
    this.samples.push({
      snapshot,
      entities: new Map(snapshot.entities.map((e) => [e.id, e])),
    });
    if (this.samples.length > 6) this.samples.shift();
  }
  private pair(now: number) {
    if (now === this.cachedAt) return this.cached;
    this.cachedAt = now;
    if (!this.samples.length) return null;
    const latest = this.samples.at(-1)!;
    const desired = Math.min(
      now - this.offset - 65,
      (latest.snapshot.tick * 1000) / 30 + 60,
    );
    const dt =
      this.lastNow === null
        ? 0
        : Math.max(0, Math.min(100, now - this.lastNow));
    this.lastNow = now;
    if (!Number.isFinite(this.displayTime) || this.samples.length === 1)
      this.displayTime = (latest.snapshot.tick * 1000) / 30;
    else {
      const rate = Math.max(
        0.85,
        Math.min(1.18, 1 + (desired - this.displayTime) / 300),
      );
      this.displayTime = Math.min(
        (latest.snapshot.tick * 1000) / 30 + 60,
        this.displayTime + dt * rate,
      );
    }
    let a = this.samples[0],
      b = a;
    for (const sample of this.samples) {
      if ((sample.snapshot.tick * 1000) / 30 <= this.displayTime) a = sample;
      else {
        b = sample;
        break;
      }
      b = sample;
    }
    if (a === b && this.samples.length > 1 && a === latest) {
      a = this.samples.at(-2)!;
      b = latest;
    }
    const span = ((b.snapshot.tick - a.snapshot.tick) * 1000) / 30;
    const t =
      span > 0
        ? Math.max(0, (this.displayTime - (a.snapshot.tick * 1000) / 30) / span)
        : 1;
    this.cached = { a, b, t };
    return this.cached;
  }
  entity(id: number, now: number): Entity | null {
    const p = this.pair(now);
    if (!p) return null;
    const fallback = this.samples.at(-1)?.entities.get(id),
      a = p.a.entities.get(id),
      b = p.b.entities.get(id);
    if (!a || !b) return fallback ?? b ?? a ?? null;
    return {
      ...b,
      x: a.x + (b.x - a.x) * p.t,
      y: a.y + (b.y - a.y) * p.t,
      mass: a.mass + (b.mass - a.mass) * Math.min(1, p.t),
    };
  }
  center(now: number) {
    const p = this.pair(now);
    if (!p) return { x: 0, y: 0 };
    return {
      x: p.a.snapshot.x + (p.b.snapshot.x - p.a.snapshot.x) * p.t,
      y: p.a.snapshot.y + (p.b.snapshot.y - p.a.snapshot.y) * p.t,
    };
  }
}
export function virusPoints(tips = 48): number[] {
  const p: number[] = [];
  for (let i = 0; i < tips * 2; i++) {
    const a = (i * Math.PI) / tips,
      r = i % 2 ? 54 : 63;
    p.push(64 + Math.cos(a) * r, 64 + Math.sin(a) * r);
  }
  return p;
}
