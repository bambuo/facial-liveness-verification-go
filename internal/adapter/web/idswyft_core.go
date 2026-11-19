package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

const (
	defaultVerificationTTL  = 24 * time.Hour
	defaultPhoneOTPTTL      = 10 * time.Minute
	defaultPhoneOTPWindow   = 10 * time.Minute
	defaultPhoneOTPSends    = 3
	defaultPhoneOTPAttempts = 5
)

var phonePattern = regexp.MustCompile(`^\+?[0-9]{8,20}$`)
var otpPattern = regexp.MustCompile(`^[0-9]{6}$`)

type idswyftContextKey struct{}

// withAPIAuth 包装 API Key/handoff/session 鉴权，并把主体放入请求上下文。
func (s *Server) withAPIAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := s.authenticateAPIKeyOrHandoff(w, r)
		if !ok {
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), idswyftContextKey{}, principal)))
	}
}

// principalFrom 取出 withAPIAuth 注入的主体。
func principalFrom(r *http.Request) *idswyftPrincipal {
	principal, _ := r.Context().Value(idswyftContextKey{}).(*idswyftPrincipal)
	return principal
}

// handleVerifyInitialize 实现 POST /api/v2/verify/initialize。
func (s *Server) handleVerifyInitialize(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		UserID         string          `json:"user_id"`
		DocumentType   string          `json:"document_type"`
		IssuingCountry string          `json:"issuing_country"`
		Sandbox        *bool           `json:"sandbox"`
		Source         string          `json:"source"`
		Addons         json.RawMessage `json:"addons"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "user_id 必须是 UUID"))
		return
	}
	if err := s.idswyft.EnsureUser(r.Context(), userID); err != nil {
		writeError(w, r, err)
		return
	}
	sandbox := principal.Sandbox
	if req.Sandbox != nil {
		sandbox = *req.Sandbox
	}
	source := req.Source
	if source == "" {
		source = "api"
	}
	token, hash, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	now := time.Now().UTC()
	expires := now.Add(defaultVerificationTTL)
	verification := &gormmodel.VerificationRequest{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, APIKeyID: &principal.APIKeyID,
		UserID: &userID, Status: "pending", Source: source, IsSandbox: &sandbox,
		VerificationMode: strPtr("full"), VerificationType: strPtr("document"),
		SessionTokenHash: &hash, SessionTokenExpiresAt: &expires,
		SessionAPIKeyID: &principal.APIKeyID,
	}
	if err := s.idswyft.CreateVerification(r.Context(), verification); err != nil {
		writeError(w, r, err)
		return
	}
	ctxPayload, _ := json.Marshal(map[string]any{
		"document_type": req.DocumentType, "issuing_country": req.IssuingCountry,
		"addons": req.Addons, "current_step": "front_document", "total_steps": 5,
	})
	_ = s.idswyft.UpsertContext(r.Context(), verification.ID, ctxPayload)
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"session_token": token, "verification_url": "/verify/" + verification.ID.String(),
		"verification_mode": "full", "status": "AWAITING_FRONT",
		"current_step": 1, "total_steps": 5, "message": "Verification initialized",
	})
}

// handleVerifyReverify 实现 POST /api/v2/verify/re-verify。
func (s *Server) handleVerifyReverify(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		UserID                 string `json:"user_id"`
		PreviousVerificationID string `json:"previous_verification_id"`
		Source                 string `json:"source"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "user_id 必须是 UUID"))
		return
	}
	previousID, err := uuid.Parse(req.PreviousVerificationID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "previous_verification_id 必须是 UUID"))
		return
	}
	previous, err := s.loadOwnedVerification(r.Context(), previousID, principal)
	if err != nil {
		writeError(w, r, err)
		return
	}
	source := req.Source
	if source == "" {
		source = previous.Source
	}
	sandbox := previous.IsSandbox != nil && *previous.IsSandbox
	mode, vType := "full", "document"
	if previous.VerificationMode != nil {
		mode = *previous.VerificationMode
	}
	if previous.VerificationType != nil {
		vType = *previous.VerificationType
	}
	verification := &gormmodel.VerificationRequest{
		ID: uuid.New(), DeveloperID: &principal.DeveloperID, APIKeyID: &principal.APIKeyID,
		UserID: &userID, ParentVerificationID: &previous.ID, Status: "pending",
		Source: source, IsSandbox: &sandbox, VerificationMode: &mode, VerificationType: &vType,
	}
	if err := s.idswyft.CreateVerification(r.Context(), verification); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"parent_verification_id": previous.ID.String(), "verification_mode": mode,
		"status": "AWAITING_FRONT", "current_step": 1, "total_steps": 5,
		"message": "Re-verification initialized",
	})
}

