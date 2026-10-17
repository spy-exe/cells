package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spy-exe/cells/apps/game-server/internal/game"
	"github.com/spy-exe/cells/apps/game-server/internal/server"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var version = "0.4.1"
var revision = "dev"

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	pool, err := pgxpool.New(ctx, env("CELLS_DATABASE", "user=cells dbname=cells host=/var/run/postgresql"))
	if err != nil {
		slog.Error("database configuration rejected")
		os.Exit(1)
	}
	defer pool.Close()
	ping, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = pool.Ping(ping)
	cancel()
	if err != nil {
		slog.Error("database unavailable")
		os.Exit(1)
	}
	st := store.New(pool)
	err = st.Migrate(ctx)
	if err != nil {
		slog.Error("schema initialization failed")
		os.Exit(1)
	}
	var hub *game.Hub
	if os.Getenv("CELLS_CANDIDATE") != "1" {
		size, _ := strconv.ParseFloat(os.Getenv("CELLS_MAP_SIZE"), 64)
		hub = game.New(st, size)
		if err = hub.Load(ctx); err != nil {
			slog.Error("checkpoint rejected")
			os.Exit(1)
		}
		go hub.Run(ctx)
		metrics := &http.Server{Addr: env("CELLS_METRICS_ADDR", "127.0.0.1:9108"), Handler: http.HandlerFunc(hub.Metrics), ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second}
		go func() {
			if e := metrics.ListenAndServe(); e != nil && e != http.ErrServerClosed {
				slog.Error("metrics listener failed")
			}
		}()
		defer metrics.Close()
	}
	app := server.New(server.Config{ExpeditionEnabled: os.Getenv("CELLS_EXPEDITION") == "1", Version: version, Revision: revision, WebDir: env("CELLS_WEB_DIR", "dist/web"), Origin: env("CELLS_ORIGIN", "http://127.0.0.1:8080"), Ready: pool.Ping, Store: st, Hub: hub})
	httpServer := &http.Server{Addr: env("CELLS_ADDR", "127.0.0.1:8080"), Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	go func() {
		slog.Info("cells started", "version", version, "revision", revision)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly")
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if hub != nil {
		hub.Close(shutdown)
	}
	_ = httpServer.Shutdown(shutdown)
}
