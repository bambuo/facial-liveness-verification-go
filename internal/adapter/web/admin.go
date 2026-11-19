package web

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// withAdminAuth 保护 admin 路由：service token 或 admin Bearer token。
func (s *Server) withAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.authenticateServiceToken(r) || s.validateAdminToken(r.Header.Get("Authorization")) {
			next(w, r)
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "admin authentication required"})
	}
}

// handleAdminLogin 实现 POST /api/auth/admin/login。
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if s.cfg.Security.AdminEmail == "" || s.cfg.Security.AdminPassword == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "admin login not configured"})
		return
	}
	if !strings.EqualFold(req.Email, s.cfg.Security.AdminEmail) || subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.cfg.Security.AdminPassword)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Invalid credentials"})
		return
	}
	token, err := s.signAdminToken(req.Email)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "token": token, "admin": map[string]any{"email": req.Email, "role": "admin"}})
}

func (s *Server) adminSecret() string {
	if s.cfg.Security.AdminSigningKey != "" {
		return s.cfg.Security.AdminSigningKey
	}
	return s.apiKeySecret()
}

func (s *Server) signAdminToken(email string) (string, error) {
	payload, _ := json.Marshal(map[string]any{"email": email, "exp": time.Now().UTC().Add(12 * time.Hour).Unix()})
	body := base64.RawURLEncoding.EncodeToString(payload)
	signature := hashToken(s.adminSecret(), body)
	return body + "." + signature, nil
}

func (s *Server) validateAdminToken(authorization string) bool {
	if !strings.HasPrefix(authorization, "Bearer ") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(authorization, "Bearer "), ".")
	if len(parts) != 2 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(hashToken(s.adminSecret(), parts[0])), []byte(parts[1])) != 1 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var claims struct {
		Email string `json:"email"`
		Exp   int64  `json:"exp"`
	}
	if json.Unmarshal(raw, &claims) != nil || claims.Exp < time.Now().Unix() {
		return false
	}
	return true
}

// handleAdminDashboard 实现 GET /api/admin/dashboard。
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	counts, _ := s.idswyft.AdminStatusCounts(r.Context(), now.AddDate(0, 0, -30))
	developers, _ := s.idswyft.AdminDeveloperCount(r.Context())
	recent, _, _ := s.idswyft.AdminVerifications(r.Context(), "", nil, 1, 10)
	writeJSON(w, http.StatusOK, map[string]any{
		"stats": counts, "recent_verifications": adminVerificationList(recent),
		"developer_count": developers,
		"system_health":   map[string]any{"status": "ok", "timestamp": now},
	})
}

// handleAdminVerifications 实现 GET /api/admin/verifications。
func (s *Server) handleAdminVerifications(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	var developerID *uuid.UUID
	if raw := r.URL.Query().Get("developer_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "developer_id must be a UUID"})
			return
		}
		developerID = &id
	}
	rows, total, err := s.idswyft.AdminVerifications(r.Context(), r.URL.Query().Get("status"), developerID, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"verifications": adminVerificationList(rows),
		"pagination":    map[string]any{"page": page, "limit": limit, "total": total, "pages": (total + int64(limit) - 1) / int64(limit)},
	})
}

