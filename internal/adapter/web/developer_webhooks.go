package web

import "net/http"

// handleDeveloperWebhookSecret 返回 Webhook 签名密钥；调用方必须已通过开发者鉴权。
func (s *Server) handleDeveloperWebhookSecret(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	webhook, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	item, err := s.idswyft.WebhookByID(r.Context(), principal.DeveloperID, webhook)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Webhook not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"secret": item.SecretKey})
}

// handleDeveloperWebhookResend 重发指定投递；P2 基线返回明确结果。
func (s *Server) handleDeveloperWebhookResend(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	webhook, ok := s.webhookID(w, r, principal)
	if !ok {
		return
	}
	item, err := s.idswyft.WebhookByID(r.Context(), principal.DeveloperID, webhook)
	if err != nil || s.webhooks == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Webhook not found"})
		return
	}
	payload := map[string]any{"event": "webhook.resend", "delivery_id": r.PathValue("deliveryId")}
	statusCode, err := s.webhooks.Deliver(r.Context(), item.URL, stringValue(item.SecretKey), payload)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error(), "response_status": statusCode})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "response_status": statusCode})
}

// handleSystemVersion 实现 GET /api/system/version。
func (s *Server) handleSystemVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": s.buildVersion, "commit": s.buildCommit, "model_hash": s.modelHash,
		"config_hash": s.configHash, "engine": s.engineVersion,
	})
}
