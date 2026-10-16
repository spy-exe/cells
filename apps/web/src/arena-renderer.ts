import "pixi.js/browser";
// CSP-compatible uniform/shader dispatch: this shim avoids eval.
import "pixi.js/unsafe-eval";
import {
  WebGLRenderer,
  Container,
  Graphics,
  MeshSimple,
  Sprite,
  Text,
  Texture,
} from "pixi.js";
import { MotionTimeline, virusPoints, cellDeformation } from "./motion";
import { entityDepth, membraneGeometry, updateMembrane, absorptionPose, detectAbsorptions, cellCaption } from "./organic";
import type { Snapshot } from "../../../packages/protocol/game";

const palette = [
  0x90bc4b, 0xe19a76, 0x8c91d6, 0x68b5bd, 0xd6b465, 0xc584ae, 0x709bd2,
  0xb2b557,
];
interface View {
  sprite: Sprite | MeshSimple;
  radius: number;
  born: number;
  retiring: number;
  own: boolean;
  label: Text | null;
  x: number; y: number; at: number; speed: number; angle: number;
  mass: number; impulseAt: number;
  membraneAt: number;
  reduced?: boolean;
}
export class ArenaRenderer {
  private renderer!: WebGLRenderer;
  private stage = new Container();
  private resize: ResizeObserver | null = null;
  private frame = 0;
  private renderedAt = 0;
  private measuredFPS = 0;
  private world = new Container();
  private views = new Map<number, View>();
  private circle: Texture = Texture.EMPTY;
  private virus: Texture = Texture.EMPTY;
  private absorptions = new Map<number, number>();
  private timeline = new MotionTimeline();
  private current: Snapshot | null = null;

  private camera = { x: 0, y: 0, scale: 1 };

