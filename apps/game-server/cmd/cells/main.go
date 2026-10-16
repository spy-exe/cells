package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spy-exe/cells/apps/game-server/internal/server"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "0.0.1"
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
	_, err = pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS cells_schema (version integer PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now()); INSERT INTO cells_schema(version) VALUES(1) ON CONFLICT DO NOTHING")
	if err != nil {
		slog.Error("schema initialization failed")
		os.Exit(1)
	}
	app := server.New(server.Config{Version: version, Revision: revision, WebDir: env("CELLS_WEB_DIR", "dist/web"), Origin: env("CELLS_ORIGIN", "http://127.0.0.1:8080"), Ready: pool.Ping})
	httpServer := &http.Server{Addr: env("CELLS_ADDR", "127.0.0.1:8080"), Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	go func() {
		slog.Info("cells started", "version", version, "revision", revision)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly")
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdown)
}
