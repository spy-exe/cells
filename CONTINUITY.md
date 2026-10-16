# Cells — active handoff

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
