package web

import (
	"net/http"
	"strings"
	"time"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// handleDeveloperProfileGet 实现 GET /api/developer/profile。
func (s *Server) handleDeveloperProfileGet(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Developer authentication required"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": developerProfile(developer)})
}

// handleDeveloperProfileUpdate 实现 PUT /api/developer/profile。
func (s *Server) handleDeveloperProfileUpdate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Name    string  `json:"name"`
		Company *string `json:"company"`
		Locale  string  `json:"locale"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 || len(req.Name) > 100 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Name must be between 2 and 100 characters"})
		return
	}
	if req.Locale != "" && req.Locale != "en" && req.Locale != "zh-CN" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Locale must be en or zh-CN"})
		return
	}
	updates := map[string]any{"name": req.Name, "updated_at": time.Now().UTC()}
	if req.Company != nil {
		updates["company"] = strings.TrimSpace(*req.Company)
	}
	if req.Locale != "" {
		updates["locale"] = req.Locale
	}
	if err := s.idswyft.UpdateDeveloper(r.Context(), principal.DeveloperID, updates); err != nil {
		writeError(w, r, err)
		return
	}
	developer, _ := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": developerProfile(developer)})
}

// handleDeveloperAvatar 实现 POST /api/developer/avatar 与 branding/logo。
func (s *Server) handleDeveloperAvatar(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal := principalFrom(r)
		if s.storage == nil || !s.cfg.Storage.Enabled {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "storage unavailable"})
			return
		}
		data, _, mime, _, err := readMultipartImage(w, r, "file")
		if err != nil {
			writeError(w, r, err)
			return
		}
		key := "avatars/" + principal.DeveloperID.String()
		if kind == "branding" {
			key = "branding/" + principal.DeveloperID.String()
		}
		if err := s.storage.Put(r.Context(), key, data, mime); err != nil {
			writeError(w, r, err)
			return
		}
		url, err := s.storage.PresignGet(r.Context(), key, 7*24*time.Hour)
		if err != nil {
			writeError(w, r, err)
			return
		}
		updates := map[string]any{"avatar_url": key}
		if kind == "branding" {
			updates = map[string]any{"branding_logo_url": key}
		}
		if err := s.idswyft.UpdateDeveloper(r.Context(), principal.DeveloperID, updates); err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"url": url}})
	}
}

// handleDeveloperAccountDelete 实现 DELETE /api/developer/account。
func (s *Server) handleDeveloperAccountDelete(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		ConfirmEmail string `json:"confirm_email"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Developer authentication required"})
		return
	}
	if !strings.EqualFold(req.ConfirmEmail, developer.Email) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Email does not match your account email"})
		return
	}
	affected, err := s.idswyft.DeleteDeveloper(r.Context(), principal.DeveloperID)
	if err != nil || affected == 0 {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Failed to delete account"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Account deleted"})
}

func developerProfile(developer *gormmodel.Developer) map[string]any {
	if developer == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id": developer.ID, "email": developer.Email, "name": developer.Name,
		"company": developer.Company, "avatar_url": developer.AvatarURL,
		"locale": developer.Locale, "created_at": developer.CreatedAt,
	}
}