// handleAdminVerification 实现 GET /api/admin/verification/:id。
func (s *Server) handleAdminVerification(w http.ResponseWriter, r *http.Request) {
	verification, ok := s.adminVerification(w, r)
	if !ok {
		return
	}
	payload := verificationPublic(verification)
	documents := make([]map[string]any, 0)
	if verification.DocumentID != nil {
		if document, err := s.idswyft.GetDocument(r.Context(), *verification.DocumentID); err == nil {
			documents = append(documents, map[string]any{
				"id": document.ID, "document_type": document.DocumentType, "file_name": document.FileName,
				"is_back_of_id": document.IsBackOfID, "ocr_data": document.OCRData,
			})
		}
	}
	if back, err := s.idswyft.GetBackDocument(r.Context(), verification.ID); err == nil {
		documents = append(documents, map[string]any{
			"id": back.ID, "document_type": back.DocumentType, "file_name": back.FileName,
			"is_back_of_id": back.IsBackOfID, "ocr_data": back.OCRData,
		})
	}
	payload["documents"] = documents
	if verification.SelfieID != nil && s.storage != nil && s.cfg.Storage.Enabled {
		if selfie, err := s.idswyft.GetSelfie(r.Context(), *verification.SelfieID); err == nil {
			if url, err := s.storage.PresignGet(r.Context(), selfie.FilePath, 5*time.Minute); err == nil {
				payload["selfie_url"] = url
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"verification": payload})
}

// handleAdminDuplicates 实现 GET /api/admin/verification/:id/duplicates。
func (s *Server) handleAdminDuplicates(w http.ResponseWriter, r *http.Request) {
	verification, ok := s.adminVerification(w, r)
	if !ok {
		return
	}
	rows, _ := s.idswyft.DuplicateVerifications(r.Context(), verification)
	writeJSON(w, http.StatusOK, map[string]any{
		"verification_id": verification.ID.String(), "linked_verifications": adminVerificationList(rows),
	})
}

// handleAdminReview 实现 PUT /api/admin/verification/:id/review。
func (s *Server) handleAdminReview(w http.ResponseWriter, r *http.Request) {
	verification, ok := s.adminVerification(w, r)
	if !ok {
		return
	}
	var req struct {
		Decision  string `json:"decision"`
		Reason    string `json:"reason"`
		NewStatus string `json:"new_status"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	status := req.NewStatus
	if status == "" {
		switch req.Decision {
		case "approve", "approved", "verified":
			status = "verified"
		case "reject", "rejected", "failed":
			status = "failed"
		default:
			status = "manual_review"
		}
	}
	if err := s.idswyft.ReviewVerification(r.Context(), verification.ID, status, req.Reason, "admin"); err != nil {
		writeError(w, r, err)
		return
	}
	verification.Status = status
	writeJSON(w, http.StatusOK, map[string]any{"verification": verificationPublic(verification), "message": "Review recorded"})
}

// handleAdminDevelopers 实现 GET /api/admin/developers。
func (s *Server) handleAdminDevelopers(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	developers, total, err := s.idswyft.ListDevelopers(r.Context(), "", "", page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(developers))
	for _, developer := range developers {
		out = append(out, map[string]any{
			"id": developer.ID, "email": developer.Email, "name": developer.Name,
			"company": developer.Company, "status": developer.Status, "created_at": developer.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"developers": out,
		"pagination": map[string]any{"page": page, "limit": limit, "total": total, "pages": (total + int64(limit) - 1) / int64(limit)},
	})
}

// handleAdminAnalytics 实现 GET /api/admin/analytics。
func (s *Server) handleAdminAnalytics(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	counts, _ := s.idswyft.AdminStatusCounts(r.Context(), now.AddDate(0, 0, -30))
	writeJSON(w, http.StatusOK, map[string]any{"period": "30d", "status_counts": counts, "generated_at": now})
}

// handleAdminHealth 实现 GET /api/admin/health。
func (s *Server) handleAdminHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus, redisStatus := "ok", "ok"
	if s.dbPing != nil && s.dbPing(r.Context()) != nil {
		dbStatus = "error"
	}
	if s.redisPing != nil && s.redisPing(r.Context()) != nil {
		redisStatus = "error"
	}
	writeJSON(w, http.StatusOK, map[string]any{"database": dbStatus, "redis": redisStatus, "timestamp": time.Now().UTC()})
}

// handleAdminFunnel 实现 GET /api/admin/analytics/funnel。
func (s *Server) handleAdminFunnel(w http.ResponseWriter, r *http.Request) {
	counts, _ := s.idswyft.AdminStatusCounts(r.Context(), time.Now().UTC().AddDate(0, 0, -30))
	writeJSON(w, http.StatusOK, map[string]any{"funnel": []map[string]any{
		{"stage": "initialized", "count": counts["pending"] + counts["processing"] + counts["verified"] + counts["failed"] + counts["manual_review"]},
		{"stage": "verified", "count": counts["verified"]},
		{"stage": "manual_review", "count": counts["manual_review"]},
	}})
}

// handleAdminRejections 实现 GET /api/admin/analytics/rejections。
func (s *Server) handleAdminRejections(w http.ResponseWriter, r *http.Request) {
	rows, _, _ := s.idswyft.AdminVerifications(r.Context(), "failed", nil, 1, 100)
	writeJSON(w, http.StatusOK, map[string]any{"rejections": adminVerificationList(rows)})
}

// handleAdminFraudPatterns 实现 GET /api/admin/analytics/fraud-patterns。
func (s *Server) handleAdminFraudPatterns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"patterns": []any{}})
}

// handleAdminRiskDistribution 实现 GET /api/admin/analytics/risk-distribution。
func (s *Server) handleAdminRiskDistribution(w http.ResponseWriter, r *http.Request) {
	counts, _ := s.idswyft.AdminStatusCounts(r.Context(), time.Now().UTC().AddDate(0, 0, -30))
	writeJSON(w, http.StatusOK, map[string]any{"distribution": counts})
}

// handleAdminProviderMetrics 实现 GET /api/admin/provider-metrics。
func (s *Server) handleAdminProviderMetrics(w http.ResponseWriter, r *http.Request) {
	summary, err := s.idswyft.ProviderMetricSummary(r.Context(), time.Now().UTC().AddDate(0, 0, -7))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": summary})
}

// handleAdminAuditExport 实现 GET /api/admin/audit/export。
func (s *Server) handleAdminAuditExport(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	from, to := now.AddDate(0, 0, -7), now
	if value := r.URL.Query().Get("from"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			from = parsed
		}
	}
	if value := r.URL.Query().Get("to"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			to = parsed
		}
	}
	var developerID *uuid.UUID
	if value := r.URL.Query().Get("developer_id"); value != "" {
		if id, err := uuid.Parse(value); err == nil {
			developerID = &id
		}
	}
	var status *int32
	if value := r.URL.Query().Get("status"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			v := int32(parsed)
			status = &v
		}
	}
	records, err := s.idswyft.AuditLogs(r.Context(), from, to, developerID, status)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"export_date": now, "record_count": len(records),
		"filters": map[string]any{"from": from, "to": to, "status": status, "developer_id": developerID},
		"records": records,
	})
}

// handleAdminDeleteUserData 实现 DELETE /api/admin/user/:userId/data。
func (s *Server) handleAdminDeleteUserData(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "userId must be a UUID"})
		return
	}
	affected, err := s.idswyft.DeleteUserData(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "User not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "User data deleted"})
}

func (s *Server) adminVerification(w http.ResponseWriter, r *http.Request) (*gormmodel.VerificationRequest, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return nil, false
	}
	verification, err := s.idswyft.GetVerification(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return nil, false
	}
	return verification, true
}

func adminVerificationList(rows []gormmodel.VerificationRequest) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, verificationPublic(&rows[i]))
	}
	return out
}
