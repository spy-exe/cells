package server

import (
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"github.com/spy-exe/cells/apps/game-server/internal/store"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

type limit struct {
	count int
	start time.Time
}
type limiter struct {
	mu    sync.Mutex
	items map[string]limit
}

func (l *limiter) allow(key string, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.items == nil {
		l.items = map[string]limit{}
	}
	now := time.Now()
	if len(l.items) > 4096 {
		for k, v := range l.items {
			if now.Sub(v.start) > time.Hour {
				delete(l.items, k)
			}
		}
		if len(l.items) > 8192 {
			return false
		}
	}
	v := l.items[key]
	if now.Sub(v.start) > window {
		v = limit{start: now}
	}
	v.count++
	l.items[key] = v
	return v.count <= max
}
func ip(r *http.Request) string {
	if value := r.Header.Get("CF-Connecting-IP"); net.ParseIP(value) != nil {
		return value
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
func (s *Server) origin(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") != s.config.Origin {
		writeJSON(w, 403, map[string]string{"error": "Origem não autorizada."})
		return false
	}
	return true
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(&struct{}{}) != io.EOF {
		writeJSON(w, 400, map[string]string{"error": "Dados inválidos."})
		return false
	}
	return true
}
func nickname(v string) bool {
	n := utf8.RuneCountInString(v)
	if n < 2 || n > 20 || !utf8.ValidString(v) {
		return false
	}
	for _, r := range v {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Mn, r) {
			return false
		}
	}
	return true
}
func (s *Server) session(r *http.Request) (store.Session, error) {
	cookie, err := r.Cookie("cells_session")
	if err != nil {
		return store.Session{}, err
	}
	return s.config.Store.Session(r.Context(), cookie.Value)
}
func (s *Server) cookie(w http.ResponseWriter, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: "cells_session", Value: value, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(s.config.Origin, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: age})
}
func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/guest", func(w http.ResponseWriter, r *http.Request) {
		if !s.origin(w, r) {
			return
		}
		if !s.limits.allow("join:"+ip(r), 40, time.Minute) {
			writeJSON(w, 429, map[string]string{"error": "Muitas entradas. Aguarde um minuto."})
			return
		}
		var input struct {
			Nickname string `json:"nickname"`
			Mode     string `json:"mode"`
			Room     string `json:"room"`
			Profile  string `json:"profile"`
			Spectate bool   `json:"spectate"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		input.Nickname = strings.TrimSpace(input.Nickname)
		if !nickname(input.Nickname) || input.Mode != "classic" || len(input.Room) > 32 {
			writeJSON(w, 400, map[string]string{"error": "Use um apelido de 2 a 20 caracteres e um modo disponível."})
			return
		}
		for _, c := range input.Room {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				writeJSON(w, 400, map[string]string{"error": "Código de sala inválido."})
				return
			}
		}
		profile := "balanced"
		if input.Spectate {
			profile = "spectator"
		}
		v, token, err := s.config.Store.Create(r.Context(), input.Nickname, input.Mode, input.Room, profile)
		if err != nil {
			writeJSON(w, 503, map[string]string{"error": "Não foi possível abrir a sessão."})
			return
		}
		s.cookie(w, token, 30*24*3600)
		writeJSON(w, 201, map[string]string{"nickname": v.Nickname, "mode": v.Mode, "room": v.Room})
	})
	mux.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.session(r)
		if err != nil {
			writeJSON(w, 401, map[string]string{"error": "Sessão encerrada."})
			return
		}
		writeJSON(w, 200, map[string]string{"nickname": v.Nickname, "mode": v.Mode, "room": v.Room})
	})
	mux.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		if !s.origin(w, r) {
			return
		}
		if v, err := s.session(r); err == nil {
			_ = s.config.Store.Delete(r.Context(), v.Hash)
		}
		s.cookie(w, "", -1)
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		if !s.origin(w, r) {
			return
		}
		if !s.limits.allow("ws:"+ip(r), 60, time.Minute) {
			http.Error(w, "Too many connections", 429)
			return
		}
		v, err := s.session(r)
		if err != nil {
			http.Error(w, "Session required", 401)
			return
		}
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{strings.TrimPrefix(strings.TrimPrefix(s.config.Origin, "https://"), "http://")}, CompressionMode: websocket.CompressionDisabled})
		if err != nil {
			return
		}
		p, err := s.config.Hub.Join(v, c)
		if err != nil {
			_ = c.Close(websocket.StatusTryAgainLater, "arena full")
			return
		}
		s.config.Hub.Serve(context.Background(), p)
	})
}
