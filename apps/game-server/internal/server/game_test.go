package server

import (
	"bytes"
	"context"
	"encoding/json"
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
	app := New(Config{ExpeditionEnabled: true, Store: st, Hub: hub, Origin: "http://cells.test", Ready: pool.Ping})
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
	unauthorized, err := http.Get(ts.URL + "/api/run")
	if err != nil {
		t.Fatal(err)
	}
	unauthorized.Body.Close()
	if unauthorized.StatusCode != 401 {
		t.Fatal("run leaked without a session", unauthorized.StatusCode)
	}
	req, _ := http.NewRequest("GET", ts.URL+"/api/run?player=someone-else", nil)
	req.AddCookie(a.Cookies()[0])
	runResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var run game.RunSummary
	err = json.NewDecoder(runResponse.Body).Decode(&run)
	runResponse.Body.Close()
	if err != nil || runResponse.StatusCode != 200 || !run.HasRun || !run.Alive || run.Peak < 36 || runResponse.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("own authoritative run summary", run, err)
	}

	// The solo mode is feature gated, session private, and rejects shared rooms.
	for _, body := range []string{`{"nickname":"Solo","mode":"rogue","room":"friends"}`, `{"nickname":"Solo","mode":"rogue","spectate":true}`} {
		bad := join("http://cells.test", body)
		bad.Body.Close()
		if bad.StatusCode != 400 {
			t.Fatal("invalid expedition entry", bad.StatusCode)
		}
	}
	disabled := httptest.NewServer(New(Config{Store: st, Hub: hub, Origin: "http://cells.test"}).Handler())
	defer disabled.Close()
	disabledReq, _ := http.NewRequest("POST", disabled.URL+"/api/guest", bytes.NewBufferString(`{"nickname":"Solo","mode":"rogue"}`))
	disabledReq.Header.Set("Origin", "http://cells.test")
	disabledResponse, err := http.DefaultClient.Do(disabledReq)
	if err != nil {
		t.Fatal(err)
	}
	disabledResponse.Body.Close()
	if disabledResponse.StatusCode != 400 {
		t.Fatal("disabled mode admitted guest")
	}
	solo := join("http://cells.test", `{"nickname":"Solo","mode":"rogue"}`)
	solo.Body.Close()
	if solo.StatusCode != 201 {
		t.Fatal("solo session", solo.StatusCode)
	}
	cs := conn(solo)
	defer cs.CloseNow()
	if _, _, err = cs.Read(ctx); err != nil {
		t.Fatal(err)
	}
	for _, scene := range []struct {
		method, path, origin, body string
		cookie                     bool
		code                       int
	}{
		{"GET", "/api/expedition", "", "", false, 401},
		{"GET", "/api/expedition?player=foreign", "", "", true, 200},
		{"POST", "/api/expedition/mutation", "http://evil.test", `{"mutation":"cilia","revision":1}`, true, 403},
		{"POST", "/api/expedition/mutation", "http://cells.test", `{"mutation":"cilia","revision":1}`, false, 401},
		{"POST", "/api/expedition/mutation", "http://cells.test", `{"mutation":"cilia","revision":1,"mass":500}`, true, 400},
		{"POST", "/api/expedition/mutation", "http://cells.test", `{"mutation":"cilia","revision":1}`, true, 409},
	} {
		request, _ := http.NewRequest(scene.method, ts.URL+scene.path, bytes.NewBufferString(scene.body))
		if scene.cookie {
			request.AddCookie(solo.Cookies()[0])
		}
		request.Header.Set("Origin", scene.origin)
		response, e := http.DefaultClient.Do(request)
		if e != nil {
			t.Fatal(e)
		}
		response.Body.Close()
		if response.StatusCode != scene.code {
			t.Fatalf("%s %s: %d want %d", scene.method, scene.path, response.StatusCode, scene.code)
		}
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
