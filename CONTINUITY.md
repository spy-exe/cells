# Cells — active handoff

## Teams candidate —2026-10-16 22:05 America/Sao_Paulo

Candidate0.5.0 on feat/teams, from live0.4.1 /b503795. New CELLS_TEAMS=1 gate,
30bots(10each faction,2each profile per faction), server-balanced humans, saved
allegiance, friendly consumption protection, three-row collective scoreboard and
faction-colored ejected mass. Default flag0 hides new entries only. Existing
schema/binaryv1 already support Teams;0.4.1 is a compatible rollback application.

19TS/lint/typecheck/build passed. Full Go race/DB/vet passed after faction metadata
and checkpoint assertions. Core90.6%, expedition95.8%, protocol93%, bot90.2%,
training90%, store87.5% statements. Two actual two-client browser tests plus two
UI/axe/color fixtures passed across desktop/mobile. Initial0bot population and
wrong ejected faction were reproduced RED before fixing. Read game-design/teams.md.

Next: commit/push feature, stop QA before final load; build exact committed native
binary and run25/50/100 loopback clients via cells-load-mode=teams (GOMAX2,
server150%CPU,client50%,fresh own DB,30seconds/stage,QA/AI lab stopped). Record
actual data; do not call targets measurements. Only then push main after gates,
back up own protected env, enable CELLS_TEAMS=1, guarded deploy, verify exact public
revision/modes, run28 public E2E cases and attach semantic release/proofs. Remove
only cells-qa nft rule/free8081 before candidate deployment.

0.4.1 mainCI38988270840/public24/24 passed; release contains build/checksum/report/
screenshots/continuity.0.4.0 release has actual public Rogue victory/mutations/
service-restart/new-seed proof and native backup restore into a disposable DB.
CT221 remains stopped/PPO rejected. User requires prompt validated live increments.
Older handoffs below are historical, not evidence of current production SHA.

## Absorption follow-up —2026-10-16 21:50 America/Sao_Paulo

Candidate0.4.1 on feat/consumption-motion extends conservative client absorption
feedback to captured cells, same-owner merges and ejected mass. Labels follow the
shrinking body and remain below its absorber. Ejected sprites use their last
rendered location. Tick resets/identity changes block false effects. No Go gameplay,
DB or binary-protocol change.19TS passed; real WebGL merge-caption regression was
RED on public0.4.0 (31 old-location pixels), GREEN in QA (0), along with existing
center/depth cases on desktop/mobile. Follow docs/releases/0.4.1.md and exact CI.

Production before this candidate is0.4.0 /e7ac8a0, verified mainCI38976489559 and
22/22 public browser cases. Semantic v0.3.1/v0.4.0 include verified builds, actual
screenshots, measured load and detailed continuity/proof assets. A real public
Rogue run won all3 sectors, chose digestion twice, preserved victory/mutations
across service restart, then restarted cleanly with a new seed. Backup restored
in an isolated DB and actually loaded/validated by native0.4.0; clone removed.
Prometheus jobcells up1. Root rollback baseline97e64bc was compatible with Rogue.

User requires every validated increment published promptly. Finish this small
release: commit/push, mainCI, stop QA/remove only cells-qa nft rule to free8081,
trigger cells-deploy, verify exact public SHA, run relevant public/browser smoke,
attach evidence/next prompt to semanticv0.4.1. Offhost backup/Safari/physical devices/
WAN soak/future community/Teams are pending; prior PPO was rejected, CT221 stopped.
No new training is implied by this visual patch. Historical handoffs follow.

## Current organic release —2026-10-16 19:33 America/Sao_Paulo

Production **0.3.1 /97e64bcd018a828292a2856f70baaab622dff1fe** was verified at
`https://cells.malha.app/readyz` after guarded deployment19:32:42. MainCI
38976082002 succeeded, including22 desktop/mobile browser cases and the expanded
native/TS/security gates. `api/status` advertises Classic and AI Arena only; new
Expedição entries remain OFF until0.4.0 activation. Actual browser cell3625/mass36
survived0.3.0→0.3.1. Hold process `.runtime/hold-organic.mjs` is keeping that client
connected for the next upgrade; evidence is in `artifacts/organic-upgrade-proof.json`
and `/tmp/cells-organic-upgrade-hold.log` (no cookies).

