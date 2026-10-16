# Reproducing the initial load measurement

`0.2.1-load.json` records an actual 2026-10-16 run, not a capacity guarantee.
The Go runner creates real PostgreSQL-backed guest sessions and binary WebSocket
connections against an isolated server/database. It opens 25, then 50, then 100
clients, for 30 seconds per stage. Each public AI room adds eight tactical bots.
The server uses the production CPU/memory limits, 3600 food particles per room,
30Hz simulation and 15Hz base snapshots plus action frames.

| Clients + bots | Ticks/s | Tick p99 ms | RTT p99 ms | RSS MB | Disconnects | Replaced snapshots |
|---|---:|---:|---:|---:|---:|---:|
| 25 + 16 | 29.998 | 8.022 | 6.480 | 21.8 | 0 | 0 |
| 50 + 32 | 29.998 | 12.422 | 9.106 | 26.8 | 0 | 0 |
| 100 + 56 | 29.997 | 19.156 | 19.020 | 34.4 | 0 | 1 |

The 100-client stage received 45,986 snapshots and 222.7 MB, about 7.42 MB/s
of application payload. One queued snapshot was replaced by newer state; no
client disconnected or received zero frames. This exercises queue handling,
not network loss or Internet latency. RTT measures protocol ping/pong, not the
complete input-to-display latency. Tick percentiles use the server's rolling
1800-tick window; at later stages that window also includes earlier stages.
CPU usage in that stage averaged 0.62 core equivalents for the server.

The host is a shared personal Ryzen 7 3700U homelab. CT220 has a two-core cap;
the isolated server was limited to 1.5 cores and the load client to 0.5 core.
Own QA and the AI lab were stopped during measurement. Other projects were
left running. Room seeds were generated normally, so this is not a controlled
seed-for-seed comparison with earlier versions. Duration is a smoke benchmark,
not a soak or a demonstrated maximum server capacity.

```sh
go build -o bin/load ./apps/game-server/cmd/load
# Point at a separately started server and disposable database, never production:
bin/load -target http://127.0.0.1:8082 \
  -metrics http://127.0.0.1:9109/metrics -duration 30s -synthetic-clients
```

The synthetic flag supplies distinct RFC5737 client buckets to exercise the
real per-IP/session limits locally. It refuses non-loopback targets. It does not
represent 100 external IPs or Cloudflare/WSS throughput. The CLI exits on join
failure, connection failure, unexpected disconnect or a client with no snapshots.
Keep the load server, guest data and ports isolated, apply resource quotas and
stop only these project-owned transient services afterward.

Earlier runs exposed reconnect-slot overfill and blocking database checkpoints.
Both received regression tests and fixes before this run. Async checkpointing
keeps database writes and JSON encoding outside the fixed-tick loop; it does
not eliminate the need to measure longer runs and WAN behavior later.
