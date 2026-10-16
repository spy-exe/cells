package server

import (
	"bytes"
	"context"
	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spy-exe/cells/apps/game-server/internal/game"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGuestOriginValidationAndMultiplayer(t *testing.T) {
	db := os.Getenv("CELLS_TEST_DATABASE")
	if db == "" {
		t.Skip("CELLS_TEST_DATABASE required for database integration")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, err := pgxpool.New(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	st := store.New(pool)
	if err = st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	hub := game.New(st, 1600)
	go hub.Run(ctx)
	defer hub.Close(context.Background())
	app := New(Config{Store: st, Hub: hub, Origin: "http://cells.test", Ready: pool.Ping})
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()
	join := func(origin, body string) *http.Response {
		req, _ := http.NewRequest("POST", ts.URL+"/api/guest", bytes.NewBufferString(body))
		req.Header.Set("Origin", origin)
		res, e := http.DefaultClient.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		return res
	}
	bad := join("http://evil.test", `{"nickname":"A"}`)
	bad.Body.Close()
	if bad.StatusCode != 403 {
		t.Fatal("CSRF origin accepted", bad.StatusCode)
	}
	bad = join("http://cells.test", `{"nickname":"A","mass":999999}`)
	bad.Body.Close()
	if bad.StatusCode != 400 {
		t.Fatal("extra input accepted")
	}
	a := join("http://cells.test", `{"nickname":"Alice","mode":"classic","room":"test-abcdefgh"}`)
	defer a.Body.Close()
	if a.StatusCode != 201 {
		t.Fatal("join", a.StatusCode)
	}
	b := join("http://cells.test", `{"nickname":"Bruno","mode":"classic","room":"test-abcdefgh"}`)
	defer b.Body.Close()
	if b.StatusCode != 201 {
		t.Fatal("join2", b.StatusCode)
	}
	if len(a.Cookies()) != 1 || !a.Cookies()[0].HttpOnly || a.Cookies()[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("insecure session")
	}
	conn := func(res *http.Response) *websocket.Conn {
		headers := http.Header{}
		headers.Set("Origin", "http://cells.test")
		headers.Set("Cookie", res.Cookies()[0].String())
		c, _, err := websocket.Dial(ctx, "ws"+ts.URL[4:]+"/ws", &websocket.DialOptions{HTTPHeader: headers})
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	ca, cb := conn(a), conn(b)
	defer ca.CloseNow()
	defer cb.CloseNow()
	_, ba, err := ca.Read(ctx)
	if err != nil || len(ba) < 53 || ba[2] != 2 {
		t.Fatal("snapshot", err)
	}
	_, bb, err := cb.Read(ctx)
	if err != nil || len(bb) < 53 {
		t.Fatal(err)
	}
	if string(ba[7:11]) == string(bb[7:11]) {
		t.Fatal("shared player identity")
	}
	if err := ca.Write(ctx, websocket.MessageBinary, []byte{206, 2, 1}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_, _, err = ca.Read(ctx)
		if err != nil {
			break
		}
	}
	if err == nil {
		t.Fatal("invalid protocol not rejected")
	}
}
