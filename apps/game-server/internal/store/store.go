package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Store struct{ Pool *pgxpool.Pool }
type Session struct{ Hash, ID, Nickname, Mode, Room, Profile string }
type SavedRoom struct {
	Key  string
	Data json.RawMessage
}

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, `BEGIN;
CREATE TABLE IF NOT EXISTS guest_sessions (
 token_hash text PRIMARY KEY, id text UNIQUE NOT NULL, nickname text NOT NULL,
 mode text NOT NULL, room text NOT NULL DEFAULT '', profile text NOT NULL DEFAULT 'balanced',
 created_at timestamptz NOT NULL DEFAULT now(), expires_at timestamptz NOT NULL,
 peak double precision NOT NULL DEFAULT 0, kills integer NOT NULL DEFAULT 0, deaths integer NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS guest_sessions_expiry ON guest_sessions(expires_at);
CREATE TABLE IF NOT EXISTS game_rooms (key text PRIMARY KEY, version integer NOT NULL, state jsonb NOT NULL, updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS cells_schema (version integer PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now());
INSERT INTO cells_schema(version) VALUES(2) ON CONFLICT DO NOTHING;
COMMIT;`)
	return err
}
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func randomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
func (s *Store) Create(ctx context.Context, nick, mode, room, profile string) (Session, string, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, "", err
	}
	id, err := randomToken()
	if err != nil {
		return Session{}, "", err
	}
	session := Session{Hash: Hash(token), ID: id[:24], Nickname: nick, Mode: mode, Room: room, Profile: profile}
	_, err = s.Pool.Exec(ctx, `INSERT INTO guest_sessions(token_hash,id,nickname,mode,room,profile,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, session.Hash, session.ID, nick, mode, room, profile, time.Now().Add(30*24*time.Hour))
	return session, token, err
}
func (s *Store) Session(ctx context.Context, token string) (Session, error) {
	var v Session
	err := s.Pool.QueryRow(ctx, `SELECT token_hash,id,nickname,mode,room,profile FROM guest_sessions WHERE token_hash=$1 AND expires_at>now()`, Hash(token)).Scan(&v.Hash, &v.ID, &v.Nickname, &v.Mode, &v.Room, &v.Profile)
	return v, err
}
func (s *Store) Delete(ctx context.Context, hash string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM guest_sessions WHERE token_hash=$1`, hash)
	return err
}
func (s *Store) SaveRooms(ctx context.Context, rooms []SavedRoom) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	keys := make([]string, 0, len(rooms))
	for _, r := range rooms {
		keys = append(keys, r.Key)
		if _, err = tx.Exec(ctx, `INSERT INTO game_rooms(key,version,state) VALUES($1,1,$2) ON CONFLICT(key) DO UPDATE SET version=1,state=EXCLUDED.state,updated_at=now()`, r.Key, r.Data); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, `DELETE FROM game_rooms WHERE NOT(key=ANY($1::text[]))`, keys); err != nil {
		return err
	}
	// Remove only abandoned Cells rooms and expired anonymous sessions.
	if _, err = tx.Exec(ctx, `DELETE FROM game_rooms WHERE updated_at < now()-interval '1 day'; DELETE FROM guest_sessions WHERE expires_at<now()`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) LoadRooms(ctx context.Context) ([]SavedRoom, error) {
	rows, err := s.Pool.Query(ctx, `SELECT key,state FROM game_rooms WHERE version=1 AND updated_at>now()-interval '1 day' ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []SavedRoom
	for rows.Next() {
		var r SavedRoom
		if err = rows.Scan(&r.Key, &r.Data); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
