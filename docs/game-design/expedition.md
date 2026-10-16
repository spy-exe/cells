# Expedição — experimental solo rules module

Expedição keeps the Classic movement, consumption, split, ejection and viruses in
an independent finite run. It is not Classic progression and cannot grant an
advantage in Classic or AI Arena. A public session gets its own private room.

## Loop

Three sectors last25 seconds of authoritative simulation time each. At the
sector deadline the player must be alive and hold60,100,160 mass respectively.
Reaching the number early does not finish the sector: protect the mass until the
countdown ends. Failure or consumption ends the run. Sector3 success wins.

The world pauses after successful sectors1 and2. The player chooses one permanent
mutation, then the next sector starts. Mutations stack additively, only in this
run: Cílios gives10% movement speed, Metabolismo gives20% extra food mass, and
Coesão reduces the normal post-split merge wait by20%. Digestion does not multiply
player consumption or ejected mass. Bonus mass is an explicit ledger source.

Four tactical agents arrive per sector; first sector uses beginner profiles,
second introduces balanced agents, third opportunists and an aggressive
Sentinela. Dead bots do not respawn. Sentinela starts at1.2 times the human's
mass, bounded180–500. There is no requirement to consume it to win.

A new run creates a new deterministic seed, clears mutations, enemies and run
statistics, and preserves only the anonymous session's mass record. This is a
first roguelike loop, not a full campaign, unlock tree or trained boss model.

## Contracts and persistence

- `engine.Rogue` config:3,200-unit map,1,400 food,10 viruses,30Hz.
- `expedition.Run`: seed, player, sector, phase, deadline, revision. Room saves
  deep-copy this alongside the world, identities and tactical memories.
- `GET /api/expedition`: own opaque cookie identity only, uncached,120/minute.
- `POST /api/expedition/mutation`: `{mutation,revision}`, strict2KiB JSON, same
  origin, authenticated session,30/minute, mutex-serialized phase/revision checks.
- Binary protocolv1 remains unchanged. Respawn command3 starts a fresh run only
  after failure, death or victory. Choice and running phases reject restart.
- Movement and actions stop during choice and terminal phases. Transport/pings
  remain alive. Client motion discards duplicate ticks and resets after pauses.
- Solo mode rejects friend-room codes and spectating at the HTTP boundary.

## Release and rollback

`CELLS_EXPEDITION=1` advertises and admits new runs. Default0 hides it. Existing
saves remain readable/playable with the flag disabled, allowing safe fallback.
Owner: Cells maintainer. Revisit the flag by2026-10-30 after early playtests.

Deploy0.3.1 with the flag OFF before enabling0.4.0. Both versions understand the
new checkpoint schema. Once any expedition exists, rollback must target0.3.1 or
newer; older binaries reject the Rogue engine mode. Disable new entries first,
then use the compatible prior application. No destructive SQL migration is used.

## Verification

Deterministic tests cover sector transitions, death, failure, victory, stale and
duplicate choices, all three actual physics effects, mass ledger, save/reload,
identity isolation, clean restart and a paused server tick. Four real simulations
using visible-information utility agents and full physics finished with victories
for seeds17/42/91, and death in sector1 for123 (pre-release balance). No mass was
granted and no simulation ticks skipped in this playability check. Four seeds
are not a statistically valid human difficulty or success-rate estimate.

Browser tests distinguish actual solo movement/reload from explicitly marked
rendering fixtures for choices/results. They inspect responsive bounds, keyboard
restart and axe accessibility. Native tests prove mutation/state consequences;
fixtures are not evidence of server gameplay.