// handleVerificationStatus 实现 GET /api/v2/verify/:id/status。
func (s *Server) handleVerificationStatus(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	payload := verificationPublic(verification)
	// 契约（statusReader.buildVerificationResponse）：前端以 ocr_data 非空判断正面证件处理完成。
	payload["front_document_uploaded"] = verification.DocumentID != nil
	payload["back_document_uploaded"] = verification.BackOfIDUploaded != nil && *verification.BackOfIDUploaded
	payload["live_capture_uploaded"] = verification.LiveCaptureCompleted != nil && *verification.LiveCaptureCompleted
	if verification.DocumentID != nil && s.idswyft != nil {
		if doc, docErr := s.idswyft.GetDocument(r.Context(), *verification.DocumentID); docErr == nil && doc != nil {
			ocrData := map[string]any{}
			_ = json.Unmarshal(doc.OCRData, &ocrData)
			payload["ocr_data"] = ocrData
		}
	}
	writeJSON(w, http.StatusOK, payload)
}

// handleVerificationSelfie 实现 GET /api/v2/verify/:id/selfie。
func (s *Server) handleVerificationSelfie(w http.ResponseWriter, r *http.Request) {
	verification, err := s.ownedVerificationFromPath(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if verification.SelfieID == nil || s.storage == nil || !s.cfg.Storage.Enabled {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "selfie not found"})
		return
	}
	selfie, err := s.idswyft.GetSelfie(r.Context(), *verification.SelfieID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "selfie not found"})
		return
	}
	url, err := s.storage.PresignGet(r.Context(), selfie.FilePath, 5*time.Minute)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"verification_id": verification.ID.String(), "selfie_url": url, "expires_in": 300,
	})
}

// handleInternalSession 实现 service-token 保护的内部分发 session。
func (s *Server) handleInternalSession(w http.ResponseWriter, r *http.Request) {
	if !s.authenticateServiceToken(r) {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "service token 无效"))
		return
	}
	id, err := pathUUID(r, "verification_id")
	if err != nil {
		writeError(w, r, err)
		return
	}
	verification, err := s.idswyft.GetVerification(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "message": "verification not found"})
		return
	}
	token, hash, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	expires := time.Now().UTC().Add(defaultVerificationTTL)
	if err := s.idswyft.UpdateVerification(r.Context(), id, map[string]any{
		"session_token_hash": hash, "session_token_expires_at": expires,
	}); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verification_id": verification.ID.String(),
		"session_token": token, "session_token_expires_at": expires,
		"verification_url": "/verify/" + verification.ID.String(),
	})
}

func (s *Server) ownedVerificationFromPath(r *http.Request) (*gormmodel.VerificationRequest, error) {
	id, err := pathUUID(r, "verification_id")
	if err != nil {
		return nil, err
	}
	return s.loadOwnedVerification(r.Context(), id, principalFrom(r))
}

