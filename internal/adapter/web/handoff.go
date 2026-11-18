package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

// handleHandoffCreate 实现 POST /api/verify/handoff/create。
func (s *Server) handleHandoffCreate(w http.ResponseWriter, r *http.Request) {
	if s.idswyft == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "handoff store unavailable"})
		return
	}
	var req struct {
		APIKey         string `json:"api_key"`
		UserID         string `json:"user_id"`
		Source         string `json:"source"`
		VerificationID string `json:"verification_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	var verificationID *string
	if req.VerificationID != "" {
		id, err := uuid.Parse(req.VerificationID)
		if err != nil {
			writeError(w, r, newValidationError("verification_id", "verification_id 必须是 UUID"))
			return
		}
		value := id.String()
		verificationID = &value
	}
	var apiKeyID uuid.UUID
	userID := req.UserID
	if sessionToken := r.Header.Get("X-Session-Token"); sessionToken != "" {
		if !hexTokenPattern.MatchString(sessionToken) {
			writeError(w, r, newAuthError("session token 格式无效"))
			return
		}
		session, err := s.idswyft.FindSessionByToken(r.Context(), hashToken(s.apiKeySecret(), sessionToken))
		if err != nil || session.ExpiresAt == nil || session.ExpiresAt.Before(time.Now()) {
			writeError(w, r, newAuthError("session token 无效或已过期"))
			return
		}
		if session.APIKeyID == nil {
			writeError(w, r, newAuthError("session 未关联 API Key"))
			return
		}
		apiKeyID = *session.APIKeyID
		if userID == "" && session.UserID != nil {
			userID = session.UserID.String()
		}
	} else {
		if req.APIKey == "" || userID == "" {
			writeError(w, r, newValidationError("body", "api_key 与 user_id 必填"))
			return
		}
		key, err := s.idswyft.FindAPIKeyByHash(r.Context(), hashToken(s.apiKeySecret(), req.APIKey))
		if err != nil || key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			writeError(w, r, newAuthError("API Key 无效或已过期"))
			return
		}
		apiKeyID = key.ID
	}
	source := req.Source
	switch source {
	case "api", "vaas", "demo":
	default:
		source = "api"
	}
	token, hash, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	expires := time.Now().UTC().Add(30 * time.Minute)
	session := &gormmodel.MobileHandoffSession{
		Token: hash, APIKeyID: apiKeyID, UserID: userID, Source: source,
		ExpiresAt: expires, Status: "pending", VerificationID: verificationID,
	}
	if err := s.idswyft.CreateHandoff(r.Context(), session); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "expires_at": expires})
}

// handleHandoffSession 实现 GET /api/verify/handoff/:token/session。
func (s *Server) handleHandoffSession(w http.ResponseWriter, r *http.Request) {
	session, hash, ok := s.loadHandoff(w, r)
	if !ok {
		return
	}
	if session.ExpiresAt.Before(time.Now()) {
		_, _ = s.idswyft.UpdateHandoff(r.Context(), hash, map[string]any{"status": "expired"})
		writeJSON(w, http.StatusGone, map[string]any{"error": "Session expired"})
		return
	}
	if session.Status != "pending" {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "Session already used"})
		return
	}
	response := map[string]any{
		"user_id": session.UserID, "source": firstNonEmpty(session.Source, "api"),
		"branding": nil, "page_builder_config": nil,
	}
	if developer, err := s.handoffDeveloper(r, session.APIKeyID); err == nil && developer != nil {
		response["branding"] = map[string]any{
			"logo_url": developer.BrandingLogoURL, "accent_color": developer.BrandingAccentColor,
			"company_name": firstNonEmpty(stringValue(developer.BrandingCompanyName), stringValue(developer.Company)),
		}
		if len(developer.PageBuilderConfig) > 0 {
			response["page_builder_config"] = developer.PageBuilderConfig
		}
	}
	if session.VerificationID != nil {
		response["verification_id"] = *session.VerificationID
	}
	writeJSON(w, http.StatusOK, response)
}

// handleHandoffLink 实现 PATCH /api/verify/handoff/:token/link。
func (s *Server) handleHandoffLink(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if !hexTokenPattern.MatchString(token) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found"})
		return
	}
	var req struct {
		VerificationID string `json:"verification_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	id, err := uuid.Parse(req.VerificationID)
	if err != nil {
		writeError(w, r, newValidationError("verification_id", "verification_id 必须是 UUID"))
		return
	}
	hash := hashToken(s.apiKeySecret(), token)
	affected, err := s.idswyft.UpdatePendingHandoff(r.Context(), hash, map[string]any{"verification_id": id.String()})
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found or already completed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleHandoffComplete 实现 PATCH /api/verify/handoff/:token/complete。
func (s *Server) handleHandoffComplete(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if !hexTokenPattern.MatchString(token) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found"})
		return
	}
	var req struct {
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.Status != "completed" && req.Status != "failed" {
		writeError(w, r, newValidationError("status", "status 必须是 completed 或 failed"))
		return
	}
	if len(req.Result) > 4096 {
		writeError(w, r, newValidationError("result", "result 不能超过 4096 字节"))
		return
	}
	hash := hashToken(s.apiKeySecret(), token)
	session, err := s.idswyft.GetHandoffByToken(r.Context(), hash)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found"})
		return
	}
	if session.ExpiresAt.Before(time.Now()) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "Session expired"})
		return
	}
	updates := map[string]any{"status": req.Status, "result": req.Result}
	affected, err := s.idswyft.UpdatePendingHandoff(r.Context(), hash, updates)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "Session already completed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleHandoffStatus 实现 GET /api/verify/handoff/:token/status。
func (s *Server) handleHandoffStatus(w http.ResponseWriter, r *http.Request) {
	session, hash, ok := s.loadHandoff(w, r)
	if !ok {
		return
	}
	if session.ExpiresAt.Before(time.Now()) {
		_, _ = s.idswyft.UpdateHandoff(r.Context(), hash, map[string]any{"status": "expired"})
		writeJSON(w, http.StatusGone, map[string]any{"error": "Session expired"})
		return
	}
	response := map[string]any{"status": session.Status}
	if session.VerificationID != nil {
		response["verification_id"] = *session.VerificationID
	}
	if session.Status != "pending" {
		response["result"] = session.Result
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) loadHandoff(w http.ResponseWriter, r *http.Request) (*gormmodel.MobileHandoffSession, string, bool) {
	if s.idswyft == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "handoff store unavailable"})
		return nil, "", false
	}
	token := r.PathValue("token")
	if !hexTokenPattern.MatchString(token) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found"})
		return nil, "", false
	}
	hash := hashToken(s.apiKeySecret(), token)
	session, err := s.idswyft.GetHandoffByToken(r.Context(), hash)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Session not found"})
		return nil, "", false
	}
	return session, hash, true
}

func (s *Server) handoffDeveloper(r *http.Request, apiKeyID uuid.UUID) (*gormmodel.Developer, error) {
	return s.idswyft.GetDeveloperByAPIKey(r.Context(), apiKeyID)
}

func newValidationError(_, message string) error {
	return domain.New(domain.CodeDocumentInvalid, message)
}

func newAuthError(message string) error {
	return domain.New(domain.CodeAuthInvalid, message)
}
