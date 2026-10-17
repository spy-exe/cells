import { useEffect, useRef, useState } from "react";
import {
  ArrowLeft,
  Split,
  Send,
  Eye,
  Maximize,
  Wifi,
  RotateCcw,
  Volume2, VolumeX,
} from "lucide-react";
import {
  controlFrame,
  decodeSnapshot,
  inputFrame,
  type Snapshot,
} from "../../../packages/protocol/game";
import { ArenaRenderer } from "./arena-renderer";
import { GrowthFeedback } from "./motion";
import { GameSound } from "./sound";
import { Teams } from "./Teams";
import { Expedition, useExpedition } from "./Expedition";

interface Props {
  nickname: string;
  mode: string;
  onLeave: () => void;
}
export function Game({ nickname, mode, onLeave }: Props) {
  const expedition = useExpedition(mode === "rogue");
  const expeditionRef = useRef(expedition.run); expeditionRef.current = expedition.run;
  const canAct = () => mode !== "rogue" || expeditionRef.current?.phase === "running";
  const host = useRef<HTMLDivElement>(null),
    renderer = useRef<ArenaRenderer | null>(null),
    socket = useRef<WebSocket | null>(null),
    latest = useRef<Snapshot | null>(null);
  const aim = useRef<{ x: number; y: number } | null>(null),
    actions = useRef(0),
    seq = useRef(0),
    keys = useRef(new Set<string>());
  const pointer = useRef<{ x: number; y: number } | null>(null),
    stick = useRef<{ x: number; y: number } | null>(null);
  const send = useRef<(extra?: number) => void>(() => {}),
    ejectHeld = useRef(false),
    lastKeyAction = useRef(0);
  const [feedback, setFeedback] = useState({ kind: 0, seq: 0 });
  const growth = useRef(new GrowthFeedback()), sound = useRef(new GameSound());
  const [gain, setGain] = useState<{ amount: number; sequence: number } | null>(null);
  const [audio, setAudio] = useState(false);
  const [run, setRun] = useState<{ has_run: boolean; alive: boolean; peak: number; record: number; kills: number; seconds: number } | null>(null);
  const requestAction = (kind: number) => {
    const s = latest.current;
    if (!s?.alive || !canAct()) return;
    const cells = s.entities.filter((e) => e.kind === 2 && e.owner === s.self);
    if (kind === 1 && (cells.length >= 16 || !cells.some((e) => e.mass >= 36)))
      return;
    if (kind === 2 && !cells.some((e) => e.mass >= 32)) return;
    setFeedback((v) => ({ kind, seq: v.seq + 1 }));
    send.current(kind);
    host.current?.querySelector("canvas")?.focus({ preventScroll: true });
  };
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null),
    [connection, setConnection] = useState("Conectando…"),
    [error, setError] = useState(""),
    [renderError, setRenderError] = useState(""),
    [fps, setFPS] = useState(0),
    [latency, setLatency] = useState<number | null>(null),
    [quality, setQuality] = useState("high"),
    [retry, setRetry] = useState(0);
  const [joystick, setJoystick] = useState<{ x: number; y: number } | null>(
    null,
  );
  useEffect(() => {
    const r = new ArenaRenderer();
    renderer.current = r;
    let disposed = false;
    void r
      .init(host.current!)
      .then(() => {
        if (!disposed) r.setQuality(quality);
      })
      .catch(() =>
        setRenderError(
          "Não foi possível iniciar WebGL. Tente atualizar o navegador.",
        ),
      );
    r.onFPS = setFPS;
    return () => {
      disposed = true;
      r.destroy();
      renderer.current = null;
    };
  }, []);
  useEffect(() => renderer.current?.setQuality(quality), [quality]);
  useEffect(() => {
    let stopped = false,
      timer: ReturnType<typeof setTimeout> | undefined,
      attempt = 0;
    const connect = () => {
      if (stopped) return;
      const ws = new WebSocket(
        `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/ws`,
      );
      ws.binaryType = "arraybuffer";
      socket.current = ws;
      seq.current = 0;
      ws.onopen = () => {
        attempt = 0;
        setConnection("Ao vivo");
        setError("");
      };
      ws.onmessage = (event) => {
        try {
          const b = event.data as ArrayBuffer,
            v = new DataView(b);
          if (v.getUint8(2) === 6) {
            setLatency(Math.round(performance.now() - v.getFloat64(3, true)));
            return;
          }
          const s = decodeSnapshot(b);
          const previous = latest.current;
          if (previous?.alive && !s.alive) sound.current.play("death");
          if (previous?.alive && s.alive) {
            const beforeCells = previous.entities.filter(e => e.kind === 2 && e.owner === previous.self).length;
            const afterCells = s.entities.filter(e => e.kind === 2 && e.owner === s.self).length;
            if (afterCells > beforeCells && s.splitIn > 0) sound.current.play("split");
            if (s.mass < previous.mass - 10 && s.ejectIn > 0) sound.current.play("eject");
          }
          const acquired = growth.current.update(s.mass, s.alive, performance.now());
          if (acquired) { setGain(acquired); sound.current.play("gain"); }
          latest.current = s;
          renderer.current?.update(s);
          setSnapshot(s);
        } catch {
          setError(
            "O servidor enviou uma atualização incompatível. Recarregue a página.",
          );
          ws.close(1000);
        }
      };
      ws.onclose = (event) => {
        if (stopped) return;
        setConnection("Conexão interrompida");
        if (event.code === 1008) {
          setError("Esta sessão foi aberta em outra janela ou foi encerrada.");
          return;
        }
        attempt++;
        if (attempt > 8) {
          setError(
            "Não foi possível reconectar. Volte ao menu para entrar novamente.",
          );
          return;
        }
        timer = setTimeout(connect, Math.min(5000, 500 * attempt));
      };
      ws.onerror = () => setConnection("Reconectando…");
    };
    connect();
    const transmit = (extra = 0) => {
      const ws = socket.current,
        s = latest.current;
      if (!s || ws?.readyState !== WebSocket.OPEN || !canAct()) return;
      let target = aim.current ?? { x: s.x, y: s.y };
      if (pointer.current && renderer.current)
        target = renderer.current.target(pointer.current.x, pointer.current.y);
      if (stick.current)
        target = {
          x: s.x + stick.current.x * 14,
          y: s.y + stick.current.y * 14,
        };
      const k = keys.current;
      const dx = Number(k.has("arrowright")) - Number(k.has("arrowleft")),
        dy = Number(k.has("arrowdown")) - Number(k.has("arrowup"));
      if (dx || dy) target = { x: s.x + dx * 800, y: s.y + dy * 800 };
      let commands = actions.current | extra;
      if (
        (keys.current.has("w") || ejectHeld.current) &&
        s.entities.some(
          (e) => e.owner === s.self && e.kind === 2 && e.mass >= 32,
        )
      )
        commands |= 2;
      ws.send(inputFrame(++seq.current, target.x, target.y, commands));
      actions.current = 0;
    };
    send.current = transmit;
    const input = setInterval(() => transmit(), 50);
    const ping = setInterval(() => {
      if (socket.current?.readyState === WebSocket.OPEN)
        socket.current.send(controlFrame(5));
    }, 2000);
    return () => {
      stopped = true;
      clearTimeout(timer);
      clearInterval(input);
      send.current = () => {};
      clearInterval(ping);
      socket.current?.close(1000);
    };
  }, [retry]);
  useEffect(() => () => sound.current.destroy(), []);
  useEffect(() => {
    if (mode === "rogue" || !snapshot || snapshot.alive) { setRun(null); return; }
    const controller = new AbortController();
    void fetch("/api/run", { signal: controller.signal })
      .then(async response => {
        if (!response.ok) return;
        const value = await response.json();
        if (value.has_run && !value.alive && [value.peak, value.record, value.kills, value.seconds].every(Number.isFinite)) setRun(value);
      }).catch(() => {});
    return () => controller.abort();
  }, [snapshot?.alive]);
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLSelectElement ||
        e.target instanceof HTMLTextAreaElement
      )
        return;
      const key = e.key.toLowerCase();
      if (key === "r" && !e.repeat && latest.current && (mode === "rogue" ? ["fallen","failed","victory"].includes(expeditionRef.current?.phase ?? "") : !latest.current.alive && latest.current.respawnIn === 0)) {
        aim.current = null; socket.current?.send(controlFrame(3)); return;
      }
      if (e.target instanceof HTMLButtonElement && key === " ") return;
      if (key.startsWith("arrow") || key === " " || key === "w")
        e.preventDefault();
      keys.current.add(key);
      if (
        (key === " " || key === "w") &&
        (!e.repeat || performance.now() - lastKeyAction.current >= 100)
      ) {
        requestAction(key === " " ? 1 : 2);
        lastKeyAction.current = performance.now();
      }
    };
    const up = (e: KeyboardEvent) => keys.current.delete(e.key.toLowerCase());
    const blur = () => {
      keys.current.clear();
      ejectHeld.current = false;
      actions.current = 0;
    };
    const release = () => {
      ejectHeld.current = false;
    };
    window.addEventListener("pointerup", release, true);
    window.addEventListener("pointercancel", release, true);
    window.addEventListener("keydown", down);
    window.addEventListener("keyup", up);
    window.addEventListener("blur", blur);
    return () => {
      window.removeEventListener("pointerup", release, true);
      window.removeEventListener("pointercancel", release, true);
      window.removeEventListener("keydown", down);
      window.removeEventListener("keyup", up);
      window.removeEventListener("blur", blur);
    };
  }, []);
  const steer = (x: number, y: number) => {
    pointer.current = { x, y };
    stick.current = null;
  };
  const cells =
    snapshot?.entities.filter(
      (e) => e.kind === 2 && e.owner === snapshot.self,
    ) ?? [];
  const splitRequirement = !snapshot?.alive
    ? ""
    : cells.length >= 16
      ? "Limite 16"
      : !cells.some((e) => e.mass >= 36)
        ? "36 por célula"
        : "";
  const ejectRequirement =
    snapshot?.alive && !cells.some((e) => e.mass >= 32) ? "32 por célula" : "";
  return (
    <main className="game-shell" aria-label="Partida Cells">
      <div
        ref={host}
        className="arena-canvas"
        onPointerMove={(e) => {
          if (e.pointerType === "mouse") steer(e.clientX, e.clientY);
        }}
        onPointerDown={(e) => steer(e.clientX, e.clientY)}
      />
      <header className="game-header">
        <button
          className="icon-button"
          onClick={onLeave}
          aria-label="Sair da partida"
        >
          <ArrowLeft size={20} />
        </button>
        <span className="game-brand">
          cells<span>.</span>
        </span>
        <span className="mode-pill">
          {mode === "arena"
            ? "AI Arena"
            : mode === "teams"
              ? "Teams"
              : mode === "rogue" ? "Expedição" : "Classic FFA"}
        </span>
        <button className="icon-button sound-control" aria-label={audio ? "Desativar sons" : "Ativar sons"}
          aria-pressed={audio} onClick={() => {
            if (audio) { sound.current.enabled = false; setAudio(false); }
            else void sound.current.enable().then(() => setAudio(true)).catch(() => setAudio(false));
          }}>
          {audio ? <Volume2 size={18} /> : <VolumeX size={18} />}
        </button>
        <label className="quality-control">
          <span>Gráficos</span>
          <select
            aria-label="Qualidade gráfica"
            value={quality}
            onChange={(e) => setQuality(e.target.value)}
          >
            <option value="high">Alta</option>
            <option value="low">Econômico</option>
          </select>
        </label>
      </header>
      <section className="mass-hud" aria-label="Seu desempenho">
        <span>Massa</span>
        <strong data-testid="mass">
          {Math.floor(snapshot?.mass ?? 0).toLocaleString("pt-BR")}
        </strong>
        <p>{nickname}</p>
        {gain && <span key={gain.sequence} className="mass-gain" aria-hidden="true">+{gain.amount.toLocaleString("pt-BR")} massa</span>}
        <small data-testid="position">
          {snapshot
            ? `${Math.round(snapshot.x)} : ${Math.round(snapshot.y)}`
            : "Entrando…"}
        </small>
      </section>
      {mode !== "rogue" && mode !== "teams" && <aside className="leaderboard" aria-label="Maiores células">
        <h2>
          Na liderança<span>TOP 10</span>
        </h2>
        <ol>
          {snapshot?.leaderboard.map((p) => (
            <li key={p.id} className={p.id === snapshot.self ? "is-you" : ""}>
              <span>
                {p.name}
                {p.bot && <small>IA</small>}
              </span>
              <b>{Math.floor(p.mass)}</b>
            </li>
          ))}
        </ol>
      </aside>}
      {mode === "teams" && <Teams snapshot={snapshot} />}
      <div className="connection-hud">
        <Wifi size={12} />
        <span>{connection}</span>
        <span>{latency === null ? "…" : latency} ms</span>
        <span>{fps} FPS</span>
      </div>
      {snapshot && (
        <div
          className="minimap"
          aria-label={`Sua posição no mapa: ${Math.round(snapshot.x)}, ${Math.round(snapshot.y)}`}
        >
          <i
            style={{
              left: `${(snapshot.x / snapshot.size) * 100}%`,
              top: `${(snapshot.y / snapshot.size) * 100}%`,
            }}
          />
          <span>N</span>
        </div>
      )}
      <div className="desktop-hint">
        <kbd>Mouse / ↑↓←→</kbd> mover <kbd>Espaço</kbd> dividir <kbd>W</kbd>{" "}
        segurar para ejetar
      </div>
      <div
        className="joystick"
        aria-label="Direcional de movimento por toque"
        role="application"
        onPointerDown={(e) => {
          e.currentTarget.setPointerCapture(e.pointerId);
          pointer.current = null;
          setJoystick({ x: e.clientX, y: e.clientY });
        }}
        onPointerMove={(e) => {
          if (!joystick) return;
          stick.current = {
            x: e.clientX - joystick.x,
            y: e.clientY - joystick.y,
          };
        }}
        onPointerUp={() => {
          setJoystick(null);
          stick.current = null;
          aim.current = null;
        }}
        onPointerCancel={() => {
          setJoystick(null);
          stick.current = null;
          aim.current = null;
        }}
      >
        <span>mover</span>
        <i />
      </div>
      <div className="action-controls">
        <button
          aria-label="Dividir célula"
          title={
            splitRequirement === "Limite 16"
              ? "Limite de 16 células atingido"
              : splitRequirement
                ? "Uma célula precisa de massa 36 para dividir"
                : "Espaço para dividir"
          }
          disabled={
            !canAct() || !snapshot?.alive || !!snapshot.splitIn || !!splitRequirement
          }
          onPointerDown={(e) => {
            e.preventDefault();
            requestAction(1);
          }}
          onClick={(e) => {
            if (e.detail === 0) requestAction(1);
          }}
        >
          <Split
            key={feedback.seq}
            className={feedback.kind === 1 ? "action-feedback" : undefined}
            size={23}
          />
          <span>{splitRequirement || "Dividir"}</span>
        </button>
        <button
          aria-label="Ejetar massa"
          title={
            ejectRequirement
              ? "Uma célula precisa de massa 32 para ejetar"
              : "Segure para ejetar massa"
          }
          disabled={
            !canAct() || !snapshot?.alive || !!snapshot.ejectIn || !!ejectRequirement
          }
          onPointerDown={(e) => {
            e.preventDefault();
            e.currentTarget.setPointerCapture(e.pointerId);
            ejectHeld.current = true;
            requestAction(2);
          }}
          onPointerUp={() => {
            ejectHeld.current = false;
          }}
          onPointerCancel={() => {
            ejectHeld.current = false;
          }}
          onClick={(e) => {
            if (e.detail === 0) requestAction(2);
          }}
        >
          <Send
            key={feedback.seq}
            className={feedback.kind === 2 ? "action-feedback" : undefined}
            size={21}
          />
          <span>{ejectRequirement || "Ejetar"}</span>
        </button>
      </div>
      {mode !== "rogue" && (!snapshot || !snapshot.alive) && !error && !renderError && (
        <section
          className="arena-overlay"
          aria-label={snapshot ? "Fim da vida" : "Conectando"}
        >
          <div className="death-orbit" aria-hidden="true" />
          <p className="eyebrow">
            {snapshot ? "Cada fim é um novo começo" : "Preparando seu espaço"}
          </p>
          <h1>{snapshot ? "Pequena outra vez." : "Entrando na arena…"}</h1>
          <p>
            {snapshot
              ? "Você está acompanhando a liderança. Escolha sua próxima tentativa."
              : "Conectando à simulação em tempo real."}
          </p>
          {run && <dl className="run-summary" aria-label="Resumo da rodada">
            <div><dt>Pico de massa</dt><dd>{Math.floor(run.peak).toLocaleString("pt-BR")}</dd></div>
            <div><dt>Sobreviveu</dt><dd>{Math.floor(run.seconds / 60)}:{String(Math.floor(run.seconds % 60)).padStart(2, "0")}</dd></div>
            <div><dt>Eliminações</dt><dd>{run.kills}</dd></div>
            <div><dt>Recorde da sessão</dt><dd>{Math.floor(run.record).toLocaleString("pt-BR")}</dd></div>
          </dl>}
          {snapshot && (
            <>
              <button
                className="play-button"
                disabled={snapshot.respawnIn > 0}
                onClick={() => {
                  aim.current = null;
                  socket.current?.send(controlFrame(3));
                }}
              >
                <RotateCcw size={18} />
                {snapshot.respawnIn > 0
                  ? `Respawn em ${Math.ceil(snapshot.respawnIn / 30)} s`
                  : "Voltar à arena"}
              </button>
              <small className="respawn-hint">R para recomeçar</small>
              <button className="text-button" onClick={onLeave}>
                Voltar ao menu
              </button>
              <span className="spectate-note">
                <Eye size={14} />
                Modo espectador
              </span>
            </>
          )}
        </section>
      )}
      {mode === "rogue" && !error && !renderError && <Expedition {...expedition} onRestart={() => { aim.current = null; socket.current?.send(controlFrame(3)); }} onLeave={onLeave} />}
      {(error || renderError) && (
        <section className="arena-overlay" role="alert">
          <h1>Uma pausa na conexão.</h1>
          <p>{renderError || error}</p>
          <button
            className="play-button"
            onClick={() => {
              setError("");
              setRetry(retry + 1);
            }}
          >
            Tentar novamente
          </button>
          <button className="text-button" onClick={onLeave}>
            Voltar ao menu
          </button>
        </section>
      )}
      <button
        className="fullscreen icon-button"
        aria-label="Alternar tela cheia"
        onClick={() => {
          if (document.fullscreenElement) void document.exitFullscreen();
          else
            void document.documentElement
              .requestFullscreen?.()
              .catch(() =>
                setError("Tela cheia indisponível neste navegador."),
              );
        }}
      >
        <Maximize size={18} />
      </button>
    </main>
  );
}
