package game

import (
	"fmt"
	"net/http"
	"runtime"
)

// Metrics is exposed only on the private metrics listener, never the public mux.
func (h *Hub) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	stats := h.Stats()
	for _, name := range []string{"players", "bots", "rooms", "entities", "tick_p95_ms", "tick_p99_ms"} {
		fmt.Fprintf(w, "# TYPE cells_%s gauge\ncells_%s %v\n", name, name, stats[name])
	}
	for _, name := range []string{"ticks", "bytes", "dropped", "rejected"} {
		fmt.Fprintf(w, "# TYPE cells_%s_total counter\ncells_%s_total %v\n", name, name, stats[name])
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(w, "# TYPE cells_heap_bytes gauge\ncells_heap_bytes %d\n# TYPE cells_goroutines gauge\ncells_goroutines %d\n", m.Alloc, runtime.NumGoroutine())
}
