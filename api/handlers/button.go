package handlers

import (
	"encoding/json"
	"net/http"
)

var buttonMilestones = []int{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000}

func nextButtonMilestone(total int) int {
	for _, m := range buttonMilestones {
		if m > total {
			return m
		}
	}
	return 0
}

type buttonPressReq struct {
	SessionID string `json:"session_id"`
}

type buttonPressResp struct {
	Total         int `json:"total"`
	Yours         int `json:"yours"`
	Rate          int `json:"rate"`
	JustHit       int `json:"just_hit,omitempty"`
	NextMilestone int `json:"next_milestone"`
}

type buttonStatsResp struct {
	Total         int `json:"total"`
	Rate          int `json:"rate"`
	NextMilestone int `json:"next_milestone"`
}

func (h *Handler) ButtonPress(w http.ResponseWriter, r *http.Request) {
	var req buttonPressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	conn := h.db.Conn()
	ctx := r.Context()

	var recentCount int
	conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM button_presses WHERE session_id = ? AND pressed_at > datetime('now', '-1 second')`,
		req.SessionID,
	).Scan(&recentCount)
	if recentCount >= 20 {
		http.Error(w, "slow down", http.StatusTooManyRequests)
		return
	}

	conn.ExecContext(ctx, `INSERT INTO button_presses (session_id) VALUES (?)`, req.SessionID)

	var total, yours, rate int
	conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM button_presses`).Scan(&total)
	conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM button_presses WHERE session_id = ?`, req.SessionID).Scan(&yours)
	conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM button_presses WHERE pressed_at > datetime('now', '-1 minute')`).Scan(&rate)

	resp := buttonPressResp{
		Total:         total,
		Yours:         yours,
		Rate:          rate,
		NextMilestone: nextButtonMilestone(total),
	}
	for _, m := range buttonMilestones {
		if total == m {
			resp.JustHit = m
			break
		}
	}

	writeJSON(w, resp)
}

func (h *Handler) ButtonStats(w http.ResponseWriter, r *http.Request) {
	conn := h.db.Conn()
	ctx := r.Context()

	var total, rate int
	conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM button_presses`).Scan(&total)
	conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM button_presses WHERE pressed_at > datetime('now', '-1 minute')`).Scan(&rate)

	writeJSON(w, buttonStatsResp{
		Total:         total,
		Rate:          rate,
		NextMilestone: nextButtonMilestone(total),
	})
}
