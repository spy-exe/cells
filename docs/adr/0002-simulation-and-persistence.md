# ADR 0002 — Fixed-step world and versioned checkpoints

Accepted, 2026-10-16.

The engine is a pure Go package with an explicit serializable xorshift64 state,
a fixed 1/30s step and a uniform collision grid. Stable slice and bucket iteration
make same-seed, same-input replays reproducible within the same engine version.
Floating-point replay across architectures is not promised.

Food and spawns are mass sources. Decay, ejection loss and expired/disconnected
entities are sinks. Splits, merges, consumption and virus feeding transfer mass.
Tests verify the ledger, contextual consumption, limits and checkpoint replay.

The server owns rooms and serializes access. PostgreSQL stores hashed guest
sessions and world checkpoints; a checkpoint includes player identities, cells,
RNG and tick. Saves occur every 30 seconds and on orderly termination. Abrupt power
loss may lose up to one interval of transient arena state. No client save is trusted.
Schema migrations are additive and idempotent. Checkpoint version 1 must remain
readable until an explicit migration exists. Rollback never restores an old database
over newer player data.

Native services avoid nesting/Docker privilege requirements. The deploy candidate
uses CELLS_CANDIDATE=1: it verifies its binary, database and web without running or
saving a second copy of production worlds. Only the active instance owns game state.
