package web

import (
	"encoding/json"
	"net/http"
	"time"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleSettingsLLMGet 实现 GET /api/developer/settings/llm。
func (s *Server) handleSettingsLLMGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"provider": developer.LlmProvider, "model": developer.LlmModel,
		"endpoint_url": developer.LlmEndpointURL, "configured": developer.LlmAPIKeyEncrypted != nil,
	}})
}

// handleSettingsLLMPut 实现 PUT /api/developer/settings/llm。
func (s *Server) handleSettingsLLMPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Provider    *string `json:"provider"`
		Model       *string `json:"model"`
		EndpointURL *string `json:"endpoint_url"`
		APIKey      *string `json:"api_key"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Provider != nil {
		updates["llm_provider"] = *req.Provider
	}
	if req.Model != nil {
		updates["llm_model"] = *req.Model
	}
	if req.EndpointURL != nil {
		updates["llm_endpoint_url"] = *req.EndpointURL
	}
	if req.APIKey != nil && *req.APIKey != "" && s.encryptor != nil {
		cipherText, err := s.encryptor.Encrypt(r.Context(), developer.ID, []byte(*req.APIKey))
		if err != nil {
			writeError(w, r, err)
			return
		}
		value := string(cipherText)
		updates["llm_api_key_encrypted"] = value
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsSMSGet 实现 GET /api/developer/settings/sms。
func (s *Server) handleSettingsSMSGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"provider": developer.SmsProvider, "phone_number": developer.SmsPhoneNumber,
		"configured": developer.SmsAPIKeyEncrypted != nil,
	}})
}

// handleSettingsSMSPut 实现 PUT /api/developer/settings/sms。
func (s *Server) handleSettingsSMSPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Provider    *string `json:"provider"`
		PhoneNumber *string `json:"phone_number"`
		APIKey      *string `json:"api_key"`
		APISecret   *string `json:"api_secret"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Provider != nil {
		updates["sms_provider"] = *req.Provider
	}
	if req.PhoneNumber != nil {
		updates["sms_phone_number"] = *req.PhoneNumber
	}
	if s.encryptor != nil {
		if req.APIKey != nil && *req.APIKey != "" {
			cipherText, err := s.encryptor.Encrypt(r.Context(), developer.ID, []byte(*req.APIKey))
			if err == nil {
				updates["sms_api_key_encrypted"] = string(cipherText)
			}
		}
		if req.APISecret != nil && *req.APISecret != "" {
			cipherText, err := s.encryptor.Encrypt(r.Context(), developer.ID, []byte(*req.APISecret))
			if err == nil {
				updates["sms_api_secret_encrypted"] = string(cipherText)
			}
		}
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsBrandingGet 实现 GET /api/developer/settings/branding。
func (s *Server) handleSettingsBrandingGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"logo_url": developer.BrandingLogoURL, "accent_color": developer.BrandingAccentColor,
		"company_name": developer.BrandingCompanyName,
	}})
}

// handleSettingsBrandingPut 实现 PUT /api/developer/settings/branding。
func (s *Server) handleSettingsBrandingPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		LogoURL     *string `json:"logo_url"`
		AccentColor *string `json:"accent_color"`
		CompanyName *string `json:"company_name"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.LogoURL != nil {
		updates["branding_logo_url"] = *req.LogoURL
	}
	if req.AccentColor != nil {
		updates["branding_accent_color"] = *req.AccentColor
	}
	if req.CompanyName != nil {
		updates["branding_company_name"] = *req.CompanyName
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsPageBuilderGet 实现 GET /api/developer/settings/page-builder。
func (s *Server) handleSettingsPageBuilderGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"config": developer.PageBuilderConfig, "slug": developer.VerificationSlug,
	}})
}

// handleSettingsPageBuilderPut 实现 PUT /api/developer/settings/page-builder。
func (s *Server) handleSettingsPageBuilderPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Config json.RawMessage `json:"config"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, map[string]any{"page_builder_config": req.Config, "updated_at": time.Now().UTC()}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsPageBuilderSlug 实现 PUT /api/developer/settings/page-builder/slug。
func (s *Server) handleSettingsPageBuilderSlug(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Slug *string `json:"slug"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if req.Slug == nil || *req.Slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "slug is required"})
		return
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, map[string]any{"verification_slug": *req.Slug, "updated_at": time.Now().UTC()}); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "slug is already in use"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "slug": *req.Slug})
}

// handleSettingsAMLGet 实现 GET /api/developer/settings/aml。
func (s *Server) handleSettingsAMLGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"enabled": developer.AMLEnabled}})
}

// handleSettingsAMLPut 实现 PUT /api/developer/settings/aml。
func (s *Server) handleSettingsAMLPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil || req.Enabled == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "enabled is required"})
		return
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, map[string]any{"aml_enabled": *req.Enabled, "updated_at": time.Now().UTC()}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsVoiceGet 实现 GET /api/developer/settings/voice-auth。
func (s *Server) handleSettingsVoiceGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"enabled": developer.VoiceAuthEnabled}})
}

// handleSettingsVoicePut 实现 PUT /api/developer/settings/voice-auth。
func (s *Server) handleSettingsVoicePut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil || req.Enabled == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "enabled is required"})
		return
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, map[string]any{"voice_auth_enabled": *req.Enabled, "updated_at": time.Now().UTC()}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsVCGet 实现 GET /api/developer/settings/vc。
func (s *Server) handleSettingsVCGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"enabled": developer.VcEnabled}})
}

// handleSettingsVCPut 实现 PUT /api/developer/settings/vc。
func (s *Server) handleSettingsVCPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil || req.Enabled == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "enabled is required"})
		return
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, map[string]any{"vc_enabled": *req.Enabled, "updated_at": time.Now().UTC()}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsDuplicateGet 实现 GET /api/developer/settings/duplicate-detection。
func (s *Server) handleSettingsDuplicateGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"enabled": developer.DuplicateDetectionEnabled, "action": developer.DuplicateDetectionAction,
	}})
}

// handleSettingsDuplicatePut 实现 PUT /api/developer/settings/duplicate-detection。
func (s *Server) handleSettingsDuplicatePut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Enabled *bool   `json:"enabled"`
		Action  *string `json:"action"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Enabled != nil {
		updates["duplicate_detection_enabled"] = *req.Enabled
	}
	if req.Action != nil {
		updates["duplicate_detection_action"] = *req.Action
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleSettingsVaultGet 实现 GET /api/developer/settings/vault。
func (s *Server) handleSettingsVaultGet(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"enabled": developer.VaultEnabled, "auto_store": developer.VaultAutoStore,
	}})
}

// handleSettingsVaultPut 实现 PUT /api/developer/settings/vault。
func (s *Server) handleSettingsVaultPut(w http.ResponseWriter, r *http.Request) {
	developer, ok := s.developerForSettings(w, r)
	if !ok {
		return
	}
	var req struct {
		Enabled   *bool `json:"enabled"`
		AutoStore *bool `json:"auto_store"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if req.Enabled != nil {
		updates["vault_enabled"] = *req.Enabled
	}
	if req.AutoStore != nil {
		updates["vault_auto_store"] = *req.AutoStore
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), developer.ID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) developerForSettings(w http.ResponseWriter, r *http.Request) (*gormmodel.Developer, bool) {
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principalFrom(r).DeveloperID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Developer authentication required"})
		return nil, false
	}
	return developer, true
}
