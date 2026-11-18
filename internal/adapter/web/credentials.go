package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

// handleCredentialIssue 实现 GET /api/v2/verify/:id/credential 与 /api/v2/:id/credential。
func (s *Server) handleCredentialIssue(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	if s.vc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Verifiable Credentials are not available on this instance"})
		return
	}
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	if err != nil || developer.VcEnabled == nil || !*developer.VcEnabled {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "Verifiable Credentials are not enabled for your account. Enable them in Settings → Integrations."})
		return
	}
	verification, err := s.credentialVerification(w, r, principal)
	if err != nil {
		return
	}
	if existing, _ := s.idswyft.GetActiveCredential(r.Context(), principal.DeveloperID, verification.ID); existing != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "already_issued": true, "jti": existing.CredentialJti,
			"issued_at": existing.IssuedAt, "expires_at": existing.ExpiresAt,
			"message": "Credential was already issued. Use the JTI to check status.",
		})
		return
	}
	credential, err := s.issueCredential(r, verification, principal.DeveloperID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "credential": credential.JWT, "jti": credential.JTI, "expires_at": credential.ExpiresAt,
	})
}

// handleCredentialSend 实现 POST /api/v2/verify/:id/credential/send 与 /api/v2/:id/credential/send。
func (s *Server) handleCredentialSend(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "A valid email address is required"})
		return
	}
	if s.vc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Verifiable Credentials are not available on this instance"})
		return
	}
	developer, err := s.idswyft.GetDeveloperByID(r.Context(), principal.DeveloperID)
	if err != nil || developer.VcEnabled == nil || !*developer.VcEnabled {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "Verifiable Credentials are not enabled for your account."})
		return
	}
	verification, err := s.credentialVerification(w, r, principal)
	if err != nil {
		return
	}
	var jwt, jti string
	var expires time.Time
	if existing, _ := s.idswyft.GetActiveCredential(r.Context(), principal.DeveloperID, verification.ID); existing != nil && existing.CredentialJWT != nil {
		jwt, jti, expires = *existing.CredentialJWT, existing.CredentialJti, existing.ExpiresAt
	} else {
		credential, issueErr := s.issueCredential(r, verification, principal.DeveloperID)
		if issueErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": issueErr.Error()})
			return
		}
		jwt, jti, expires = credential.JWT, credential.JTI, credential.ExpiresAt
	}
	emailSent := false
	if s.mail != nil && s.cfg.SMTP.Enabled {
		link := "http://" + r.Host + "/verify-credential?jwt=" + jwt
		if err := s.mail.Send(req.Email, "Your verifiable credential", "Verify your credential: "+link); err == nil {
			emailSent = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "jti": jti, "expires_at": expires, "email_sent": emailSent})
}

// handleCredentialRevoke 实现 POST /api/v2/credentials/:jti/revoke 与 /api/v2/verify/credentials/:jti/revoke。
func (s *Server) handleCredentialRevoke(w http.ResponseWriter, r *http.Request) {
	principal := principalFrom(r)
	var req struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSONBody(w, r, &req)
	affected, err := s.idswyft.RevokeCredential(r.Context(), principal.DeveloperID, r.PathValue("jti"), req.Reason)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if affected == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Credential not found or already revoked"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Credential revoked"})
}

// handleCredentialStatus 实现公开 GET /api/v2/credentials/:jti/status。
func (s *Server) handleCredentialStatus(w http.ResponseWriter, r *http.Request) {
	credential, err := s.idswyft.GetCredentialByJTI(r.Context(), r.PathValue("jti"))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"active": false, "reason": "not_found"})
		return
	}
	if credential.RevokedAt != nil {
		writeJSON(w, http.StatusOK, map[string]any{"active": false, "reason": "revoked"})
		return
	}
	if credential.ExpiresAt.Before(time.Now()) {
		writeJSON(w, http.StatusOK, map[string]any{"active": false, "reason": "expired"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"active": true})
}

func (s *Server) credentialVerification(w http.ResponseWriter, r *http.Request, principal *idswyftPrincipal) (*gormmodel.VerificationRequest, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "verification id must be a UUID"})
		return nil, err
	}
	verification, err := s.loadOwnedVerification(r.Context(), id, principal)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Verification not found"})
		return nil, err
	}
	if verification.Status != "verified" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Credential can only be issued for verified identities"})
		return nil, domain.New(domain.CodeDocumentInvalid, "not verified")
	}
	return verification, nil
}

func (s *Server) issueCredential(r *http.Request, verification *gormmodel.VerificationRequest, developerID uuid.UUID) (vcCredential, error) {
	identity, err := s.identityData(r.Context(), verification)
	if err != nil {
		return vcCredential{}, err
	}
	subject := map[string]any{
		"name": identityString(identity, "full_name"), "dateOfBirth": identityString(identity, "date_of_birth"),
		"documentType": identityString(identity, "document_type"), "verifiedAt": time.Now().UTC().Format(time.RFC3339),
	}
	credential, err := s.vc.Issue(subject)
	if err != nil {
		return vcCredential{}, err
	}
	record := &gormmodel.VerifiableCredential{
		ID: uuid.New(), VerificationRequestID: verification.ID, DeveloperID: developerID,
		CredentialJti: credential.JTI, CredentialJWT: &credential.JWT,
		IssuedAt: credential.IssuedAt, ExpiresAt: credential.ExpiresAt, CreatedAt: time.Now().UTC(),
	}
	if err := s.idswyft.CreateCredential(r.Context(), record); err != nil {
		return vcCredential{}, err
	}
	return vcCredential{JWT: credential.JWT, JTI: credential.JTI, ExpiresAt: credential.ExpiresAt}, nil
}

type vcCredential struct {
	JWT       string
	JTI       string
	ExpiresAt time.Time
}

func identityString(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return value
}
