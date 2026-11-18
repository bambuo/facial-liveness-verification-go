package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

// handleVaultStore 实现 POST /api/v2/vault/store。
func (s *Server) handleVaultStore(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
		return
	}
	var req struct {
		VerificationID string `json:"verification_id"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	verificationID, err := uuid.Parse(req.VerificationID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "verification_id 必须是 UUID"))
		return
	}
	verification, err := s.loadOwnedVerification(r.Context(), verificationID, principal)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return
	}
	if verification.Status != "verified" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Only verified identities can be stored in the vault"})
		return
	}
	existing, _ := s.idswyft.GetVaultByVerification(r.Context(), principal.DeveloperID, verificationID)
	if existing != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "already_stored": true, "vault_token": existing.VaultToken,
			"message": "Identity is already in the vault.",
		})
		return
	}
	identity, err := s.identityData(r.Context(), verification)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Verification is missing required identity data (OCR)"})
		return
	}
	plain, _ := json.Marshal(identity)
	if s.encryptor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "vault encryption unavailable"})
		return
	}
	cipherText, err := s.encryptor.Encrypt(r.Context(), principal.DeveloperID, plain)
	if err != nil {
		writeError(w, r, err)
		return
	}
	token, err := randomToken("ivt_")
	if err != nil {
		writeError(w, r, err)
		return
	}
	now := time.Now().UTC()
	entry := &gormmodel.IdentityVault{
		ID: uuid.New(), DeveloperID: principal.DeveloperID, VerificationID: verificationID,
		VaultToken: token, EncryptedData: base64.StdEncoding.EncodeToString(cipherText),
		EncryptionKeyID: "v1", Status: "active", CreatedAt: now,
	}
	if err := s.idswyft.CreateVaultEntry(r.Context(), entry); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "vault_token": token})
}

// handleVaultGet 实现 GET /api/v2/vault/:token。
func (s *Server) handleVaultGet(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
		return
	}
	entry, ok := s.loadVault(w, r, principal)
	if !ok {
		return
	}
	data, err := s.decryptVault(r, principal.DeveloperID, entry)
	if err != nil {
		writeError(w, r, err)
		return
	}
	_ = s.idswyft.TouchVaultAccess(r.Context(), entry.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "vault_token": entry.VaultToken, "status": entry.Status,
		"data": data, "created_at": entry.CreatedAt, "expires_at": entry.ExpiresAt,
	})
}

// handleVaultAttribute 实现 GET /api/v2/vault/:token/attributes/:attr。
func (s *Server) handleVaultAttribute(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
		return
	}
	entry, ok := s.loadVault(w, r, principal)
	if !ok {
		return
	}
	data, err := s.decryptVault(r, principal.DeveloperID, entry)
	if err != nil {
		writeError(w, r, err)
		return
	}
	attr := r.PathValue("attr")
	value, exists := data[attr]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": fmt.Sprintf("Attribute '%s' not available", attr)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "attribute": attr, "value": value})
}

// handleVaultDelete 实现 DELETE /api/v2/vault/:token。
func (s *Server) handleVaultDelete(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
		return
	}
	token := r.PathValue("token")
	if !strings.HasPrefix(token, "ivt_") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid vault token format"})
		return
	}
	affected, err := s.idswyft.DeleteVault(r.Context(), principal.DeveloperID, token)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Vault entry not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Vault entry permanently deleted"})
}

// handleVaultList 实现 GET /api/v2/vault。
func (s *Server) handleVaultList(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
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
	entries, total, err := s.idswyft.ListVault(r.Context(), principal.DeveloperID, page, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	type entryView struct {
		VaultToken     string     `json:"vault_token"`
		VerificationID uuid.UUID  `json:"verification_id"`
		Status         string     `json:"status"`
		AccessCount    int32      `json:"access_count"`
		LastAccessedAt *time.Time `json:"last_accessed_at"`
		ExpiresAt      *time.Time `json:"expires_at"`
		CreatedAt      time.Time  `json:"created_at"`
	}
	out := make([]entryView, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entryView{
			VaultToken: entry.VaultToken, VerificationID: entry.VerificationID, Status: entry.Status,
			AccessCount: entry.AccessCount, LastAccessedAt: entry.LastAccessedAt,
			ExpiresAt: entry.ExpiresAt, CreatedAt: entry.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "entries": out,
		"pagination": map[string]any{"page": page, "limit": limit, "total": total, "total_pages": (total + int64(limit) - 1) / int64(limit)},
	})
}

// handleVaultShareCreate 实现 POST /api/v2/vault/:token/share。
func (s *Server) handleVaultShareCreate(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if !s.requireVaultEnabled(w, r, principal) {
		return
	}
	entry, ok := s.loadVault(w, r, principal)
	if !ok {
		return
	}
	if entry.Status != "active" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Vault entry is not active"})
		return
	}
	var req struct {
		Attributes     []string `json:"attributes"`
		ExpiresIn      int      `json:"expires_in"`
		RecipientLabel string   `json:"recipient_label"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if len(req.Attributes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "attributes array is required"})
		return
	}
	if req.ExpiresIn == 0 {
		req.ExpiresIn = 86400
	}
	if req.ExpiresIn < 60 || req.ExpiresIn > 604800 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "expires_in must be between 60 and 604800 seconds"})
		return
	}
	shareToken, err := randomToken("ivs_")
	if err != nil {
		writeError(w, r, err)
		return
	}
	expires := time.Now().UTC().Add(time.Duration(req.ExpiresIn) * time.Second)
	maxAccesses := int32(10)
	link := &gormmodel.VaultShareLink{
		ID: uuid.New(), VaultID: entry.ID, DeveloperID: principal.DeveloperID,
		ShareToken: shareToken, AllowedAttributes: req.Attributes, RecipientLabel: strPtrOrNil(req.RecipientLabel),
		ExpiresAt: expires, CreatedAt: time.Now().UTC(), MaxAccesses: &maxAccesses,
	}
	if err := s.idswyft.CreateShareLink(r.Context(), link); err != nil {
		writeError(w, r, err)
		return
	}
	base := "http://" + r.Host
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true, "share_url": base + "/api/v2/vault/share/" + shareToken,
		"share_token": shareToken, "expires_at": expires, "allowed_attributes": req.Attributes,
	})
}