Organic membranes, food absorption, centered adaptive Unicode-safe names,
mass ordering, outline-free viruses, cursor falloff,24 public Classic/32 AI Arena
bots and evenly scheduled5Hz decisions are implemented. Private Classic remains
human-only. Expedição is a real isolated seeded three-sector solo module with
mutations and saves. Rules in docs/game-design/expedition.md. Both0.3.1 and0.4.0
understand Rogue checkpoints. Never roll back those saves to older than0.3.1;
disabling the new-entry flag keeps existing runs playable.

Final native30-second25/50/100-client benchmark, source4fb3db7:100humans+
224bots,29.958ticks/s,p99tick30.993ms,RSS32.8MB,0disconnects,0drops. Raw JSON in
`docs/performance/0.3.1-load.json`; not WAN/soak/hardware-FPS measurement. Actual
four full-physics expedition playability seeds:17/42/91 win,123 dies in sector1.

Next executable steps at this note: finish public0.3.1 browser suite and tag;
merge the prepared release/0.4.0 after the compatible baseline, wait mainCI,
deploy via existing root helper, safely back up own /etc/cells/cells.env and set
CELLS_EXPEDITION=1, restart own Cells service, assert exact public revision/modes,
run full22-case public suite plus actual Rogue checkpoint/new-run proof and
backup/isolated-restore. Attach final evidence/continuation to semantic releases.
QA/load services are stopped and the sole nft rule tagged cells-qa(handle18) was
removed. Candidate8081 is free. CT221 remains stopped; do not retrain or rebuild
its frozen source/binary as a routine deploy. Prior PPO failed promotion.

User preference: **publish every validated increment promptly** so they can follow
the live game. Commit/push main only with the relevant gates; the native deployer
accepts successful main/push CI artifacts and polls every5minutes. Trigger its own
`cells-deploy` service after CI to avoid waiting for the poll. No shared services,
credentials, global Proxmox settings or corporate environments may be changed.

Older handoffs below are historical; verify current public /readyz and the release
assets before assuming that this in-repository note is the last deployed SHA.

## Learning and polish — 2026-10-16 17:44 America/Sao_Paulo

Current delivery is0.3.0 on feat/learning-lab. Production was0.2.2/c11026d when
this note was written; confirm publication from public /readyz and releasev0.3.0.
Implementation367ee36 passed full feature CI38963476202, local14/14E2E plus final
6/6feedback cases,10TS,7native Python contracts,4deployment boundary tests,
Go race/integration/vet, critical statement coverage at least90%, npm audit0 and secret
pattern scan. New native per-round stats, authenticated /api/run, elastic cells,
actual collection feedback, gesture-enabled synthesized sounds and1.5sfood-goal
commitment are implemented. No destructive DB/protocol migration.

PPO training ACTUALLY completed in CT221: source60f09e81c1008feecdbd0d4d4f589333393bcc1d,
32768steps plus verified512step resume=33280. Model checksum changed after resume.
300held-out episodes (100each random/utility/PPO), nine exact-hash action replays.
PPO mass-area2340.20 vs9533.49utility, ratio.24547, paired95%CI.21509–.27641;
failed promotion. Production remains utility bots, now with independently tested
objective memory. Frozen benchmark is pre-goal-memory policy, compact3200arena,
not the production6000map. See docs/ai/2026-10-16 and docs/releases/0.3.0.md.
Model/native simulator/licenses/replays/resolved dependencies are in ignored
artifacts/ai/cells-ppo17-20261016.tar.gz for a GitHub release attachment, with
AI-SHA256SUMS. Raw data and real plots are versioned; no fabricated training UI.

