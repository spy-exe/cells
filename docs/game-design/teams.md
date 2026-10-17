# Teams

Three factions—Musgo, Coral and Maré—compete for the total mass of their living
cells. This is a continuous arena, with no timer or declared round winner. A
player's physical mass remains their own; the scoreboard aggregates their team.

New arenas start with30 tactical bots:10 per faction and two of each of the five
programmed profiles in each faction. Humans enter the least-populated living
team. Membership persists on respawn and checkpoints; existing factions are never
reassigned to rebalance a save. Existing-room backfill respects living-player
balance, so its bot-only counts need not equal10 each.

Allied players cannot eat one another, and cells belonging to different allies
do not recombine. Other Classic movement/split/eject/virus rules remain. Ejected
mass can feed any eligible cell and carries its source faction color. The team
bot policy excludes allies from both threats and prey; it is programmed utility
behavior, not a learned cooperative policy.

The three-segment scoreboard reports the shares of current living mass; the
player's faction remains identified after death. Names/percentages supplement
color, and reduced motion disables bar transitions. The Classic and AI Arena
leaderboards are unchanged.

`CELLS_TEAMS=1` enables new entries and advertises the mode. Default0 hides new
entry without disabling existing saved sessions. The server chooses allegiance;
public JSON cannot assign `team`. Teams uses the already-versioned binaryv1 team
flag/team-mass fields and existing persisted entity/player faction fields. No SQL
migration or arbitrary rule scripting is involved.0.4.1 understands saved Teams
worlds and remains a compatible fallback, with the older individual leaderboard.

Verification covers ally protection, profile/faction balance, save/reload,
metadata on ejected mass, actual two-client browser sessions, proportional score
bars, own-faction retention after death, screenshot color and axe accessibility.
A final load gate measures25/50/100 real sessions with `cells-load -mode=teams` on
isolated loopback. Public activation follows that gate and successful main CI.