func (s *Server) loadOwnedVerification(ctx context.Context, id uuid.UUID, principal *idswyftPrincipal) (*gormmodel.VerificationRequest, error) {
	if principal == nil {
		return nil, domain.New(domain.CodeAuthInvalid, "未认证")
	}
	if principal.VerificationID != nil && *principal.VerificationID != id {
		return nil, domain.New(domain.CodeOperationNotFound, "验证请求不存在")
	}
	verification, err := s.idswyft.GetVerification(ctx, id)
	if err != nil {
		return nil, domain.New(domain.CodeOperationNotFound, "验证请求不存在")
	}
	if verification.DeveloperID == nil || *verification.DeveloperID != principal.DeveloperID {
		return nil, domain.New(domain.CodeOperationNotFound, "验证请求不存在")
	}
	return verification, nil
}

func verificationPublic(v *gormmodel.VerificationRequest) map[string]any {
	status, step, total, final := externalStatus(v)
	payload := map[string]any{
		"success": true, "verification_id": v.ID.String(), "status": status,
		"current_step": step, "total_steps": total, "final_result": final,
		"verification_mode": stringValue(v.VerificationMode), "is_sandbox": v.IsSandbox != nil && *v.IsSandbox,
		"created_at": v.CreatedAt, "updated_at": v.UpdatedAt,
		"verification_type": stringValue(v.VerificationType), "source": v.Source,
	}
	if v.UserID != nil {
		payload["user_id"] = v.UserID.String()
	}
	if v.DocumentID != nil {
		payload["document_id"] = v.DocumentID.String()
	}
	if v.SelfieID != nil {
		payload["selfie_id"] = v.SelfieID.String()
	}
	return payload
}

// externalStatus 把 DB 5 值状态映射为 shared 12 值状态机（D1）。
func externalStatus(v *gormmodel.VerificationRequest) (string, int, int, any) {
	total := 5
	if v.VoiceChallenge != nil {
		total = 6
	}
	switch v.Status {
	case "verified":
		return "COMPLETE", total, total, "verified"
	case "failed":
		return "HARD_REJECTED", 0, total, "failed"
	case "manual_review":
		return "COMPLETE", total, total, "manual_review"
	case "processing":
		if v.DocumentID == nil {
			return "FRONT_PROCESSING", 1, total, nil
		}
		if v.BackOfIDUploaded == nil || !*v.BackOfIDUploaded {
			return "AWAITING_BACK", 2, total, nil
		}
		if v.LiveCaptureCompleted == nil || !*v.LiveCaptureCompleted {
			return "AWAITING_LIVE", 4, total, nil
		}
		if v.VoiceChallenge != nil && v.VoiceMatchScore == nil {
			return "AWAITING_VOICE", 6, total, nil
		}
		return "FACE_MATCHING", 5, total, nil
	default:
		return "AWAITING_FRONT", 1, total, nil
	}
}

func externalStatusName(v *gormmodel.VerificationRequest) string {
	status, _, _, _ := externalStatus(v)
	return status
}

func pathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		return uuid.Nil, domain.New(domain.CodeDocumentInvalid, name+" 必须是 UUID")
	}
	return id, nil
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(target); err != nil {
		return domain.New(domain.CodeDocumentInvalid, "请求体不是合法 JSON")
	}
	return nil
}

func newSessionToken(secret string) (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw)
	return token, hashToken(secret, token), nil
}

func boolPtr(v bool) *bool { return &v }

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Server) apiKeySecret() string {
	if s.cfg.Security.APIKeySecret != "" {
		return s.cfg.Security.APIKeySecret
	}
	return s.cfg.Security.MasterKey
}

func (s *Server) authenticateServiceToken(r *http.Request) bool {
	expected := s.cfg.Security.ServiceToken
	got := r.Header.Get("X-Service-Token")
	if expected == "" || got == "" {
		return false
	}
	macExpected := []byte(hashToken(expected, expected))
	macGot := []byte(hashToken(expected, got))
	return subtle.ConstantTimeCompare(macExpected, macGot) == 1
}