CT221 now has CPU torch2.14.1+cpu/SB32.9.0/Gym1.4/PettingZoo1.27/NumPy2.5.4;
optional matplotlib3.11.2 was installed after training to render reports. Venv/code
root-owned, runs owned by cells-ai. Source and native binary remain frozen at the
first experiment (source.json60f09e8, simulator4b5e7acecf02c656ae57c9604916ef2a07f50dc111acd2e8128520be858f4baa).
Do not blindly sync/rebuild and then claim the old model is resumable: its source
and binary hashes are deliberately enforced. No units are training now; shut down
CT221 after copying verified artifacts. Stop QA/remove only cells-qa nft rules
before production deployment (candidate needs8081). Privileged deployer unchanged.

Next executable steps: finish report commit/push; successful final CI; fast-forward
main and wait its CI artifact; stop QA/remove rules; native guarded deploy; verify
exact public SHA, full public Playwright and authenticated run stats. Attach actual
PPO archive/checksum/screenshots to semantic release. Record final evidence there.
Then improve experiments with new validation seeds/longer horizon/production-size
arenas, compare escape planning and profile outcomes, prepare Teams/community.
Off-host backups, physical Safari/phone/WAN-soak and authenticated Grafana remain
open. No corporate systems or unrelated projects changed. Older notes below are
historical; never redevelop the completed foundations or claim model promotion.

## Controls follow-up — 2026-10-16 16:54 America/Sao_Paulo

Patch0.2.2 on branchfix/action-availability, code1507d39 plus documentation:
split/eject button eligibility now checks individual cell mass and the16cell cap,
with visible36/32per-cell requirements. Two18cells total36 cannot act; previous
buttons were enabled despite requestAction/server correctly rejecting the action.
New E2E was actually RED on public0.2.1, then GREEN on the fixed UI. Full local
suite12/12passed, TS8passed, lint/typecheck/build passed. Mobile390x844 and desktop
screenshots inspected; controlled-fixture image committed with explicit caption.
See docs/releases/0.2.2.md and its GitHub release link for publication evidence,
and public /readyz for active revision. No engine/rules changes in this patch.
At the time of this note, feature CI is running; publish only after its gate,
then main gate and real public smoke. Do not redevelop the already-fixed controls.

## Latest verified state — 2026-10-16 16:43 America/Sao_Paulo

**Cells 0.2.1 is live**, initial deployment from main revision
**5a01e3318d47b3a5a2e88913add95a93a2c2a9b7**. Check public `/readyz` for the
current artifact: later documentation-only commits may advance its revision.
Main CI **38956116138 succeeded**, feature CI **38922728307 succeeded**.
Public HTTPS/WSS browser run `/tmp/cells-public-0.2.1.log`: **10/10 passed**
(desktop+mobile emulation, 1.1min). Latest local full run also10/10. Public
Classic tests use two independent real clients; AI Arena observes real bots.
Two fixture cases cover only large-cell rendering and held input; distinguish
them from authoritative gameplay. Actual public extra smoke measured W36→22
and split preserving36mass, distance/summed-radii2.08 on its first result frame.
Raw two-sample timings75/81ms are not latency percentiles or input-to-photon.

User feedback is implemented and published: sharp vector cells, 3600food with
three masses/colors, 48virus tips, immediate input/heldW or touch eject,
split200ms/eject100ms, longer separated split launch, shared interpolation clock,
growth/spawn/consume/button transitions and native high-refresh rendering.
Server snapshots15Hz+actions, visibility cache, async checkpoints. Names recover
opacity after leaving/reentering view. Eight tactical bots/five profiles are live
in AI Arena; public Classic mixes profiles and private Classic is human-only.
**No trained model exists.**

Final Go race statement coverage: engine91.4%, protocol93.0%, bot89.5%, store87.5%.
Eight TS and four Python tests passed, plus lint/typecheck/vet/build/dependency and
configured-secret-pattern gates. See docs/releases/0.2.1.md for full report,
versioned real screenshots, test limitations and continuation prompt. No physical
phones/Safari/high-refresh hardware were measured; do not claim universal60FPS.
Final25/50/100-client benchmark data is in docs/performance, with100+56bots at
29.997ticks/s,p99tick19.156ms,0disconnects,1supersededqueued snapshot; local loopback,
not a WAN/WSS capacity guarantee. Earlier benchmark numbers below are historical.

