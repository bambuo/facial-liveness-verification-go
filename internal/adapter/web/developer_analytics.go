package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// handleDeveloperStats 实现 GET /api/developer/stats。
func (s *Server) handleDeveloperStats(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	counts, err := s.idswyft.VerificationStatusCounts(r.Context(), principal.DeveloperID, monthStart)
	if err != nil {
		writeError(w, r, err)
		return
	}
	total := int64(0)
	for _, count := range counts {
		total += count
	}
	success := counts["verified"]
	successRate := "0%"
	if total > 0 {
		successRate = formatPercent(float64(success) / float64(total) * 100)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"period": "month", "period_start": monthStart, "total_requests": total,
		"successful_requests": success, "failed_requests": counts["failed"],
		"pending_requests": counts["pending"], "manual_review_requests": counts["manual_review"],
		"success_rate": successRate, "monthly_limit": 50, "monthly_usage": total,
		"remaining_quota":  max(0, 50-total),
		"quota_reset_date": time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC),
	})
}

// handleDeveloperActivity 实现 GET /api/developer/activity。
func (s *Server) handleDeveloperActivity(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	logs, err := s.idswyft.RecentActivity(r.Context(), principal.DeveloperID, 100)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": logs, "total": len(logs)})
}

// handleDeveloperVerification 实现 GET /api/developer/verifications/:verificationId。
func (s *Server) handleDeveloperVerification(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, err := uuid.Parse(r.PathValue("verificationId"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return
	}
	verification, err := s.loadOwnedVerification(r.Context(), id, principal)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "verification": verificationPublic(verification)})
}

// handleDeveloperAnalytics 实现 GET /api/developer/analytics。
func (s *Server) handleDeveloperAnalytics(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	since := time.Now().UTC().AddDate(0, 0, -30)
	counts, err := s.idswyft.VerificationStatusCounts(r.Context(), principal.DeveloperID, since)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"period": "30d", "since": since, "counts": counts,
	})
}

func formatPercent(value float64) string {
	return strconvFormat(value) + "%"
}

func strconvFormat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}
