export type Mode = "classic" | "arena" | "teams" | "rogue";
export interface Status {
  ok: boolean;
  ready: boolean;
  version: string;
  revision: string;
  modes: Mode[];
  players: number;
  bots: number;
}
export function parseStatus(value: unknown): Status {
  if (!value || typeof value !== "object") throw new Error("Status inválido");
  const v = value as Record<string, unknown>;
  if (
    typeof v.ok !== "boolean" ||
    typeof v.ready !== "boolean" ||
    typeof v.version !== "string" ||
    typeof v.revision !== "string" ||
    !Array.isArray(v.modes) ||
    v.modes.some((m) => !["classic", "arena", "teams", "rogue"].includes(String(m))) ||
    typeof v.players !== "number" ||
    typeof v.bots !== "number"
  )
    throw new Error("Status inválido");
  return v as unknown as Status;
}
