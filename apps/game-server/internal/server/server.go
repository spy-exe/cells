package server

import (
	"context"
	"encoding/json"
	"github.com/spy-exe/cells/apps/game-server/internal/game"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Version, Revision, WebDir, Origin string
	Ready                             func(context.Context) error
	Store                             *store.Store
	Hub                               *game.Hub
}
type Server struct {
	config Config
	limits limiter
}

func New(config Config) *Server { return &Server{config: config} }
func (s *Server) ready() bool {
	if s.config.Ready == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.config.Ready(ctx) == nil
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "version": s.config.Version, "revision": s.config.Revision})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ready := s.ready()
		code := 200
		if !ready {
			code = 503
		}
		writeJSON(w, code, map[string]any{"ok": ready, "version": s.config.Version, "revision": s.config.Revision})
	})
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		modes := []string{}
		players, bots := any(0), any(0)
		if s.config.Hub != nil {
			modes = []string{"classic", "arena"}
			stats := s.config.Hub.Stats()
			players = stats["players"]
			bots = stats["bots"]
		}
		writeJSON(w, 200, map[string]any{"ok": true, "ready": s.ready(), "version": s.config.Version, "revision": s.config.Revision, "modes": modes, "players": players, "bots": bots})
	})
	if s.config.Store != nil && s.config.Hub != nil {
		s.routes(mux)
	}
	if s.config.WebDir != "" {
		files := http.FileServer(http.Dir(s.config.WebDir))
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			path := filepath.Join(s.config.WebDir, filepath.Clean("/"+r.URL.Path))
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				if strings.HasPrefix(r.URL.Path, "/assets/") {
					w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, filepath.Join(s.config.WebDir, "index.html"))
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src 'self'; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if strings.HasPrefix(s.config.Origin, "https://") && r.Header.Get("X-Forwarded-Proto") == "http" {
			http.Redirect(w, r, s.config.Origin+r.URL.RequestURI(), http.StatusPermanentRedirect)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
