# ADR0005 — organic client and finite solo rules

Status: accepted,2026-10-16.

## Context

Cell motion felt rigid, small captions detached from their bodies and sprite
insertion order could make small players appear over larger predators. Players
also asked for a separate roguelike experience while preserving Classic.

## Decision

Render each cell as a vector96-segment triangle fan. Low-amplitude seeded radial
waves normalize polygon area; movement stretch is bounded and area preserving.
This avoids scaling a blurry raster circle and keeps the collision circle clear.
Labels share their body's mass-based depth and remain centered. Text fitting uses
screen radius, abbreviating safely and retaining the AI identifier when legible.

A short food absorption effect is inferred only when a disappeared visible food
has a nearby, already present growing cell with sufficient mass gain. It is not a
server consumption event, is bounded to32 effects per snapshot, and can miss or
misattribute simultaneous hidden consumption. No new trusted client state or
protocol change is introduced. A future compact eat-event packet could replace
this inference if measured ambiguity justifies it.

Rogue is an engine rules mode composed with a small `expedition.Run` state machine
in the room coordinator. World ticking pauses between finite sectors; transport
continues. Mutation choices are session-scoped revision-checked HTTP commands,
serialized under the same room mutex as ticks. Existing binary motion snapshots
continue unchanged. The state machine and every persistent physics modifier are
saved with the authoritative world; no client timer can end a sector or grant a
mutation. There is no arbitrary rule scripting.

## Consequences

The client updates membrane vertices at bounded cadence, with a lower cadence in
economy mode. Reduced motion restores circles and skips absorption travel. Mesh
geometry must be explicitly destroyed when entities retire. Physics and bot
perception continue using the same exact circular server geometry.

The first solo loop has three25-second sectors and two upgrade choices. It does
not provide cross-mode progression or any Classic advantage. New rooms consume
bounded resources like other rooms and share the existing room limit. Further
modes should compose explicit rules modules instead of copying the simulation.

Deploy a compatible disabled-flag release before admitting Rogue sessions, so
application rollback preserves readable saves. Older-than0.3.1 binaries are no
longer valid rollback targets once these checkpoints exist. The feature flag
controls new entry, not destruction or abandonment of existing runs.
