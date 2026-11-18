package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// requireService 校验 X-Service-Token。
func (s *Server) requireService(w http.ResponseWriter, r *http.Request) bool {
	if !s.authenticateServiceToken(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "unauthorized"})
		return false
	}
	return true
}

// handleVAASCreateUser 实现 POST /api/vaas/users。
func (s *Server) handleVAASCreateUser(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	var req struct {
		Email     string         `json:"email"`
		Phone     string         `json:"phone"`
		FirstName string         `json:"first_name"`
		LastName  string         `json:"last_name"`
		External  string         `json:"external_id"`
		Metadata  map[string]any `json:"metadata"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}
	now := time.Now().UTC()
	user := &gormmodel.User{ID: uuid.New(), Email: &req.Email, FirstName: &req.FirstName, LastName: &req.LastName, CreatedAt: &now}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if err := s.idswyft.CreateVAASUser(r.Context(), user); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Failed to create user"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]any{
		"id": user.ID, "email": req.Email, "phone": req.Phone, "first_name": req.FirstName,
		"last_name": req.LastName, "external_id": req.External, "metadata": req.Metadata,
	}})
}

// handleVAASVerify 实现 POST /api/vaas/verify。
func (s *Server) handleVAASVerify(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	var req struct {
		UserID   string `json:"user_id"`
		Document string `json:"document_url"`
		Selfie   string `json:"selfie_url"`
		Addons   any    `json:"addons"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "user_id must be a UUID"})
		return
	}
	verification := &gormmodel.VerificationRequest{
		ID: uuid.New(), UserID: &userID, Status: "pending", Source: "vaas",
	}
	if err := s.idswyft.CreateVerification(r.Context(), verification); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Failed to create verification"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID, "status": verification.Status,
		"message": "Verification request submitted successfully",
	})
}

// handleVAASStatus 实现 GET /api/vaas/verify/:verification_id/status。
func (s *Server) handleVAASStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("verification_id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Verification not found"})
		return
	}
	verification, err := s.idswyft.GetVerification(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Verification not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "verification": map[string]any{
		"id": verification.ID, "status": verification.Status, "confidence_score": verification.ConfidenceScore,
		"failure_reason": verification.FailureReason, "updated_at": verification.UpdatedAt,
	}})
}

// handleVAASDevelopers 实现 GET /api/vaas/developers。
func (s *Server) handleVAASDevelopers(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	developers, total, err := s.idswyft.ListDevelopers(r.Context(), r.URL.Query().Get("search"), r.URL.Query().Get("status"), page, perPage)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Failed to list developers"})
		return
	}
	data := make([]map[string]any, 0, len(developers))
	for _, developer := range developers {
		data = append(data, map[string]any{
			"id": developer.ID, "email": developer.Email, "name": developer.Name, "company": developer.Company,
			"status": developer.Status, "is_verified": developer.IsVerified, "avatar_url": developer.AvatarURL,
			"created_at": developer.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": data, "meta": map[string]any{"total": total, "page": page, "per_page": perPage}})
}

// handleVAASDeveloper 实现 GET /api/vaas/developers/:id。
func (s *Server) handleVAASDeveloper(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	developer, ok := s.vaasDeveloper(w, r)
	if !ok {
		return
	}
	keys, _ := s.idswyft.CountAPIKeys(r.Context(), developer.ID)
	verifications, _ := s.idswyft.CountVerifications(r.Context(), developer.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"id": developer.ID, "email": developer.Email, "name": developer.Name, "company": developer.Company,
		"status": developer.Status, "avatar_url": developer.AvatarURL, "api_key_count": keys, "verification_count": verifications,
	}})
}

// handleVAASSuspend 与 unsuspend 共用状态更新。
func (s *Server) handleVAASDeveloperStatus(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.requireService(w, r) {
			return
		}
		developer, ok := s.vaasDeveloper(w, r)
		if !ok {
			return
		}
		affected, err := s.idswyft.SetDeveloperStatus(r.Context(), developer.ID, status)
		if err != nil || affected == 0 {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Developer not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"id": developer.ID, "status": status}})
	}
}

// handleVAASHealth 实现 GET /api/vaas/health。
func (s *Server) handleVAASHealth(w http.ResponseWriter, r *http.Request) {
	if !s.requireService(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "ok", "timestamp": time.Now().UTC()})
}

func (s *Server) vaasDeveloper(w http.ResponseWriter, r *http.Request) (*gormmodel.Developer, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Developer not found"})
		return nil, false
	}
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Developer not found"})
		return nil, false
	}
	return developer, true
}
