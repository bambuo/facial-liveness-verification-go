package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

// handleScheduleCreate 实现 POST /api/v2/monitoring/schedules。
func (s *Server) handleScheduleCreate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		UserID                string `json:"user_id"`
		IntervalDays          int32  `json:"interval_days"`
		VerificationRequestID string `json:"verification_request_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "user_id 必须是 UUID"))
		return
	}
	if req.IntervalDays < 30 || req.IntervalDays > 730 {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "interval_days 必须在 30-730 之间"))
		return
	}
	var verificationID *uuid.UUID
	if req.VerificationRequestID != "" {
		id, err := uuid.Parse(req.VerificationRequestID)
		if err != nil {
			writeError(w, r, domain.New(domain.CodeDocumentInvalid, "verification_request_id 必须是 UUID"))
			return
		}
		verificationID = &id
	}
	now := time.Now().UTC()
	schedule := &gormmodel.ReverificationSchedule{
		ID: uuid.New(), DeveloperID: principal.DeveloperID, UserID: userID,
		IntervalDays: req.IntervalDays, NextVerificationAt: now.AddDate(0, 0, int(req.IntervalDays)),
		Status: "active", CreatedAt: &now, VerificationRequestID: verificationID,
	}
	if err := s.idswyft.CreateSchedule(r.Context(), schedule); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "schedule": scheduleView(schedule)})
}

// handleScheduleList 实现 GET /api/v2/monitoring/schedules。
func (s *Server) handleScheduleList(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	page, limit := pageLimit(r)
	status := r.URL.Query().Get("status")
	schedules, total, err := s.idswyft.ListSchedules(r.Context(), principal.DeveloperID, status, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(schedules))
	for i := range schedules {
		out = append(out, scheduleView(&schedules[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedules": out, "total": total, "page": page, "limit": limit})
}

// handleScheduleGet 实现 GET /api/v2/monitoring/schedules/:schedule_id。
func (s *Server) handleScheduleGet(w http.ResponseWriter, r *http.Request) {
	schedule, ok := s.schedule(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedule": scheduleView(schedule)})
}

// handleScheduleCancel 实现 DELETE /api/v2/monitoring/schedules/:schedule_id。
func (s *Server) handleScheduleCancel(w http.ResponseWriter, r *http.Request) {
	schedule, ok := s.schedule(w, r)
	if !ok {
		return
	}
	affected, err := s.idswyft.CancelSchedule(r.Context(), principalFrom(r).DeveloperID, schedule.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "Schedule not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Schedule cancelled"})
}

// handleExpiringDocuments 实现 GET /api/v2/monitoring/expiring-documents。
func (s *Server) handleExpiringDocuments(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	page, limit := pageLimit(r)
	days, _ := strconv.Atoi(r.URL.Query().Get("days_ahead"))
	if days < 1 || days > 365 {
		days = 90
	}
	alerts, total, err := s.idswyft.ExpiringDocuments(r.Context(), principal.DeveloperID, days, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"alerts": alerts, "total": total, "page": page, "limit": limit,
	})
}

func (s *Server) schedule(w http.ResponseWriter, r *http.Request) (*gormmodel.ReverificationSchedule, bool) {
	id, err := uuid.Parse(r.PathValue("schedule_id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "Schedule not found"})
		return nil, false
	}
	schedule, err := s.idswyft.GetSchedule(r.Context(), principalFrom(r).DeveloperID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "Schedule not found"})
		return nil, false
	}
	return schedule, true
}

func scheduleView(schedule *gormmodel.ReverificationSchedule) map[string]any {
	return map[string]any{
		"id": schedule.ID, "user_id": schedule.UserID, "interval_days": schedule.IntervalDays,
		"next_verification_at": schedule.NextVerificationAt, "last_verification_at": schedule.LastVerificationAt,
		"status": schedule.Status, "created_at": schedule.CreatedAt,
	}
}

func pageLimit(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
