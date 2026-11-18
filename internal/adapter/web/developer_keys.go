package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleDeveloperKeyCreate 实现 POST /api/developer/api-key。
func (s *Server) handleDeveloperKeyCreate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Name          string `json:"name"`
		IsSandbox     bool   `json:"is_sandbox"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 100 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "API key name is required and must be less than 100 characters"})
		return
	}
	if req.ExpiresInDays < 0 || req.ExpiresInDays > 365 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "expires_in_days must be between 1 and 365"})
		return
	}
	keys, err := s.idswyft.DeveloperAPIKeys(r.Context(), principal.DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	active := 0
	for _, key := range keys {
		if key.IsActive != nil && *key.IsActive {
			active++
		}
	}
	maxKeys := 3
	if req.IsSandbox {
		maxKeys = 10
	}
	if active >= maxKeys {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "API key limit reached"})
		return
	}
	raw, hash, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	apiKey := "ik_" + raw
	keyHash := hashToken(s.apiKeySecret(), apiKey)
	_ = hash
	now := time.Now().UTC()
	key := &gormmodel.APIKey{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, Name: req.Name,
		KeyHash: keyHash, KeyPrefix: apiKey[:8], IsActive: boolPtr(true),
		IsSandbox: &req.IsSandbox, CreatedAt: &now,
	}
	if req.ExpiresInDays > 0 {
		expires := now.AddDate(0, 0, req.ExpiresInDays)
		key.ExpiresAt = &expires
	}
	if err := s.idswyft.CreateDeveloperAPIKey(r.Context(), key); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"api_key": apiKey, "key_id": key.ID, "name": key.Name, "is_sandbox": key.IsSandbox,
		"created_at": key.CreatedAt, "expires_at": key.ExpiresAt, "key_prefix": key.KeyPrefix,
		"message": "API key created successfully. Store it securely - it will not be shown again.",
	})
}

// handleDeveloperKeyList 实现 GET /api/developer/api-keys。
func (s *Server) handleDeveloperKeyList(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	keys, err := s.idswyft.DeveloperAPIKeys(r.Context(), principal.DeveloperID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(keys))
	active := 0
	for _, key := range keys {
		if key.IsActive == nil || !*key.IsActive {
			continue
		}
		status := "active"
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			status = "expired"
		} else {
			active++
		}
		out = append(out, map[string]any{
			"id": key.ID, "name": key.Name, "key_preview": key.KeyPrefix + "...",
			"is_sandbox": key.IsSandbox, "is_active": *key.IsActive, "last_used_at": key.LastUsedAt,
			"created_at": key.CreatedAt, "expires_at": key.ExpiresAt, "status": status,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_keys": out, "total_keys": len(out), "active_keys": active})
}

// handleDeveloperKeyRevoke 实现 DELETE /api/developer/api-key/:keyId。
func (s *Server) handleDeveloperKeyRevoke(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, err := uuid.Parse(r.PathValue("keyId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid API key ID format"})
		return
	}
	affected, err := s.idswyft.RevokeDeveloperAPIKey(r.Context(), principal.DeveloperID, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "API key not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "API key revoked successfully"})
}

// handleDeveloperKeyRotate 实现 POST /api/developer/api-key/:id/rotate。
func (s *Server) handleDeveloperKeyRotate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid API key ID format"})
		return
	}
	old, err := s.idswyft.GetDeveloperAPIKey(r.Context(), principal.DeveloperID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "API key not found"})
		return
	}
	if _, err := s.idswyft.RevokeDeveloperAPIKey(r.Context(), principal.DeveloperID, id); err != nil {
		writeError(w, r, err)
		return
	}
	raw, _, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	apiKey := "ik_" + raw
	now := time.Now().UTC()
	key := &gormmodel.APIKey{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, Name: old.Name,
		KeyHash: hashToken(s.apiKeySecret(), apiKey), KeyPrefix: apiKey[:8],
		IsActive: boolPtr(true), IsSandbox: old.IsSandbox, ExpiresAt: old.ExpiresAt, CreatedAt: &now,
	}
	if err := s.idswyft.CreateDeveloperAPIKey(r.Context(), key); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"api_key": apiKey, "key_id": key.ID, "name": key.Name, "is_sandbox": key.IsSandbox,
		"created_at": key.CreatedAt, "expires_at": key.ExpiresAt, "key_prefix": key.KeyPrefix,
		"message": "API key rotated successfully. Store it securely - it will not be shown again.",
	})
}