// handleVaultShareGet 实现公开 GET /api/v2/vault/share/:shareToken。
func (s *Server) handleVaultShareGet(w http.ResponseWriter, r *http.Request) {
	link, err := s.idswyft.GetShareLink(r.Context(), r.PathValue("shareToken"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Share link not found"})
		return
	}
	if link.ExpiresAt.Before(time.Now()) {
		writeJSON(w, http.StatusGone, map[string]any{"error": "Share link has expired"})
		return
	}
	if link.MaxAccesses != nil && link.AccessCount >= *link.MaxAccesses {
		writeJSON(w, http.StatusGone, map[string]any{"error": "Share link access limit reached"})
		return
	}
	affected, err := s.idswyft.IncrementShareAccess(r.Context(), link.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusGone, map[string]any{"error": "Share link access limit reached"})
		return
	}
	entry, err := s.idswyft.GetVaultByID(r.Context(), link.VaultID)
	if err != nil || entry.Status != "active" || entry.ExpiresAt != nil && entry.ExpiresAt.Before(time.Now()) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Vault entry no longer available"})
		return
	}
	data, err := s.decryptVault(r, link.DeveloperID, entry)
	if err != nil {
		writeError(w, r, err)
		return
	}
	shared := map[string]any{}
	for _, attr := range link.AllowedAttributes {
		if value, ok := data[attr]; ok {
			shared[attr] = value
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "attributes": shared, "expires_at": link.ExpiresAt})
}

func (s *Server) requireVaultEnabled(w http.ResponseWriter, r *http.Request, principal *idswyftPrincipal) bool {
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	if err != nil || developer.VaultEnabled == nil || !*developer.VaultEnabled {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "Identity Vault is not enabled for your account. Enable it in Settings."})
		return false
	}
	return true
}

func (s *Server) loadVault(w http.ResponseWriter, r *http.Request, principal *idswyftPrincipal) (*gormmodel.IdentityVault, bool) {
	token := r.PathValue("token")
	if !strings.HasPrefix(token, "ivt_") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid vault token format"})
		return nil, false
	}
	entry, err := s.idswyft.GetVaultByToken(r.Context(), principal.DeveloperID, token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Vault entry not found"})
		return nil, false
	}
	return entry, true
}

func (s *Server) decryptVault(r *http.Request, developerID uuid.UUID, entry *gormmodel.IdentityVault) (map[string]any, error) {
	if s.encryptor == nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "vault encryption unavailable")
	}
	cipherText, err := base64.StdEncoding.DecodeString(entry.EncryptedData)
	if err != nil {
		return nil, err
	}
	plain, err := s.encryptor.Decrypt(r.Context(), developerID, cipherText)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(plain, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Server) identityData(ctx context.Context, verification *gormmodel.VerificationRequest) (map[string]any, error) {
	if verification.DocumentID == nil {
		return nil, domain.New(domain.CodeDocumentInvalid, "missing document")
	}
	document, err := s.idswyft.GetDocument(ctx, *verification.DocumentID)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(document.OCRData, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func randomToken(prefix string) (string, error) {
	raw, _, err := newSessionToken("vault-token")
	if err != nil {
		return "", err
	}
	return prefix + raw, nil
}

func strPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
