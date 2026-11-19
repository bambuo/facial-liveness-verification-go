package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
)

var hexTokenPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// idswyftPrincipal 是 API Key、handoff 或 session token 解析后的请求主体。
type idswyftPrincipal struct {
	DeveloperID    uuid.UUID
	APIKeyID       uuid.UUID
	Sandbox        bool
	IsService      bool
	VerificationID *uuid.UUID
}

// authenticateAPIKeyOrHandoff 按 X-API-Key、X-Handoff-Token、X-Session-Token 顺序鉴权。
func (s *Server) authenticateAPIKeyOrHandoff(w http.ResponseWriter, r *http.Request) (*idswyftPrincipal, bool) {
	if s.idswyft == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "idswyft 鉴权存储未启用"))
		return nil, false
	}
	secret := s.cfg.Security.APIKeySecret
	if secret == "" {
		secret = s.cfg.Security.MasterKey
	}
	if secret == "" {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "API Key 密钥未配置"))
		return nil, false
	}
	switch {
	case r.Header.Get("X-API-Key") != "":
		return s.authenticateAPIKey(w, r, secret)
	case r.Header.Get("X-Handoff-Token") != "":
		return s.authenticateHandoff(w, r, secret)
	case r.Header.Get("X-Session-Token") != "":
		return s.authenticateSession(w, r, secret)
	case strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "):
		return s.authenticateDeveloperBearer(w, r)
	default:
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "缺少 X-API-Key、X-Handoff-Token 或 X-Session-Token"))
		return nil, false
	}
}

func (s *Server) authenticateAPIKey(w http.ResponseWriter, r *http.Request, secret string) (*idswyftPrincipal, bool) {
	hash := hashToken(secret, r.Header.Get("X-API-Key"))
	key, err := s.idswyft.FindAPIKeyByHash(r.Context(), hash)
	if err != nil || key == nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "API Key 无效"))
		return nil, false
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "API Key 已过期"))
		return nil, false
	}
	if key.DeveloperStatus == "suspended" {
		writeError(w, r, domain.New(domain.CodeTenantDisabled, "开发者账号已暂停"))
		return nil, false
	}
	_ = s.idswyft.TouchAPIKey(r.Context(), key.ID)
	return &idswyftPrincipal{DeveloperID: key.DeveloperID, APIKeyID: key.ID, Sandbox: key.IsSandbox, IsService: key.IsService}, true
}

func (s *Server) authenticateHandoff(w http.ResponseWriter, r *http.Request, secret string) (*idswyftPrincipal, bool) {
	token := r.Header.Get("X-Handoff-Token")
	if !hexTokenPattern.MatchString(token) {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "handoff token 格式无效"))
		return nil, false
	}
	session, err := s.idswyft.FindHandoffByToken(r.Context(), hashToken(secret, token))
	if err != nil || session == nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "handoff token 无效"))
		return nil, false
	}
	if session.ExpiresAt.Before(time.Now()) || session.Status == "expired" {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "handoff token 已失效"))
		return nil, false
	}
	if session.APIKey.DeveloperStatus == "suspended" {
		writeError(w, r, domain.New(domain.CodeTenantDisabled, "开发者账号已暂停"))
		return nil, false
	}
	_ = s.idswyft.TouchAPIKey(r.Context(), session.APIKey.ID)
	principal := &idswyftPrincipal{
		DeveloperID: session.APIKey.DeveloperID,
		APIKeyID:    session.APIKey.ID,
		Sandbox:     session.APIKey.IsSandbox,
		IsService:   session.APIKey.IsService,
	}
	if session.VerificationID != nil {
		if id, parseErr := uuid.Parse(*session.VerificationID); parseErr == nil {
			principal.VerificationID = &id
		}
	}
	return principal, true
}

func (s *Server) authenticateSession(w http.ResponseWriter, r *http.Request, secret string) (*idswyftPrincipal, bool) {
	token := r.Header.Get("X-Session-Token")
	if !hexTokenPattern.MatchString(token) {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "session token 格式无效"))
		return nil, false
	}
	session, err := s.idswyft.FindSessionByToken(r.Context(), hashToken(secret, token))
	if err != nil || session == nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "session token 无效"))
		return nil, false
	}
	if session.ExpiresAt == nil || session.ExpiresAt.Before(time.Now()) {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "session token 已过期"))
		return nil, false
	}
	principal := &idswyftPrincipal{
		DeveloperID:    session.DeveloperID,
		VerificationID: &session.VerificationID,
	}
	if session.APIKeyID != nil {
		principal.APIKeyID = *session.APIKeyID
	}
	return principal, true
}

// hashToken 对 API Key/handoff/session token 做 HMAC-SHA256。
func hashToken(secret, token string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

// developerClaims 是开发者门户会话 token 的载荷。
type developerClaims struct {
	DeveloperID string `json:"developer_id"`
	Email       string `json:"email"`
	Exp         int64  `json:"exp"`
}

// signDeveloperToken 签发开发者门户 Bearer token。
func (s *Server) signDeveloperToken(developerID uuid.UUID, email string) (string, error) {
	payload, err := json.Marshal(developerClaims{
		DeveloperID: developerID.String(), Email: email, Exp: time.Now().UTC().Add(12 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	return "dev_" + body + "." + hashToken(s.apiKeySecret(), body), nil
}

// authenticateDeveloperBearer 校验开发者门户 token。
func (s *Server) authenticateDeveloperBearer(w http.ResponseWriter, r *http.Request) (*idswyftPrincipal, bool) {
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !strings.HasPrefix(raw, "dev_") {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "开发者 token 无效"))
		return nil, false
	}
	parts := strings.Split(strings.TrimPrefix(raw, "dev_"), ".")
	if len(parts) != 2 || hmac.Equal([]byte(hashToken(s.apiKeySecret(), parts[0])), []byte(parts[1])) == false {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "开发者 token 无效"))
		return nil, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "开发者 token 无效"))
		return nil, false
	}
	var claims developerClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp < time.Now().Unix() {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "开发者 token 已过期"))
		return nil, false
	}
	developerID, err := uuid.Parse(claims.DeveloperID)
	if err != nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "开发者 token 无效"))
		return nil, false
	}
	return &idswyftPrincipal{DeveloperID: developerID}, true
}
