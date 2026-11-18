package web

import (
	"net/http"
	"time"

	"facial-liveness-verification-go/internal/config"
)

func defaultThresholds(cfg *config.Config) map[string]map[string]float64 {
	return map[string]map[string]float64{
		"production": {
			"photo_consistency": 0.85, "face_matching": cfg.Face.MatchThreshold,
			"liveness": cfg.Face.LivenessThreshold, "cross_validation": 0.7,
			"quality_minimum": 0.5, "ocr_confidence": 0.5, "pdf417_confidence": 0.8,
		},
		"sandbox": {
			"photo_consistency": 0.80, "face_matching": 0.80,
			"liveness": 0.65, "cross_validation": 0.7,
			"quality_minimum": 0.4, "ocr_confidence": 0.4, "pdf417_confidence": 0.7,
		},
	}
}

// handleThresholdsGet 实现 GET /api/admin/thresholds。
func (s *Server) handleThresholdsGet(w http.ResponseWriter, r *http.Request) {
	s.thresholdMu.RLock()
	defer s.thresholdMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"production": s.thresholds["production"], "sandbox": s.thresholds["sandbox"],
		"meta": map[string]any{"using_defaults": false},
	}})
}

// handleThresholdsUpdate 实现 PUT /api/admin/thresholds。
func (s *Server) handleThresholdsUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AutoApprove  *float64 `json:"auto_approve_threshold"`
		ManualReview *float64 `json:"manual_review_threshold"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.AutoApprove == nil || req.ManualReview == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": map[string]string{"code": "MISSING_THRESHOLDS", "message": "Auto-approve and manual review thresholds are required"}})
		return
	}
	if *req.AutoApprove < 0 || *req.AutoApprove > 1 || *req.ManualReview < 0 || *req.ManualReview > 1 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": map[string]string{"code": "INVALID_THRESHOLDS", "message": "Thresholds must be between 0 and 1"}})
		return
	}
	s.thresholdMu.Lock()
	s.thresholds["production"]["face_matching"] = *req.AutoApprove
	s.thresholds["production"]["liveness"] = *req.ManualReview
	production := cloneThresholds(s.thresholds["production"])
	sandbox := cloneThresholds(s.thresholds["sandbox"])
	s.thresholdMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"updated_settings": map[string]any{"auto_approve_threshold": *req.AutoApprove, "manual_review_threshold": *req.ManualReview},
		"production":       production, "sandbox": sandbox,
		"meta": map[string]any{"updated_at": timeNow()},
	}})
}

// handleThresholdsPreview 实现 POST /api/admin/thresholds/preview。
func (s *Server) handleThresholdsPreview(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"preview": req, "explanation": "Preview only; no settings were changed.",
	}})
}

// handleThresholdsReset 实现 POST /api/admin/thresholds/reset。
func (s *Server) handleThresholdsReset(w http.ResponseWriter, r *http.Request) {
	s.thresholdMu.Lock()
	s.thresholds = defaultThresholds(s.cfg)
	defaults := cloneThresholds(s.thresholds["production"])
	s.thresholdMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"message": "Thresholds reset to defaults", "default_settings": defaults,
	}})
}

func cloneThresholds(source map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}

func timeNow() string { return time.Now().UTC().Format(time.RFC3339) }
