# Protocol v1

Binary WebSocket, little endian, one message per frame. First bytes are `CE 01 <kind>`.
The HTTPS origin and HttpOnly guest cookie are validated before upgrading. Clients
send intentions only. No packet accepts mass, speed, identity or a cell position.

| Kind | Length | Payload |
|---|---:|---|
| 1 input | 16 | uint32 sequence at 3, float32 target X at 7, Y at 11, uint8 actions at 15 (split=1, eject=2) |
| 2 snapshot | bounded | header below, visible entities and top ten |
| 3 respawn | 3 | server validates death and delay |
| 4 spectate | 3 | dead players already follow the leader |
| 5 ping / 6 pong | 11 | opaque float64 client timestamp at 3, echoed for RTT |

Snapshot header: tick uint32 at 3, own player uint32 at 7, flags at 11 (alive=1,
Teams=2), float32 size/X/Y/view/mass at 12/16/20/24/28, uint16 remaining split,
eject and respawn ticks at 32/34/36, three float32 team masses at 38/42/46,
uint16 entity count at 50, uint8 rank count at 52. Header length is 53 bytes.
Entity: uint32 ID/owner, float32 X/Y/mass, uint8 kind/team/bot/name-byte-length,
then UTF-8 name. Fixed portion 24 bytes. Rank: uint32 ID, float32 mass, uint8
bot/team/name-byte-length followed by UTF-8 name (11-byte fixed portion).

Input maximum 64 bytes, 60 messages/s including keepalive, strictly increasing
sequence per connection. Unknown versions/actions, text messages, nonfinite values
and malformed lengths close the connection. Snapshot maximum 2048 entities,
10 ranks and 80 UTF-8 bytes per name. Clients also check lengths and finite values.

Simulation: 30 fixed steps/s, snapshots every second step (15/s), with an extra
snapshot on ticks that execute split or eject. One writer per
connection; at most one pending snapshot and four control replies. A slow client
loses obsolete snapshots rather than accumulating a queue. Visible state includes
only a bounded circular view, own cells and public leaderboard. Enemy cooldowns,
world RNG and entities outside the view are never transmitted. Rendering
uses a monotonic buffered server-time clock, targets 65ms of interpolation delay
and bounds extrapolation to 60ms. Camera and entities share the same clock. The
client sends movement every 50ms and sends action input immediately. Holding W
or the touch eject button repeats intentions; the server enforces all cooldowns.
Rendering has no authority over collisions.

Disconnect keeps the vulnerable body for 30 seconds. Reconnecting with the same
cookie resumes its identity. A second connection replaces the first. Anonymous
sessions expire after 30 days and can be revoked through logout.