Deploy root helper verified manifest/checksum/candidate and made a backup before
activation16:38:54. Scheduled backup+isolated restore verification manually rerun
at16:39:15 succeeded. Prometheus jobcellsup=1 after deploy. App/tunnel/PG and
deploy/backup timers active. QA/load services stopped; **all cells-qa temporary
firewall rules removed**. Candidate port8081 free. Do not rebuild/reprovision to
rediscover these facts. CT221lab is stopped; resources1CPU/cap.75,1.5GiB,8GiB,
10.10.1.221; no Torch/SB3 or model yet. Its provisioning script is committed and
includes firewall/SSH controls; do not run it against the existing CT.

The original mission remains open. Next: headless authoritative environment,
real bounded PPO in CT221,100held-out-seed episodes per random/heuristic/learned
policy and honest promotion decision; then Teams/community. Off-host backups,
DHCP reservation, authenticated Grafana integration and physical-device/WAN/soak
validation remain pending. Production remains heuristics until evidence supports
any learned replacement. Do not revisit Tribunal/OUTRAVIDA or corporate systems.

## Earlier chronological notes — superseded where stated above

The user's active mission is Cells. Do not resume Tribunal or OUTRAVIDA. The user
explicitly authorizes public GitHub, isolated personal infrastructure and incremental
production releases. Continue beyond foundation; do not report the whole mission done
while Classic, tactical bots/AI Arena, real PPO training/evaluation, Teams/community,
load tests and operations remain viable.

## Production foundation

Repository: https://github.com/spy-exe/cells (public, AGPL-3.0-only).
Stable main foundation: aa1fbcd489b6e37146d43beac3cfabf7f9ace566, version 0.0.1.
Public URL https://cells.malha.app passed HTTPS readiness and HTTP 308 redirect.
Actual public foundation E2E: 2 passed (desktop and mobile). A PostgreSQL dump was
restored into cells_restore_verify, queried (cells_schema count 1) and the temporary
clone removed successfully. Native automatic deploy is active every five minutes.

Personal node pve, Proxmox 9.2.4. Dedicated unprivileged CT 220 cells-prd-01,
Debian 13.6, 4 vCPU/cpulimit 2, 6144MiB RAM, 512MiB swap, 40GiB local-lvm,
vmbr0, 10.10.1.220/24, gateway 10.10.1.254, timezone America/Sao_Paulo.
No nesting. Do not alter other projects, corporate environments or global PVE rules.
Host root only has about 2.5GiB free; Go/Python heavy builds belong in Cells-owned
CT storage. No unrelated cache or volumes may be cleaned. DHCP reservation is not
proven; the static address followed existing service ranges and passed ARP checks.

Native PostgreSQL 17, peer-authenticated cells role/database, loopback/socket only.
App user cells, tunnel user cells-tunnel. Cloudflare dedicated Cells tunnel and DNS;
see docs/INFRASTRUCTURE.md. Never print credentials or origin certificates.
/etc/cells/cells.env, /etc/cells/cloudflared.yml and tunnel.json are private.
Service units: cells, cells-tunnel, cells-deploy.timer, postgresql, nftables.
Reviewed deploy helper /usr/local/libexec/cells/pull-release.py downloads only a
successful main CI artifact, verifies its digest/manifest, backs up DB, starts a
nonprivileged candidate and atomically activates with local/public readiness. A
rollback changes application symlinks, never overwrites newer DB state.

## Current work (Classic)

Branch feat/classic-engine extends main. Check git status before work. Engine and
binary protocol are committed separately; remaining integration/UI changes may be
uncommitted. Do not merge main until full hosted CI succeeds. GitHub Actions uses
hosted Ubuntu, not a privileged self-hosted public-repository runner.

