package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestPersistentSessionsAndRoomReplacement(t *testing.T) {
	db := os.Getenv("CELLS_TEST_DATABASE")
	if db == "" {
		t.Skip("CELLS_TEST_DATABASE required")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(db)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgxpool.NewWithConfig(ctx, cfg.Copy())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("cells_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	st := New(pool)
	if err = st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err = st.Migrate(ctx); err != nil {
		t.Fatal("migration not idempotent", err)
	}
	v, token, err := st.Create(ctx, "Persistente", "classic", "friends", "balanced")
	if err != nil {
		t.Fatal(err)
	}
	if v.Hash == token || len(token) != 64 {
		t.Fatal("plaintext token/hash")
	}
	found, err := st.Session(ctx, token)
	if err != nil || found.ID != v.ID {
		t.Fatal("session not persistent", err)
	}
	if _, err = st.Session(ctx, "invalid"); err == nil {
		t.Fatal("invalid token accepted")
	}
	if err = st.SaveRooms(ctx, []SavedRoom{{Key: "a", Data: json.RawMessage(`{"tick":1}`)}}); err != nil {
		t.Fatal(err)
	}
	if rows, err := st.LoadRooms(ctx); err != nil || len(rows) != 1 {
		t.Fatal("room not persisted", err)
	}
	if err = st.SaveRooms(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if rows, err := st.LoadRooms(ctx); err != nil || len(rows) != 0 {
		t.Fatal("abandoned room resurrected", err)
	}
	if err = st.Delete(ctx, v.Hash); err != nil {
		t.Fatal(err)
	}
	if _, err = st.Session(ctx, token); err == nil {
		t.Fatal("revoked session accepted")
	}
}

func TestDatabaseOutageReturnsErrors(t *testing.T) {
	db := os.Getenv("CELLS_TEST_DATABASE")
	if db == "" {
		t.Skip("CELLS_TEST_DATABASE required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	pool.Close()
	s := New(pool)
	if s.Migrate(ctx) == nil {
		t.Fatal("migration hid database outage")
	}
	if _, _, err = s.Create(ctx, "Guest", "classic", "", "balanced"); err == nil {
		t.Fatal("session created without persistence")
	}
	if s.SaveRooms(ctx, nil) == nil {
		t.Fatal("checkpoint hid database outage")
	}
	if _, err = s.LoadRooms(ctx); err == nil {
		t.Fatal("load hid database outage")
	}
}
