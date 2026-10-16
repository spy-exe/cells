# Cells operations

All `pct` commands run on personal node **pve**, for **CT220 cells-prd-01** only.
4 vCPU (cap2), 6GiB RAM, 40GiB local-lvm; 10.10.1.220/24 on vmbr0.
Production URL https://cells.malha.app. No Docker nesting or privileged container.

## Service and release management

```sh
pct exec 220 -- systemctl status cells cells-tunnel postgresql
pct exec 220 -- systemctl restart cells
pct exec 220 -- journalctl -u cells -n 100 --no-pager
pct exec 220 -- systemctl start cells-deploy
curl -fsS https://cells.malha.app/readyz
```

Reviewed main builds must pass hosted GitHub CI. A release contains only a Go binary,
web assets, licenses and a manifest. The deployer verifies artifact SHA256 and exact
revision, backs up DB, starts a candidate as user cells, then swaps a root-owned
symlink and verifies local/public readiness. It runs every five minutes. Public CI
holds no homelab credentials and cannot execute privileged code on the host.
Candidate uses port 8081 and CELLS_CANDIDATE=1; stop own QA before deployment.
Application artifacts cannot update the privileged deploy helper or service units.

Manual rollback uses the previously active application and holds automatic upgrades:

```sh
pct exec 220 -- /usr/local/libexec/cells/pull-release.py --rollback
# Check exact readiness and investigate, then explicitly resume:
pct exec 220 -- /usr/local/libexec/cells/pull-release.py --resume
```

Rollback changes code, never overwrites new player data. Additive schema changes and
checkpoint reader compatibility are mandatory. Keep an old release until its rollback
window closes. A migration that removes/rewrites saves needs a separate recovery plan.

## Backups and recovery

Daily 05:20 America/Sao_Paulo (with up to 5min jitter), and before deployment.
Files live in root-only `/var/backups/cells`: custom-format PostgreSQL dumps and
private configs/tunnel credentials, never public GitHub artifacts. Retention 14 days.
The scheduled backup restores a disposable `cells_verify_<timestamp>` database,
checks schema and drops only that newly-created clone. Manually rerun:

```sh
pct exec 220 -- systemctl start cells-backup
pct exec 220 -- journalctl -u cells-backup -n 30 --no-pager
```

A full production recovery requires stopping Cells, preserving the latest damaged DB
for investigation, restoring a selected verified dump to a new database, checking
its schema/checkpoints with the matching release, then switching the connection and
restarting. Never restore over a running database or assume an old schema accepts a
newer binary. Do not execute that destructive recovery without explicit authorization.

**Known limitation:** backups currently share CT storage. They demonstrate logical
recovery, not resilience to loss of the Proxmox host/storage. No authorized external
backup destination or DHCP reservation was discovered. Configure an off-host encrypted
destination before treating the service as disaster resilient. Do not put private data
in the public repository. Avoid dumps on the nearly full Proxmox root filesystem.

## Monitoring and network

Prometheus at CT200 scrapes `10.10.1.220:9108` every 15s. Its config was backed up,
validated with promtool and safely reloaded; existing jobs were preserved. App metrics
report connections, bots, rooms/entities, tick p95/p99, queue drops, bytes, rejected
frames, heap and goroutines (process CPU/RSS added in the next bot release).
The private firewall permits this port only from CT200. No DB/SSH/admin Internet port.
Public origin listens on loopback8080, reached only by the Cells Cloudflare tunnel.
Structured JSON logs omit tokens, cookies and credentials. Readiness requires PostgreSQL.

Grafana dashboard provisioning has not been claimed: the existing provider is inactive
and no authenticated management token was established. Metrics and Prometheus target
health are operational; dashboard import remains a separate optional integration.
