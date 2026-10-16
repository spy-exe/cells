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
  if r.Code != 200 { t.Fatalf("%s: %d", path, r.Code) }
  var body map[string]any
  if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil { t.Fatal(err) }
  if body["version"] != "0.0.1" { t.Fatal(body) }
 }
 r := httptest.NewRecorder()
 s.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/readyz", nil))
 if r.Code != 503 { t.Fatal("readiness must fail without the database") }
}
