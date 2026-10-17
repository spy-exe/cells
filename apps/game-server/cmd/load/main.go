// load exercises the actual HTTP session and binary WebSocket boundary.
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coder/websocket"
	"github.com/spy-exe/cells/apps/game-server/internal/protocol"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type client struct {
	conn                       *websocket.Conn
	cookie, source             string
	frames, bytes, disconnects int
	intervals, rtt             []float64
}

func quantile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	return values[int(float64(len(values)-1)*q)]
}
func metrics(endpoint string) map[string]float64 {
	result := map[string]float64{}
	res, err := http.Get(endpoint)
	if err != nil {
		return result
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 100000))
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && !strings.HasPrefix(f[0], "#") {
			result[f[0]], _ = strconv.ParseFloat(f[1], 64)
		}
	}
	return result
}
func headers(origin, source string) http.Header {
	h := http.Header{}
	h.Set("Origin", origin)
	if source != "" {
		h.Set("CF-Connecting-IP", source)
	}
	return h
}
func join(ctx context.Context, target, origin string, i int, synthetic bool, mode string) (*client, error) {
	source := ""
	if synthetic {
		source = fmt.Sprintf("192.0.2.%d", i+1)
	}
	body, _ := json.Marshal(map[string]string{"nickname": fmt.Sprintf("Load %03d", i), "mode": mode, "profile": "balanced"})
	req, _ := http.NewRequestWithContext(ctx, "POST", target+"/api/guest", bytes.NewReader(body))
	req.Header = headers(origin, source)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 201 || len(res.Cookies()) != 1 {
		return nil, fmt.Errorf("guest rejected: HTTP %d", res.StatusCode)
	}
	return &client{cookie: res.Cookies()[0].String(), source: source}, nil
}
func (c *client) run(ctx context.Context, target, origin string, index int) error {
	h := headers(origin, c.source)
	h.Set("Cookie", c.cookie)
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(target, "http")+"/ws", &websocket.DialOptions{HTTPHeader: h})
	if err != nil {
		return err
	}
	c.conn = conn
	conn.SetReadLimit(200000)
	go func() { <-ctx.Done(); conn.CloseNow() }()
	var write sync.WaitGroup
	write.Add(1)
	defer write.Wait()
	go func() {
		defer write.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		seq := uint32(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				seq++
				angle := float64(seq)/40 + float64(index)
				b := make([]byte, 16)
				copy(b, []byte{protocol.Magic, protocol.Version, 1})
				binary.LittleEndian.PutUint32(b[3:], seq)
				binary.LittleEndian.PutUint32(b[7:], math.Float32bits(float32(3000+math.Cos(angle)*1800)))
				binary.LittleEndian.PutUint32(b[11:], math.Float32bits(float32(3000+math.Sin(angle)*1800)))
				if seq%100 == 0 {
					b[15] = 1
				}
				if err := conn.Write(ctx, websocket.MessageBinary, b); err != nil {
					return
				}
				if seq%10 == 0 {
					ping := make([]byte, 11)
					copy(ping, []byte{protocol.Magic, protocol.Version, 5})
					binary.LittleEndian.PutUint64(ping[3:], math.Float64bits(float64(time.Now().UnixMicro())))
					if conn.Write(ctx, websocket.MessageBinary, ping) != nil {
						return
					}
					_ = conn.Write(ctx, websocket.MessageBinary, []byte{protocol.Magic, protocol.Version, 3})
				}
			}
		}
	}()
	last := time.Time{}
	for {
		_, b, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() == nil {
				c.disconnects++
			}
			return nil
		}
		c.bytes += len(b)
		if len(b) < 3 {
			return fmt.Errorf("short frame")
		}
		if b[2] == 2 {
			if len(b) < 53 {
				return fmt.Errorf("short snapshot")
			}
			now := time.Now()
			if !last.IsZero() {
				c.intervals = append(c.intervals, float64(now.Sub(last).Microseconds())/1000)
			}
			last = now
			c.frames++
		} else if b[2] == 6 && len(b) == 11 {
			sent := math.Float64frombits(binary.LittleEndian.Uint64(b[3:]))
			c.rtt = append(c.rtt, (float64(time.Now().UnixMicro())-sent)/1000)
		}
	}
}
func main() {
	mode := flag.String("mode", "arena", "public rules mode: arena, classic or teams")
 target := flag.String("target", "http://127.0.0.1:8082", "isolated server")
	metricURL := flag.String("metrics", "http://127.0.0.1:9109/metrics", "private metrics")
	duration := flag.Duration("duration", 20*time.Second, "duration per stage")
	synthetic := flag.Bool("synthetic-clients", false, "loopback-only distinct RFC5737 client buckets")
	flag.Parse()
 if *mode != "arena" && *mode != "classic" && *mode != "teams" { fmt.Fprintln(os.Stderr,"unsupported load mode"); os.Exit(2) }
	u, err := url.Parse(*target)
	if err != nil || u.Scheme != "http" {
		panic("use explicit HTTP isolated target")
	}
	if *synthetic && u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" {
		panic("synthetic headers are permitted only on loopback")
	}
	http.DefaultClient.Timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	clients := make([]*client, 100)
	for i := range clients {
		clients[i], err = join(ctx, *target, *target, i, *synthetic, *mode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	for _, count := range []int{25, 50, 100} {
		before := metrics(*metricURL)
		start := time.Now()
		stage, stop := context.WithTimeout(ctx, *duration)
		var wg sync.WaitGroup
		errs := make(chan error, count)
		for i, c := range clients[:count] {
			c.frames = 0
			c.bytes = 0
			c.disconnects = 0
			c.intervals = nil
			c.rtt = nil
			wg.Add(1)
			go func(i int, c *client) {
				defer wg.Done()
				if e := c.run(stage, *target, *target, i); e != nil {
					errs <- e
				}
			}(i, c)
		}
		wg.Wait()
		stop()
		close(errs)
		for e := range errs {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		elapsed := time.Since(start).Seconds()
		after := metrics(*metricURL)
		frames, bytes, disconnects, empty := 0, 0, 0, 0
		var intervals, rtt []float64
		for _, c := range clients[:count] {
			frames += c.frames
			bytes += c.bytes
			disconnects += c.disconnects
			if c.frames == 0 {
				empty++
			}
			intervals = append(intervals, c.intervals...)
			rtt = append(rtt, c.rtt...)
		}
		result := map[string]any{"mode": *mode, "connections": count, "seconds": elapsed, "snapshots": frames, "snapshot_interval_p95_ms": quantile(intervals, .95), "snapshot_interval_p99_ms": quantile(intervals, .99), "rtt_p95_ms": quantile(rtt, .95), "rtt_p99_ms": quantile(rtt, .99), "bytes_received": bytes, "disconnects": disconnects, "empty_clients": empty, "ticks_per_second": (after["cells_ticks_total"] - before["cells_ticks_total"]) / elapsed, "tick_window_p95_ms": after["cells_tick_p95_ms"], "tick_window_p99_ms": after["cells_tick_p99_ms"], "cpu_core_equivalents": (after["process_cpu_seconds_total"] - before["process_cpu_seconds_total"]) / elapsed, "rss_bytes": after["process_resident_memory_bytes"], "bots_at_end": after["cells_bots"], "rooms_at_end": after["cells_rooms"], "dropped_snapshots": after["cells_dropped_total"] - before["cells_dropped_total"], "synthetic_distinct_rate_buckets": *synthetic}
		encoded, _ := json.Marshal(result)
		fmt.Println(string(encoded))
		if disconnects > 0 || empty > 0 {
			os.Exit(1)
		}
	}
}
