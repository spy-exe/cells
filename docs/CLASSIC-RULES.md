# Classic rules — engine v1

- Arena: 6000 × 6000 world units. Spawn mass 36. Radius = 4 × sqrt(mass).
- Movement: 220 × (mass/36)^(-0.25) units/s toward the target. Map boundaries apply to the whole radius.
- Eat another cell at ≥1.18 times its mass, when its center is within own radius minus 35% of the prey radius. Equal cells do not eat each other.
- Split: minimum mass 36, half to each cell, directional impulse, 0.8s action cooldown, maximum 16 cells. Recombination requires overlap after 12s + a mass-dependent delay. Cells repel during the merge delay.
- Eject: minimum mass 32, spend 14 per eligible cell, emit 12, lose 2; 0.2s cooldown. Own freshly ejected pellets have a 0.5s collection grace. Pellets expire after 30s.
- Viruses: small cells pass through them. A cell with ≥1.3 times the virus mass consumes it and bursts into up to 16 cells. Seven fed pellets launch another virus; transferred mass is conserved.
- Large-cell decay removes 0.15% of the mass above 36 every second. Food is replenished; replenishment is counted as a source, not duplicated consumed food.
- Death occurs when the last cell is consumed. Respawn opens after 2s. Disconnected bodies remain vulnerable for 30s. There is no spawn shield or paid advantage.
- The view grows with mass and is capped. Public leaderboard shows the top ten, identifies bots and is independent of view radius. No enemy radar on the minimap.

Numbers are game rules, not claims of matching another implementation. Seeds,
input streams and versioned checkpoints support reproducible tests. Later modes
reuse physics and change rule policies, such as friendly consumption in Teams.
