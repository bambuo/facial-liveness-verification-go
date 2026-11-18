package web

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleReviewerInvite 实现 POST /api/developer/reviewers/invite。
func (s *Server) handleReviewerInvite(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Email string  `json:"email"`
		Name  *string `json:"name"`
		Role  string  `json:"role"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "email is required"})
		return
	}
	if req.Role == "" {
		req.Role = "reviewer"
	}
	now := time.Now().UTC()
	reviewer := &gormmodel.VerificationReviewer{
		ID: uuid.New(), DeveloperID: principal.DeveloperID, Email: req.Email, Name: req.Name,
		Role: req.Role, Status: "invited", InvitedAt: &now,
	}
	if err := s.idswyft.CreateReviewer(r.Context(), reviewer); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "reviewer": reviewer})
}

// handleReviewerList 实现 GET /api/developer/reviewers。
func (s *Server) handleReviewerList(w http.ResponseWriter, r *http.Request) {
	reviewers, err := s.idswyft.Reviewers(r.Context(), principalFrom(r).DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reviewers": reviewers})
}

// handleReviewerDelete 实现 DELETE /api/developer/reviewers/:id。
func (s *Server) handleReviewerDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Reviewer not found"})
		return
	}
	affected, err := s.idswyft.DeleteReviewer(r.Context(), principalFrom(r).DeveloperID, id)
	if err != nil || affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Reviewer not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Reviewer removed"})
}

// handlePageConfig 实现 GET /api/v2/verify/page-config?api_key=...
func (s *Server) handlePageConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("api_key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "api_key is required"})
		return
	}
	record, err := s.idswyft.FindAPIKeyByHash(r.Context(), hashToken(s.apiKeySecret(), key))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Invalid API key"})
		return
	}
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), record.DeveloperID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Developer not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "config": developer.PageBuilderConfig,
		"branding": map[string]any{
			"logo_url": developer.BrandingLogoURL, "accent_color": developer.BrandingAccentColor,
			"company_name": developer.BrandingCompanyName,
		},
	})
}

// handleSessionInfo 实现 GET /api/v2/verify/session-info。
func (s *Server) handleSessionInfo(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Session-Token")
	if token == "" {
		token = r.Header.Get("X-Handoff-Token")
	}
	if !hexTokenPattern.MatchString(token) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "session token required"})
		return
	}
	if session, err := s.idswyft.FindSessionByToken(r.Context(), hashToken(s.apiKeySecret(), token)); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"verification_id": session.VerificationID, "developer_id": session.DeveloperID})
		return
	}
	if handoff, err := s.idswyft.FindHandoffByToken(r.Context(), hashToken(s.apiKeySecret(), token)); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"verification_id": handoff.VerificationID, "developer_id": handoff.APIKey.DeveloperID})
		return
	}
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid session token"})
}

// handleAPIKeyDiagnostic 实现公开 HMAC 诊断，不回显密钥。
func (s *Server) handleAPIKeyDiagnostic(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "key is required"})
		return
	}
	_, err := s.idswyft.FindAPIKeyByHash(r.Context(), hashToken(s.apiKeySecret(), key))
	writeJSON(w, http.StatusOK, map[string]any{"valid": err == nil, "key_prefix": key[:min(8, len(key))]})
}

// handleLogout 清除浏览器认证 Cookie。
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "idswyft_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleCSRFToken 生成一次性 CSRF token。
func (s *Server) handleCSRFToken(w http.ResponseWriter, r *http.Request) {
	token, _, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "idswyft_csrf", Value: token, Path: "/", HttpOnly: false, SameSite: http.SameSiteLaxMode})
	writeJSON(w, http.StatusOK, map[string]any{"csrf_token": token})
}
