import type { Snapshot } from "../../../packages/protocol/game";

/** Bodies and their own captions share a depth; a larger body covers both. */
export function entityDepth(kind: number, mass: number, label = false) {
  return kind === 2 ? 10 + Math.log1p(mass) * 1000 + (label ? .01 : 0) : kind === 4 ? 2 : kind === 3 ? 1 : 0;
}

export function membraneGeometry(segments = 96) {
  const vertices = new Float32Array((segments + 2) * 2);
  const indices = new Uint32Array(segments * 3);
  const uvs = new Float32Array(vertices.length).fill(.5);
  for (let i = 0; i < segments; i++) indices.set([0, i + 1, i + 2], i * 3);
  updateMembrane(vertices, 0, 0, 10000, 0, true);
  return { vertices, indices, uvs };
}

export function updateMembrane(vertices: Float32Array, seed: number, now: number, age: number, direction: number, reduced: boolean) {
  const segments = vertices.length / 2 - 2, t = now / 1000, phase = seed * 2.399963;
  const impulse = reduced ? 0 : Math.exp(-Math.max(0, age) / 140) * .018;
  for (let i = 0; i < segments; i++) {
    const angle = i / segments * Math.PI * 2;
    const wave = reduced ? 0 : Math.sin(angle * 3 + t * 1.3 + phase) * .008 + Math.sin(angle * 5 - t * .8 + phase) * .005;
    const local = impulse * Math.cos(angle - direction) * Math.sin(age / 70 - angle * 2);
    const radius = 1 + wave + local;
    vertices[2 + i * 2] = Math.cos(angle) * radius;
    vertices[3 + i * 2] = Math.sin(angle) * radius;
  }
  vertices[vertices.length - 2] = vertices[2]; vertices[vertices.length - 1] = vertices[3];
  let area = 0;
  for (let i = 2; i < vertices.length - 2; i += 2) area += vertices[i] * vertices[i + 3] - vertices[i + 2] * vertices[i + 1];
  const scale = Math.sqrt(Math.PI / (area / 2));
  for (let i = 2; i < vertices.length; i++) vertices[i] *= scale;
}

export function absorptionPose(x: number, y: number, targetX: number, targetY: number, progress: number) {
  const t = Math.max(0, Math.min(1, progress)), ease = 1 - (1 - t) ** 3;
  return { x: x + (targetX - x) * ease, y: y + (targetY - y) * ease, scale: 1 - t, alpha: 1 - t };
}

/** Cosmetic inference requires local disappearance AND matching observed growth. */
export function detectAbsorptions(before: Snapshot, after: Snapshot) {
  const next = new Map(after.entities.map(e => [e.id, e]));
  const previous = new Map(before.entities.map(e => [e.id, e]));
  const eaters = after.entities.filter(e => e.kind === 2 && previous.has(e.id))
    .map(e => ({ e, growth: e.mass - previous.get(e.id)!.mass + .01 }));
  const events: { item: number; eater: number }[] = [];
  for (const item of before.entities) {
    if (item.kind !== 1 || next.has(item.id)) continue;
    const candidate = eaters.filter(({ e, growth }) => growth >= item.mass * .75 && Math.hypot(e.x - item.x, e.y - item.y) <= 4 * Math.sqrt(e.mass) + 18)
      .sort((a, b) => Math.hypot(a.e.x - item.x, a.e.y - item.y) - Math.hypot(b.e.x - item.x, b.e.y - item.y))[0];
    if (!candidate) continue;
    candidate.growth -= item.mass;
    events.push({ item: item.id, eater: candidate.e.id });
    if (events.length >= 32) break;
  }
  return events;
}

export function cellCaption(name: string, bot: boolean, screenRadius: number) {
  const letters = Array.from(name);
  const font = Math.min(14, Math.max(10, screenRadius * .58));
  const capacity = Math.floor(Math.max(1, screenRadius * 2 - 4) / (font * .56));
  const max = Math.max(1, Math.min(14, capacity - (bot ? 3 : 0)));
  const short = letters.length > max ? max === 1 ? letters[0] : `${letters.slice(0, max - 1).join("")}…` : name;
  return { text: short + (bot ? "·IA" : ""), font };
}
