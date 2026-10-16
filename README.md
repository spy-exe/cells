# Cells

**Comece pequeno. Ocupe espaço.** An open source multiplayer cell arena, with an
authoritative Go simulation and an original React/PixiJS client.

Play: **https://cells.malha.app** · [Source](https://github.com/spy-exe/cells)

Classic FFA is the first gameplay release. Guest sessions, movement, food,
player consumption, split/eject, viruses, delayed merging, respawn, spectating,
leaderboard, minimap and reconnection share the same server-owned world.
AI Arena adds eight visible-information tactical bots with five selectable profiles.
Public Classic mixes bot profiles; private Classic rooms are human-only.
Real learning experiments and further modes use separate release gates;
see [scope](SPEC.md) and [continuity](CONTINUITY.md) for actual status.

## Development

Requires Node 24.18, Go 1.27.2 and PostgreSQL 17. Exact JS dependencies are locked.
Create a local `cells` database owned by your development user. Never reuse
production secrets or expose PostgreSQL to the Internet.

```sh
npm ci
export CELLS_DATABASE='postgres://cells@127.0.0.1:5432/cells?sslmode=disable'
export CELLS_ORIGIN='http://127.0.0.1:5173'
go run ./apps/game-server/cmd/cells
# A separate terminal:
npm run dev
```

For a standalone build, run `npm run build`, set `CELLS_ORIGIN` to the actual origin
(default `http://127.0.0.1:8080`) and run the Go binary. Vite's development proxy
forwards `/api` and `/ws` to port 8080. Go serves `dist/web` in production.
Database initialization is additive and idempotent.

## Validation

```sh
npm run lint
npm run typecheck
npm test
npm run build
CELLS_TEST_DATABASE="$CELLS_DATABASE" go test -race -coverprofile=coverage-go.out ./...
python3 scripts/critical-coverage.py
python3 -m unittest discover -s tests/python -v
go vet ./...
npx playwright install chromium
npm run test:e2e
```

Database tests explicitly require `CELLS_TEST_DATABASE`; without it they report a
skip. Use an isolated test database. E2E requires a running built application,
same-origin API/WebSocket and a real database. Set `E2E_BASE_URL` for another origin.
Screenshots are generated in ignored `artifacts/`, traces on failure. Mobile testing
uses browser emulation, not a claim of physical iPhone/Android validation.

## Design and operations

- [Classic rules](docs/CLASSIC-RULES.md), [binary protocol](docs/architecture/protocol.md)
- [Architecture decisions](docs/adr/), [infrastructure discovery](docs/INFRASTRUCTURE.md)
- [UI system](DESIGN.md), [interaction contract](UX-CONTRACT.md)
- [Native deployment](infra/deploy/), [continuation state](CONTINUITY.md)

Native Debian services run in a dedicated unprivileged LXC. Successful `main` CI
builds produce checksummed artifacts; a reviewed root-owned deployer verifies and
activates them. Public repository code never runs on a privileged homelab CI runner.
The application, database and tunnel use separate service identities. Game checkpoints
are stored every 30 seconds and at orderly shutdown; abrupt failure can lose at most
the last checkpoint interval of transient arena state. Backups and rollback are
project-scoped and never restore old data over newer saves automatically.

## License and originality

AGPL-3.0-only. Cells uses original branding, vector shapes and game assets. It is
inspired by the classic cell-eating genre and is not affiliated with Agar.io.
Third-party notices ship with each release; local fonts are under OFL and PixiJS,
React and other dependencies retain their applicable licenses.