Engine tests were RED before implementation. Initial Go race run: engine 89.1%
statement coverage, protocol 88.7%, server 63% (package-local instrumentation).
Classic integration covers guest origins, strict JSON, secure cookies, separate
identities and invalid WebSocket input. Store integration caught abandoned rooms
being resurrected after save: actual RED; fix replaces persisted room keys atomically.
A follow-up race run with cross-package instrumentation is being executed; read
/tmp/cells-classic-go-final.log and inspect results before claiming success.

UI: React/TS/Vite, explicit Pixi WebGL renderer, custom textures, grid, interpolation,
keyboard/mouse/touch, HUD/minimap/ranking, death/respawn and reconnect. Visual QA
caught blank canvas despite working HUD. Fixed using Pixi's CSP-safe polyfill module
pixi.js/unsafe-eval (does not relax CSP), explicit renderer and separate render errors.
E2E now requires canvas plus nonzero measured FPS. Four E2E tests passed on isolated
QA after that fix; see /tmp/cells-classic-final.log. Screenshots in artifacts/ were
actually inspected. Latest UI resume/quality changes need a rebuilt QA and E2E.

QA: cells-qa.service at 10.10.1.220:8081, DB cells_qa, native binary /opt/cells/qa-binary,
web /opt/cells/qa-web. nftables temporary rules tagged cells-qa allow only host
10.10.1.253. Remove ONLY these rules and stop QA before deployment: candidate uses
8081. There is an explicit drop near the end of input; QA allow must precede it.
Never run two Playwright processes writing to the same test-results directory.

Go toolchain /opt/cells/toolchains/go/bin/go inside CT (not on host). Use PATH with
/usr/sbin for runuser. infra/deploy/sync-build.sh copies own sources excluding secrets
and generated deps to /opt/cells/build. Format Go there and pull formatted files back.
Node 24.18.0 already exists on host. Chromium for Playwright:
CHROMIUM_PATH=/root/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome.
Commands npm run lint, npm run typecheck, npm test, npm run build, npm run test:e2e.
Public E2E uses E2E_BASE_URL=https://cells.malha.app. Preserve original screenshots.

## Next executable steps

1. Complete latest race/vet/store regression and UI E2E; inspect screenshots/console.
2. Ensure candidate helper sets CELLS_CANDIDATE=1 so it cannot save production rooms.
   Install reviewed helper only after its tests. Commit remaining Classic increment,
   push feature branch, wait CI, fast-forward main, wait main release artifact.
3. Stop QA/remove own firewall rules, deploy verified artifact, public two-client WSS
   E2E and exact SHA readiness. Update this handoff with actual version/commit.
4. Tactical utility bots using only public engine.Observation (five profiles), AI Arena.
5. Separate Cells-owned training CT (221 was free but NOT created); remeasure host
   memory/storage, cap CPU/memory. Real SB3 PPO, headless authoritative Go simulator,
   100 held-out episodes comparing random/heuristic/trained, honest report and baseline
   in production unless learned model improves. No heavy training on game server.
6. Teams, public community, measured load 25/50/100, Prometheus scrape integration,
   daily backups/isolated restoration and rollback demonstration. Other projects untouched.

Do not spawn agents without explicit user/applicable skill instruction. Old
social_reproduction_tests belongs to OUTRAVIDA and is irrelevant. Communicate concise
Portuguese progress at least every minute. No secrets in logs/commits. Do not claim
bots, training, modes or performance until actually implemented, executed and measured.

## Update 2026-10-16 — live Classic and player feedback

Production is **0.1.0 / 2763fe2f292fe739fbfb047194fe7a00dbd6aa78**, not the old
foundation SHA above. Public two-client desktop and mobile WSS smoke passed. Main
CI 38913970988 succeeded. Prometheus CT200 job cells is live and up=1; config backed
up at /etc/prometheus/prometheus.yml.before-cells, promtool validated, safe reload.
Daily backup/isolated restore succeeds. Actual foundation rollback/resume was tested.
See docs/releases/0.1.0.md and docs/OPERATIONS.md for recorded evidence.

