# Tactical policies v1

These are programmed utility agents, not trained neural networks. Both UI labels
and API mode descriptions identify them as bots. Names and visuals are original.

The policy accepts exactly `engine.Observation`, the same visible entities,
leaderboard, own state and cooldowns serialized to a human client. It never accepts
a World pointer, invisible entities, enemy commands or PRNG state. A bounded
memory estimates velocity from successive visible positions and forgets tracks
not seen for four seconds.

Each agent evaluates threats relative to its smallest body, chooses among 24
escape directions with wall/virus penalties, searches for safe food, predicts prey
motion and discounts pursuits where its mass makes it slower. Split hunting
requires a post-split mass advantage, landing range and absence of nearby threats
or dangerous viruses. Agents can shelter behind viruses when small and eject
mass toward a virus aligned with a distant larger enemy. Risk differs by profile:

| Profile | Behavior |
|---|---|
| aggressive | hunts and splits more readily, shorter safety margin |
| cautious | wider threat margin, food preference, no attack splits |
| opportunist | seeks isolated prey while preserving escape options |
| balanced | moderate pursuit and defensive utility |
| beginner | low aggression, short defensive horizon, no split hunting |

Decisions run at 5Hz; movement and collisions remain at 30Hz. AI Arena uses the
selected profile for eight agents. Public Classic rooms mix profiles; private
Classic friend rooms are human-only. Bodies and bots use identical physics,
consumption rules and cooldowns. Dead bots respawn after at least three seconds.
No paid model API or training process runs in production.

Unit behavior tests cover food, threats, safe/unsafe split, cooldowns, wall and
virus avoidance, virus feeding, team awareness, velocity prediction, deterministic
choice and memory expiry. E2E observes actual bot positions changing in AI Arena.
Coverage measured locally: 89.5% of policy statements before the release gate.
Longer comparative learning evaluation is tracked separately; these tests do not
claim a learned policy or a win-rate advantage.
