# Classic rules — engine v1

- Arena: 6000 × 6000 world units. Spawn mass 36. Radius = 4 × sqrt(mass).
- Movement: 220 × (mass/36)^(-0.25) units/s toward the target. Map boundaries apply to the whole radius.
- Eat another cell at ≥1.18 times its mass, when its center is within own radius minus 35% of the prey radius. Equal cells do not eat each other.
- Split: minimum mass 36, half to each cell, 0.2s action cooldown, maximum 16 cells. The child starts two child radii + 8 units ahead and launches at 1050 units/s; the parent recoils at 70 units/s. Launch velocity decays by 0.92 per tick. Recombination requires overlap after 12s + a mass-dependent delay. Cells repel during the merge delay.
- Eject: minimum mass 32, spend 14 per eligible cell, emit 12, lose 2; 0.1s cooldown. Pellets launch at 650 units/s and their velocity decays by 0.9 per tick. Own freshly ejected pellets have a 0.5s collection grace. Pellets expire after 30s. One action received during cooldown is buffered until the next eligible tick; it never repeats by itself.
- Viruses: small cells pass through them. A cell with ≥1.3 times the virus mass consumes it and bursts into up to 16 cells. Seven fed pellets launch another virus; transferred mass is conserved.
- Large-cell decay removes 0.15% of the mass above 36 every second. There are 3600 food particles, with mass 1/2/3 in seeded proportions 70%/25%/5%. Consumed food is replenished and counted as a source, not duplicated. Existing standard-map checkpoints are upgraded from the old 2000-food budget without resetting players.
- Death occurs when the last cell is consumed. Respawn opens after 2s. Disconnected bodies remain vulnerable for 30s. There is no spawn shield or paid advantage.
- The view grows with mass and is capped. Public leaderboard shows the top ten, identifies bots and is independent of view radius. No enemy radar on the minimap.

Numbers are game rules, not claims of matching another implementation. Seeds,
input streams and versioned checkpoints support reproducible tests. Later modes
reuse physics and change rule policies, such as friendly consumption in Teams.

Server safety budgets added with tactical bots: at most 512 ejected pellets and
128 viruses per room. Ejection at the pellet cap does not charge mass. Feeding at
the virus cap consumes the growth remainder as an explicit mass sink instead of
spawning an unbounded entity. Neither limit grants a client an alternate physics path.