Current branch **feat/tactical-arena**, remote head **82a869b**. Tactical bots and AI
Arena are implemented (five profiles, same visible observations as humans), not yet
published. First CI 6d40a00 failed an assumption about enemy visibility; fixed via
legitimate spectator E2E (42e8457 CI succeeded). 82a869b CI then exposed a real logout
race: menu rendered before session revocation, so reload could auto-resume. Current
uncommitted fix awaits logout before menu. Do not rerun/merge the failed CI unchanged.

The latest user played production and requests sharp large cells, denser/brighter
variable food, double virus tips, smooth movement, immediate Space/W, better split
separation/range and purposeful animations. **Prioritize and publish this feedback
before training.** Work is in the same branch, planned version **0.2.1**. Latest code:
- vector cell GraphicsContext instead of enlarging a 128px bitmap; virus 48 tips;
- food default 3600 (was2000), mass1/2/3, stronger colors and rare-pellet subtle pulse;
- buffered server-time interpolation with bounded extrapolation (motion.ts), camera
  follows interpolated state; mouse direction uses current authoritative centroid;
- immediate action send, held W/touch eject, split200ms/eject100ms with single buffered
  commands, stronger/farther non-overlapping split launch;
- snapshot15Hz plus immediate action snapshots; asynchronous checkpoint I/O and cloned
  data outside tick lock; cached spatial visibility and aggregate masses per tick;
- consume fade, launch/growth transitions, input feedback, reduced-motion support;
- await logout before returning to the entry screen.

Tests: actual initial RED for split reach/food diversity and missing motion module,
then green. Eight TS unit tests currently pass. Last full Go race gate before the
latest small changes: engine91.1%, protocol93%, bot89.5%, store87.5%. New checkpoint
blocking-store regression passed: ticks continue while persistence waits. Ten local
E2E passed (desktop+mobile, including real multiplayer and a clearly identified
controlled large-cell/input fixture). Read /tmp/cells-feel-e2e.log. Newer hold-button
assertion and latest bot reach/eject-speed changes still need final gate/build.
Screenshots inspected: artifacts/fixture-large-cell-desktop.png (controlled fixture,
not a real player record), artifacts/qa-arena-desktop.png and mobile equivalents.

Uncommitted files include apps/game-server/cmd/load, motion.ts, feel.spec.ts, lab
provision script and current feedback changes. Do not lose them. Commit small units,
push, wait feature CI green, then ff main, main publish, deploy exact artifact and
public smoke. Stop cells-qa and remove only rules tagged cells-qa before deploy;
candidate needs8081. Root helper already installed with CELLS_CANDIDATE=1.

Benchmark: load client/server run as own native transient units on CT220, loopback
8082/metrics9109, DB cells_load. Synthetic RFC5737 CF-Connecting-IP rate buckets only
on loopback simulate distinct clients; this is not an external 100-IP/WSS test.
Clients25/50/100 plus bots, real binary WS and HTTP sessions. First20s stages:
100 clients+48bots, zero disconnects/drops,29.845ticks/s,p99tick24.098ms, RSS24.8MB.
It exposed reservation overfill; regression RED, fixed in82a869b. Second30s stages:
25+16bots29.966ticks/s,p99tick8.015ms;50+32bots28.799ticks/s,p99tick12.513ms;
100+56bots28.730ticks/s,p99tick22.103ms,RSS25.5MB, zero disconnects/drops.
That fall corresponded to synchronous checkpoint I/O; current feedback fix is async.
**Rerun final benchmark with new physics/food/snapshot/cache/save code**, record actual
results (do not present target30 as achieved). Prior logs are in journal units
cells-load-client and cells-load-client-final. Load server is stopped now. QA may be
active on8081; inspect systemctl and own firewall rules. Only one benchmark/browser
run per output directory; previous overlap corrupted Playwright artifacts.

