import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  ArrowUpRight,
  ArrowRight,
  Circle,
  Radio,
  Code2,
  Settings2,
} from "lucide-react";
import { parseStatus, type Status } from "../../../packages/protocol/status";
import "./style.css";
import { Game } from "./Game";

function App() {
  const [status, setStatus] = useState<Status | null>(null);
  const [nickname, setNickname] = useState("");
  const [playing, setPlaying] = useState(false);
  const [mode, setMode] = useState("classic");
  const [profile, setProfile] = useState("balanced");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [room, setRoom] = useState(
    new URLSearchParams(location.search).get("room") ?? "",
  );
  const join = async (spectate = false) => {
    setBusy(true);
    setError("");
    try {
      const res = await fetch("/api/guest", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nickname: spectate ? nickname.trim() || "Visitante" : nickname.trim(),
          mode,
          room,
          profile,
          spectate,
        }),
      });
      if (!res.ok) {
        const body = await res.json();
        throw Error(body.error || "Não foi possível entrar.");
      }
      setPlaying(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Não foi possível conectar.");
    } finally {
      setBusy(false);
    }
  };
  const leave = () => {
    setPlaying(false);
    void fetch("/api/logout", { method: "POST" });
  };
  useEffect(() => {
    let active = true;
    void fetch("/api/me")
      .then(async (res) => {
        if (res.ok) {
          const me = await res.json();
          if (active) {
            setNickname(me.nickname);
            setMode(me.mode);
            setPlaying(true);
          }
        }
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);
  const [effects, setEffects] = useState(() => {
    try {
      return localStorage.getItem("cells.motion") !== "reduce";
    } catch {
      return true;
    }
  });
  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        const r = await fetch("/api/status");
        if (!r.ok) throw Error();
        const value = parseStatus(await r.json());
        if (active) setStatus(value);
      } catch {
        if (active) setStatus(null);
      }
    };
    void load();
    const timer = setInterval(() => void load(), 15000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, []);
  useEffect(() => {
    document.documentElement.dataset.motion = effects ? "full" : "reduce";
    try {
      localStorage.setItem("cells.motion", effects ? "full" : "reduce");
    } catch {
      /* Preference remains usable for this visit. */
    }
  }, [effects]);
  if (playing) return <Game nickname={nickname} mode={mode} onLeave={leave} />;
  return (
    <main className="entry">
      <div className="field-art" aria-hidden="true">
        <div className="field-grid" />
        {Array.from({ length: 18 }, (_, i) => (
          <i
            key={i}
            className={`orb orb-${i % 5}`}
            style={
              {
                "--x": `${(i * 37 + 11) % 100}%`,
                "--y": `${(i * 53 + 7) % 100}%`,
                "--size": `${i % 4 === 0 ? 140 : i % 3 === 0 ? 72 : 12}px`,
                "--delay": `${(i % 6) * -0.8}s`,
              } as React.CSSProperties
            }
          />
        ))}
      </div>
      <header className="entry-header">
        <a className="wordmark" href="/" aria-label="Cells, início">
          <span className="brand-orbits" aria-hidden="true">
            <i />
            <i />
          </span>
          cells<span className="wordmark-dot">.</span>
        </a>
        <a
          className="source-link"
          href="https://github.com/spy-exe/cells"
          target="_blank"
          rel="noreferrer"
        >
          <Code2 size={17} aria-hidden="true" />
          <span>Código aberto</span>
          <ArrowUpRight size={15} aria-hidden="true" />
        </a>
      </header>
      <section className="entry-content" aria-labelledby="entry-title">
        <div className="intro-copy">
          <p className="eyebrow">
            <span />
            Um mundo, muitas células
          </p>
          <h1 id="entry-title">
            Comece pequeno.
            <br />
            <em>Ocupe espaço.</em>
          </h1>
          <p className="intro-description">
            Coma. Cresça. Divida. Sobreviva.
            <br />O próximo gigante da arena pode ser você.
          </p>
          <div className="entry-keys" aria-label="Controles do jogo">
            <span>
              <span className="keycap">↗</span>Mover
            </span>
            <span>
              <span className="keycap">Espaço</span>Dividir
            </span>
            <span>
              <span className="keycap">W</span>Ejetar
            </span>
          </div>
        </div>
        <div className="join-surface">
          <p className="join-label">
            <Circle size={15} aria-hidden="true" />
            Sua próxima partida
          </p>
          <form
            noValidate
            onSubmit={(e) => {
              e.preventDefault();
              void join();
            }}
          >
            <label htmlFor="nickname">Como você quer aparecer?</label>
            <input
              id="nickname"
              autoComplete="nickname"
              maxLength={20}
              placeholder="Seu apelido na arena"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
            />
            {status && status.modes.length > 1 && (
              <>
                <label htmlFor="mode">Modo</label>
                <select
                  id="mode"
                  value={mode}
                  onChange={(e) => setMode(e.target.value)}
                >
                  {status.modes.map((m) => (
                    <option key={m} value={m}>
                      {m === "classic"
                        ? "Classic FFA"
                        : m === "arena"
                          ? "AI Arena"
                          : "Teams"}
                    </option>
                  ))}
                </select>
              </>
            )}
            {mode === "arena" && (
              <>
                <label htmlFor="bot-profile">Como as IAs vão jogar?</label>
                <select
                  id="bot-profile"
                  value={profile}
                  onChange={(e) => setProfile(e.target.value)}
                >
                  <option value="balanced">Equilibradas</option>
                  <option value="aggressive">Agressivas</option>
                  <option value="cautious">Cautelosas</option>
                  <option value="opportunist">Oportunistas</option>
                  <option value="beginner">Iniciantes</option>
                </select>
                <p className="foundation-note">
                  IAs táticas, com regras programadas. Sem modelos treinados
                  nesta arena.
                </p>
              </>
            )}
            <details className="room-options">
              <summary>Jogar com amigos</summary>
              <label htmlFor="room">Mesmo código, mesma arena</label>
              <input
                id="room"
                placeholder="Ex.: turma-cells"
                maxLength={32}
                value={room}
                onChange={(e) => setRoom(e.target.value.toLowerCase())}
              />
              <p>Escolha um código e compartilhe com seus amigos.</p>
            </details>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <button
              className="play-button"
              type="submit"
              disabled={!status?.modes.length || busy}
            >
              {busy ? "Conectando…" : "Entrar na arena"}
              <ArrowRight size={20} aria-hidden="true" />
            </button>
          </form>
          <p className="server-state" role="status">
            <Radio size={15} aria-hidden="true" />
            <span>
              {status?.ready
                ? `Servidor online · ${status.players} jogando`
                : "Conectando ao servidor…"}
            </span>
          </p>
          <p className="foundation-note">
            Sem cadastro. Sem vantagens pagas. Só você, suas células e uma arena
            de possibilidades.
          </p>
          {!!status?.modes.length && (
            <button
              className="text-button"
              onClick={() => void join(true)}
              disabled={busy}
            >
              Só quero assistir
            </button>
          )}
          <button
            type="button"
            role="switch"
            aria-checked={!effects}
            onClick={() => setEffects(!effects)}
            className="effects-toggle"
          >
            <Settings2 size={16} aria-hidden="true" />
            Reduzir movimento<span>{effects ? "Desligado" : "Ligado"}</span>
          </button>
        </div>
      </section>
      <footer className="entry-footer">
        <p>Pequenas escolhas. Grandes células.</p>
        <span className="version">
          {status ? `v${status.version}` : "Cells"}
          <span aria-hidden="true"> / </span>
          <a
            href="https://github.com/spy-exe/cells/blob/main/SPEC.md"
            target="_blank"
            rel="noreferrer"
          >
            Acompanhe a construção
            <ArrowUpRight size={13} aria-hidden="true" />
          </a>
        </span>
      </footer>
    </main>
  );
}

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
