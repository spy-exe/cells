package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHealthAndReadiness(t *testing.T) {
	s := New(Config{Version: "0.0.1", Revision: "test"})
	for _, path := range []string{"/healthz", "/api/status"} {
		r := httptest.NewRecorder()
		s.Handler().ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 {
			t.Fatalf("%s: %d", path, r.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["version"] != "0.0.1" {
			t.Fatal(body)
		}
	}
	r := httptest.NewRecorder()
	s.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/readyz", nil))
	if r.Code != 503 {
		t.Fatal("readiness must fail without the database")
	}
}

func TestHTTPSRedirectHasFixedOrigin(t *testing.T) {
	s := New(Config{Origin: "https://cells.malha.app"})
	r := httptest.NewRequest("GET", "http://untrusted.example/path?x=1", nil)
	r.Header.Set("X-Forwarded-Proto", "http")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 308 || w.Header().Get("Location") != "https://cells.malha.app/path?x=1" {
		t.Fatal(w.Code, w.Header())
	}
}