Additional Cells-only lab was provisioned: **CT221 cells-ai-lab-01**, Debian13.6,
unprivileged/no nesting,1vCPU/cpulimit0.75,1536MiB RAM,256MiBswap,8GiB local-lvm,
10.10.1.221/24/gw.254, timezone SãoPaulo,onboot0. Resource checks: host7GB available
and thinpool~50GB physical free beforehand; CTIDfree and no ARP reply. Optional failed
mounts masked as CT220. python3-venv installed, /opt/cells-ai/venv, cells-ai service
user and own dirs. **No Torch/SB3, training or model yet.** SSH disabled, nftables
cells_lab inputDROP (lo/established/ICMP/ND only), outputALLOW. Provision script currently
misses final firewall/SSH manual steps: add before committing for reproducibility.
Keep lab stopped while idle; finish feedback release first. PyPI confirmed SB3 stable
2.9.0 requires Torch>=2.8,<3, Python>=3.10. Use official CPU Torch index, avoid CUDA
wheels. Original mission still includes real PPO+100episode comparison, Teams,
community/suggestions/moderation, CI Python tests, final benchmarks and full report.

## Feedback gate update — 2026-10-16 12:09 America/Sao_Paulo

Latest executable candidate is **0.2.1**, SHA **8bd24e8288e89a079d619778172418fb4f27afe0**.
Committed and pushed feedback fixes: dd25f02 engine/hub, ccf627c web/motion/tests,
7002e4d load/lab tools, 8bd24e8 version/rules. Hosted feature CI **38921286452
SUCCEEDED**, including the complete ten-test desktop/mobile E2E suite. Latest Go
race coverage: engine91.4%, protocol93.0%, bot89.5%, persistence87.5%; coverage is
statements, not branches. Eight TS unit tests, four Python deploy tests,
lint/typecheck/build/vet/secret/dependency gates passed. npm audit found0 issues.
One local AI E2E initially failed because it required a transient first leader to
stay visible; live binary-frame inspection showed actual bot movement and leader
changes. It now tracks movement across all received visible frames, and passed six
repeat desktop/mobile runs. All scenarios then passed together in hosted CI.
The latest vector-circle fixture screenshot has been visually inspected; it is
an explicit controlled rendering fixture, not an actual public player record.

Final benchmark is now **completed and recorded** in docs/performance/0.2.1-load.json:
25+16bots29.998ticks/s,p99tick8.022ms;50+32bots29.998,p9912.422;
100+56bots29.997,p9919.156,RSS34.4MB,0disconnects,1replacedqueued snapshot.
Each stage30s, real HTTP session/binaryWS loopback, not external WSS or WAN latency.
QA and CT221 were stopped; shared homelab left intact. Test DB
cells_load_feel_20261016 is isolated and retained for investigation. Load services
are stopped. Temporary firewall rule taggedcells-qa was removed (handle14).
The AI lab script now includes its SSH-disable and firewall configuration and is
committed. CT221 is stopped, with no trained model or training dependencies yet.

Next required action: fast-forward main **to verified8bd24e8**, wait main CI/release,
trigger project deploy, verify public /readyz exact SHA and run actual public
Classic+AI E2E desktop/mobile. Additional documentation on this feature branch may
be newer than the executable candidate; do not claim a newer SHA is active until
readiness proves it. Production is still0.1 until that activation is verified.
Continue original mission afterward: real bounded PPO +100episode evaluation,
Teams/community and remaining ops gaps. Do not claim learned bots already exist.

## Organic release work —2026-10-16

Branch `feat/organic-expedition` starts from0.3.0 d128267. Commits5a4e764 and
2fc4c1e add organic vector membranes/absorption/centered names/layering and cursor
falloff plus24/32 bots. Expedição is a new isolated finite Rogue module, whose
rules/API/rollback boundary are documented in `docs/game-design/expedition.md`.
Publish0.3.1 with new-entry flag OFF first, then0.4.0 with `CELLS_EXPEDITION=1`.
Both understand Rogue saves, so rollback from0.4 must use0.3.1 or newer.
Production publication is confirmed only by exact `/readyz` and GitHub CI artifact.
