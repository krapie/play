package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krapie/play-api/db"
)

type Handler struct{ db *db.DB }

func New(d *db.DB) *Handler { return &Handler{db: d} }

// POST /api/results
type submitReq struct {
	Game     string `json:"game"`
	PlayerID string `json:"player_id"`
	Value    int    `json:"value"`
	Meta     string `json:"meta"`
}

type submitResp struct {
	ID   int64 `json:"id"`
	Rank int   `json:"rank"`
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Game == "" || req.PlayerID == "" || req.Value <= 0 || req.Value > 60000 {
		http.Error(w, "invalid fields", http.StatusBadRequest)
		return
	}
	req.Game = strings.ToLower(strings.TrimSpace(req.Game))

	conn := h.db.Conn()
	res, err := conn.ExecContext(r.Context(),
		`INSERT INTO results (game, player_id, value, meta) VALUES (?, ?, ?, ?)`,
		req.Game, req.PlayerID, req.Value, nullableString(req.Meta),
	)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()

	// rank = number of players with a better best + 1
	var rank int
	conn.QueryRowContext(r.Context(), `
		SELECT COUNT(*) + 1 FROM (
			SELECT player_id, MIN(value) AS best
			FROM results WHERE game = ?
			GROUP BY player_id
		) WHERE best < (
			SELECT MIN(value) FROM results WHERE game = ? AND player_id = ?
		)
	`, req.Game, req.Game, req.PlayerID).Scan(&rank)

	writeJSON(w, submitResp{ID: id, Rank: rank})
}

// GET /api/results/{game}/leaderboard
type leaderboardEntry struct {
	Rank      int    `json:"rank"`
	PlayerID  string `json:"player_id"`
	Value     int    `json:"value"`
	CreatedAt string `json:"created_at"`
}

type leaderboardResp struct {
	Leaderboard []leaderboardEntry `json:"leaderboard"`
}

func (h *Handler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")

	rows, err := h.db.Conn().QueryContext(r.Context(), `
		SELECT player_id, MIN(value) AS best, MIN(created_at) AS first_at
		FROM results
		WHERE game = ?
		GROUP BY player_id
		ORDER BY best ASC
		LIMIT 10
	`, game)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	entries := []leaderboardEntry{}
	rank := 1
	for rows.Next() {
		var e leaderboardEntry
		if err := rows.Scan(&e.PlayerID, &e.Value, &e.CreatedAt); err != nil {
			continue
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}

	writeJSON(w, leaderboardResp{Leaderboard: entries})
}

// GET /api/results/{game}/player/{pid}
type playerResp struct {
	Best  int `json:"best"`
	Avg   int `json:"avg"`
	Count int `json:"count"`
	Rank  int `json:"rank"`
}

func (h *Handler) PlayerStats(w http.ResponseWriter, r *http.Request) {
	game := chi.URLParam(r, "game")
	pid := chi.URLParam(r, "pid")

	var resp playerResp
	err := h.db.Conn().QueryRowContext(r.Context(), `
		SELECT MIN(value), CAST(AVG(value) AS INTEGER), COUNT(*)
		FROM results WHERE game = ? AND player_id = ?
	`, game, pid).Scan(&resp.Best, &resp.Avg, &resp.Count)
	if err == sql.ErrNoRows || resp.Count == 0 {
		writeJSON(w, playerResp{})
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	h.db.Conn().QueryRowContext(r.Context(), `
		SELECT COUNT(*) + 1 FROM (
			SELECT player_id, MIN(value) AS best
			FROM results WHERE game = ?
			GROUP BY player_id
		) WHERE best < ?
	`, game, resp.Best).Scan(&resp.Rank)

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
