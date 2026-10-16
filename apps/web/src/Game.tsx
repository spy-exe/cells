import { useEffect, useRef, useState } from "react";
import {
  ArrowLeft,
  Split,
  Send,
  Eye,
  Maximize,
  Wifi,
  RotateCcw,
} from "lucide-react";
import {
  controlFrame,
  decodeSnapshot,
  inputFrame,
  type Snapshot,
} from "../../../packages/protocol/game";
import { ArenaRenderer } from "./arena-renderer";

interface Props {
  nickname: string;
  mode: string;
  onLeave: () => void;
}
export function Game({ nickname, mode, onLeave }: Props) {
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
    const input = setInterval(() => {
      const ws = socket.current,
        s = latest.current;
      if (!s || ws?.readyState !== WebSocket.OPEN) return;
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
      ws.send(inputFrame(++seq.current, target.x, target.y, actions.current));
      actions.current = 0;
    }, 50);
    const ping = setInterval(() => {
      if (socket.current?.readyState === WebSocket.OPEN)
        socket.current.send(controlFrame(5));
    }, 2000);
    return () => {
      stopped = true;
      clearTimeout(timer);
      clearInterval(input);
      clearInterval(ping);
      socket.current?.close(1000);
    };
  }, [retry]);
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLSelectElement ||
        e.target instanceof HTMLButtonElement
      )
        return;
      const key = e.key.toLowerCase();
      if (key.startsWith("arrow") || key === " " || key === "w")
        e.preventDefault();
      keys.current.add(key);
      if (!e.repeat && key === " ") actions.current |= 1;
      if (!e.repeat && key === "w") actions.current |= 2;
    };
    const up = (e: KeyboardEvent) => keys.current.delete(e.key.toLowerCase());
    const blur = () => {
      keys.current.clear();
      actions.current = 0;
    };
    window.addEventListener("keydown", down);
    window.addEventListener("keyup", up);
    window.addEventListener("blur", blur);
    return () => {
      window.removeEventListener("keydown", down);
      window.removeEventListener("keyup", up);
      window.removeEventListener("blur", blur);
    };
  }, []);
  const steer = (x: number, y: number) => {
    pointer.current = { x, y };
    stick.current = null;
  };
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
              : "Classic FFA"}
        </span>
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
        <small data-testid="position">
          {snapshot
            ? `${Math.round(snapshot.x)} : ${Math.round(snapshot.y)}`
            : "Entrando…"}
        </small>
      </section>
      <aside className="leaderboard" aria-label="Maiores células">
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
      </aside>
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
        ejetar
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
          disabled={!snapshot?.alive || !!snapshot.splitIn}
          onClick={() => (actions.current |= 1)}
        >
          <Split size={23} />
          <span>Dividir</span>
        </button>
        <button
          aria-label="Ejetar massa"
          disabled={
            !snapshot?.alive || !!snapshot.ejectIn || snapshot.mass < 32
          }
          onClick={() => (actions.current |= 2)}
        >
          <Send size={21} />
          <span>Ejetar</span>
        </button>
      </div>
      {(!snapshot || !snapshot.alive) && !error && !renderError && (
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