  private quality = "high";
  private destroyed = false;
  private ready = false;
  private gridSize = 0;
  private grid: Graphics | null = null;
  private fpsAt = 0;
  onFPS: (fps: number) => void = () => {};
  async init(host: HTMLElement) {
    this.renderer = new WebGLRenderer();

    await this.renderer.init({
      width: host.clientWidth,
      height: host.clientHeight,
      background: 0xf4f7ef,
      antialias: true,
      resolution: Math.min(devicePixelRatio, 2),
      autoDensity: true,
      failIfMajorPerformanceCaveat: false,
      skipExtensionImports: true,
      powerPreference: "low-power",
    });

    if (this.destroyed) {
      this.renderer.destroy({ removeView: true });
      return;
    }
    host.appendChild(this.renderer.canvas);
    this.renderer.canvas.setAttribute(
      "aria-label",
      "Arena Cells. Use mouse, setas ou toque para mover. Espaço divide e W ejeta massa.",
    );
    this.renderer.canvas.setAttribute("role", "img");
    this.renderer.canvas.tabIndex = 0;
    const circle = new Graphics()
      .circle(64, 64, 61)
      .fill(0xffffff)
      .stroke({ width: 3, color: 0xffffff, alpha: 0.6 });
    this.circle = this.renderer.generateTexture(circle);
    circle.destroy();
    const points = virusPoints();
    const virus = new Graphics()
      .poly(points)
      .fill(0xffffff);
    this.virus = this.renderer.generateTexture(virus);
    virus.destroy();
    this.stage.addChild(this.world);
    this.ready = true;
    this.resize = new ResizeObserver(() =>
      this.renderer.resize(host.clientWidth, host.clientHeight),
    );
    this.resize.observe(host);
    const tick = (now: number) => {
      if (this.destroyed) return;
      this.frame = requestAnimationFrame(tick);
      const dt = now - this.renderedAt;
      // High quality follows the display refresh rate. A fixed 14ms gate
      // drops regular frames on 90/144Hz displays and creates uneven motion.
      if (this.quality === "low" && dt < 32) return;
      this.measuredFPS = 1000 / Math.max(1, dt);
      this.renderedAt = now;
      this.render();
      this.renderer.render({ container: this.stage });
    };
    this.frame = requestAnimationFrame(tick);
  }
  setQuality(quality: string) {
    this.quality = quality;
  }
  update(snapshot: Snapshot) {
    if (this.current && this.current.self === snapshot.self && snapshot.tick > this.current.tick) {
      for (const event of detectAbsorptions(this.current, snapshot)) this.absorptions.set(event.item, event.eater);
    }
    this.timeline.push(snapshot);
    const first = !this.current || this.current.alive !== snapshot.alive;
    this.current = snapshot;
    if (first) {
      this.camera.x = snapshot.x;
      this.camera.y = snapshot.y;
    }
  }
  target(clientX: number, clientY: number) {
    if (!this.ready)
      return { x: this.current?.x ?? 0, y: this.current?.y ?? 0 };
    const rect = this.renderer.canvas.getBoundingClientRect();
    return {
      x:
        this.camera.x +
        (clientX - rect.left - rect.width / 2) / this.camera.scale,
      y:
        this.camera.y +
        (clientY - rect.top - rect.height / 2) / this.camera.scale,
    };
  }
  private render() {
    const s = this.current;
    if (!s || !this.ready) return;
    if (this.gridSize !== s.size) {
      this.grid?.destroy();
      const g = new Graphics();
      for (let x = 0; x <= s.size; x += 80) {
        g.moveTo(x, 0).lineTo(x, s.size);
        g.moveTo(0, x).lineTo(s.size, x);
      }
      g.stroke({ width: 1, color: 0x9bac8e, alpha: 0.16 });
      g.rect(0, 0, s.size, s.size).stroke({
        width: 8,
        color: 0x9bac8e,
        alpha: 0.5,
      });
      g.zIndex = -1;
      this.world.addChild(g);
      this.grid = g;
      this.gridSize = s.size;
    }
    const { width, height } = this.renderer.screen;
    const targetScale = Math.max(width / 1.9, height / 1.4) / s.view;
    const now = performance.now(),
      center = this.timeline.center(now);
    const reduced =
      document.documentElement.dataset.motion === "reduce" ||
      matchMedia("(prefers-reduced-motion: reduce)").matches;
    this.camera.x = center.x;
    this.camera.y = center.y;
    this.camera.scale +=
      (targetScale - this.camera.scale) *
      (reduced ? 1 : 1 - Math.exp(-1000 / Math.max(1, this.measuredFPS) / 180));
    this.world.position.set(
      width / 2 - this.camera.x * this.camera.scale,
      height / 2 - this.camera.y * this.camera.scale,
    );
    this.world.scale.set(this.camera.scale);
    const seen = new Set<number>();
    for (const e of s.entities) {
      seen.add(e.id);
      let v = this.views.get(e.id);
      if (!v) {
        const sprite =
          e.kind === 2
            ? new MeshSimple({ texture: Texture.WHITE, ...membraneGeometry() })
            : new Sprite(e.kind === 4 ? this.virus : this.circle);
        if (sprite instanceof Sprite) sprite.anchor.set(0.5);
        else sprite.autoUpdate = false;
        this.world.addChild(sprite);
        const label =
          e.kind === 2
            ? new Text({
                text: "",
                style: {
                  fontFamily: "Inter Variable, sans-serif",
                  fontWeight: "600",
                  fontSize: 14,
                  fill: 0x17251c,
                  align: "center",
                },
              })
            : null;
        if (label) {
          label.anchor.set(0.5);
          this.world.addChild(label);
        }
        v = {
          sprite,
          label,
          radius: 4 * Math.sqrt(e.mass),
          born: now,
          retiring: 0,
          own: e.owner === s.self && e.kind === 2,
          x: e.x, y: e.y, at: now, speed: 0, angle: 0,
          mass: e.mass, impulseAt: now,
          membraneAt: -Infinity,
        };
        this.views.set(e.id, v);
      }
      const pose = e.kind === 1 ? e : (this.timeline.entity(e.id, now) ?? e);
      const x = pose.x,
        y = pose.y,
        targetRadius = 4 * Math.sqrt(pose.mass);
      v.radius +=
        (targetRadius - v.radius) *
        (reduced
          ? 1
          : 1 - Math.exp(-1000 / Math.max(1, this.measuredFPS) / 95));
      v.retiring = 0;
      this.absorptions.delete(e.id);
      v.sprite.zIndex = entityDepth(e.kind, e.mass);
      const bornScale =
        reduced || e.kind === 1 ? 1 : Math.min(1, 0.82 + (now - v.born) / 650);
      const pulse =
        !reduced && e.kind === 1 && e.mass >= 3
          ? 1 + 0.06 * Math.sin(now / 320 + e.id)
          : 1;
      const r = v.radius * bornScale * pulse;
      v.sprite.position.set(x, y);
      if (v.sprite instanceof MeshSimple) {
        const dt = Math.max(1, now - v.at), dx = x - v.x, dy = y - v.y;
        const speed = Math.hypot(dx, dy) * 1000 / dt;
        v.speed += (speed - v.speed) * (1 - Math.exp(-dt / 90));
        if (speed > 5) {
          const next = Math.atan2(dy, dx), delta = Math.atan2(Math.sin(next - v.angle), Math.cos(next - v.angle));
          v.angle += delta * (1 - Math.exp(-dt / 75));
        }
        if (e.mass > v.mass * 1.02) v.impulseAt = now;
        v.mass = e.mass; v.x = x; v.y = y; v.at = now;
        const shape = cellDeformation(v.speed, now - v.impulseAt, reduced);
        v.sprite.rotation = v.angle;
        v.sprite.scale.set(r * shape.x, r * shape.y);
        if (v.reduced !== reduced || !reduced && now - v.membraneAt >= (this.quality === "low" ? 66 : 16)) {
          updateMembrane(v.sprite.geometry.positions, e.id, now, now - v.impulseAt, 0, reduced);
          v.sprite.geometry.getBuffer("aPosition").update();
          v.membraneAt = now;
          v.reduced = reduced;
        }
      }
      else {
        v.sprite.width = r * 2;
        v.sprite.height = r * 2;
      }
      v.sprite.tint =
        e.kind === 4
          ? 0x8bae65
          : s.teams && e.kind === 2
            ? [0xa6bd65, 0xe89b87, 0x88b6d7][e.team % 3]
            : (e.kind === 1
                ? [0x729e19, 0xe27747, 0x696ed0, 0x27a3b1, 0xc79531, 0xbd5c95]
                : palette)[
                (e.kind === 2 || e.kind === 3 ? e.owner : e.id) %
                  (e.kind === 1 ? 6 : palette.length)
              ];
      v.sprite.alpha = 1;
      if (v.label) {
        v.label.alpha = 1;
        const {text,font} = cellCaption(e.name,e.bot,r*this.camera.scale);
        if (v.label.text !== text) v.label.text = text;
        v.label.position.set(x, y);
        v.label.zIndex = entityDepth(e.kind, e.mass, true);
        const width = v.label.getLocalBounds().width;
        const scale = Math.min(font / 14, Math.max(0, r * this.camera.scale * 2 - 4) / Math.max(1, width));
        v.label.scale.set(scale / this.camera.scale);
        v.label.visible = r * this.camera.scale > 4 && (this.quality !== "low" || v.own);
      }
    }
    for (const [id, v] of this.views) {
      if (!seen.has(id)) {
        if (!v.retiring) v.retiring = now;
        const eater = this.views.get(this.absorptions.get(id) ?? -1);
        const progress = reduced ? 1 : Math.min(1, (now - v.retiring) / (eater ? 170 : 100));
        if (eater) {
          const pose = absorptionPose(v.x, v.y, eater.sprite.x, eater.sprite.y, progress);
          v.sprite.position.set(pose.x, pose.y);
          v.sprite.scale.set(v.radius * 2 / this.circle.width * pose.scale);
          v.sprite.zIndex = eater.sprite.zIndex + .001;
        }
        v.sprite.alpha = 1 - progress;
        if (v.label) v.label.alpha = 1 - progress;
        if (progress >= 1) {
          if (v.sprite instanceof MeshSimple) v.sprite.geometry.destroy();
          v.sprite.destroy();
          v.label?.destroy();
          this.views.delete(id);
          this.absorptions.delete(id);
        }
      }
    }
    this.world.sortableChildren = true;
    if (performance.now() - this.fpsAt > 500) {
      this.onFPS(Math.round(this.measuredFPS));
      this.fpsAt = performance.now();
    }
  }
  destroy() {
    this.destroyed = true;
    cancelAnimationFrame(this.frame);
    this.resize?.disconnect();
    if (this.ready) {
      for (const view of this.views.values()) if (view.sprite instanceof MeshSimple) view.sprite.geometry.destroy();
      this.stage.destroy({ children: true });
      this.renderer.destroy({ removeView: true });
      this.circle.destroy(true);
      this.virus.destroy(true);
    }
    this.views.clear();
    this.absorptions.clear();
  }
}
