package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleWebhookRegister 实现 POST /api/webhooks/register。
func (s *Server) handleWebhookRegister(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		URL         string `json:"url"`
		IsSandbox   bool   `json:"is_sandbox"`
		SecretToken string `json:"secret_token"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !safeHTTPURL(req.URL) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Valid webhook URL is required"})
		return
	}
	if req.SecretToken != "" && (len(req.SecretToken) < 10 || len(req.SecretToken) > 100) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Secret token must be between 10 and 100 characters"})
		return
	}
	if existing, _ := s.idswyft.WebhookByURL(r.Context(), principal.DeveloperID, req.URL, req.IsSandbox); existing != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Webhook already exists for this URL"})
		return
	}
	now := time.Now().UTC()
	webhook := &gormmodel.Webhook{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, URL: req.URL,
		IsSandbox: &req.IsSandbox, IsActive: boolPtr(true), CreatedAt: &now,
		Events: pq.StringArray{"verification.completed", "verification.failed", "verification.manual_review"},
	}
	if req.SecretToken != "" {
		webhook.SecretKey = &req.SecretToken
	}
	if principal.IsService {
		webhook.APIKeyID = &principal.APIKeyID
	}
	if err := s.idswyft.CreateWebhook(r.Context(), webhook); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"webhook": webhookView(webhook), "message": "Webhook registered successfully"})
}

// handleWebhookList 实现 GET /api/webhooks。
func (s *Server) handleWebhookList(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	webhooks, err := s.idswyft.WebhooksByDeveloper(r.Context(), principal.DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(webhooks))
	for i := range webhooks {
		out = append(out, webhookView(&webhooks[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"webhooks": out})
}

// handleWebhookUpdate 实现 PUT /api/webhooks/:webhookId。
func (s *Server) handleWebhookUpdate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	var req struct {
		URL         *string `json:"url"`
		IsActive    *bool   `json:"is_active"`
		SecretToken *string `json:"secret_token"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{}
	if req.URL != nil {
		if !safeHTTPURL(*req.URL) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Valid webhook URL is required"})
			return
		}
		updates["url"] = *req.URL
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SecretToken != nil {
		if len(*req.SecretToken) < 10 || len(*req.SecretToken) > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Secret token must be between 10 and 100 characters"})
			return
		}
		updates["secret_key"] = *req.SecretToken
	}
	if _, err := s.idswyft.UpdateWebhook(r.Context(), principal.DeveloperID, id, updates); err != nil {
		writeError(w, r, err)
		return
	}
	webhook, _ := s.idswyft.WebhookByID(r.Context(), principal.DeveloperID, id)
	writeJSON(w, http.StatusOK, map[string]any{"webhook": webhookView(webhook), "message": "Webhook updated successfully"})
}

// handleWebhookDelete 实现 DELETE /api/webhooks/:webhookId。
func (s *Server) handleWebhookDelete(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	if _, err := s.idswyft.DeleteWebhook(r.Context(), principal.DeveloperID, id); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Webhook deleted successfully"})
}

// handleWebhookDeliveries 实现 GET /api/webhooks/:webhookId/deliveries。
func (s *Server) handleWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	deliveries, total, err := s.idswyft.WebhookDeliveries(r.Context(), id, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"webhook_id": id.String(), "page": page, "limit": limit, "total": total,
		"deliveries": deliveries,
	})
}

// handleWebhookTest 实现 POST /api/webhooks/:webhookId/test。
func (s *Server) handleWebhookTest(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	webhook, err := s.idswyft.WebhookByID(r.Context(), principal.DeveloperID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Webhook not found"})
		return
	}
	payload := map[string]any{
		"user_id": "test-user-123", "verification_id": "test-verification-456",
		"status": "verified", "timestamp": time.Now().UTC(),
		"data": map[string]any{"ocr_data": map[string]any{"name": "Test User"}, "face_match_score": 0.95},
	}
	statusCode := 0
	var deliverErr error
	if s.webhooks != nil {
		statusCode, deliverErr = s.webhooks.Deliver(r.Context(), webhook.URL, stringValue(webhook.SecretKey), payload)
	}
	result := map[string]any{"delivered": deliverErr == nil, "response_status": statusCode}
	if deliverErr != nil {
		result["error"] = deliverErr.Error()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": func() string {
			if deliverErr == nil {
				return "Test webhook delivered"
			}
			return "Test webhook sent, but the endpoint did not accept it"
		}(), "result": result, "payload": payload,
	})
}

func (s *Server) webhookID(w http.ResponseWriter, r *http.Request, principal *idswyftPrincipal) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("webhookId"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Webhook not found"})
		return uuid.Nil, false
	}
	if _, err := s.idswyft.WebhookByID(r.Context(), principal.DeveloperID, id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Webhook not found"})
		return uuid.Nil, false
	}
	return id, true
}

func webhookView(webhook *gormmodel.Webhook) map[string]any {
	if webhook == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id": webhook.ID, "url": webhook.URL, "is_sandbox": webhook.IsSandbox,
		"is_active": webhook.IsActive, "created_at": webhook.CreatedAt,
	}
}
