package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// handleOTPCompleteRegistration 完成 OTP 注册：校验邮箱验证码并返回账号状态。
func (s *Server) handleOTPCompleteRegistration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !otpPattern.MatchString(req.Code) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "code 必须是 6 位数字"})
		return
	}
	result, err := s.idswyft.VerifyDeveloperOTP(r.Context(), req.Email, hashToken(s.apiKeySecret(), req.Code), 3)
	if err != nil || result.Status != "valid" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "message": "验证码无效"})
		return
	}
	developer, err := s.idswyft.FindDeveloperByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "verified": true})
		return
	}
	token, err := s.signDeveloperToken(developer.ID, developer.Email)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "verified": true, "token": token, "developer": map[string]any{
		"id": developer.ID, "email": developer.Email, "name": developer.Name,
	}})
}

// handleTOTPSetup 生成 TOTP secret 与 otpauth URL。
func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		writeError(w, r, err)
		return
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)
	label := url.QueryEscape("idswyft:" + principalFrom(r).DeveloperID.String())
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "secret": secret,
		"otpauth_url": fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=idswyft", label, secret),
	})
}

// handleTOTPVerify 校验 TOTP 6 位码。
func (s *Server) handleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if !validateTOTP(req.Secret, req.Code, time.Now().UTC()) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "invalid TOTP code"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "verified": true})
}

func validateTOTP(secret, code string, now time.Time) bool {
	if secret == "" || !otpPattern.MatchString(code) {
		return false
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return false
	}
	counter := uint64(now.Unix() / 30)
	for offset := int64(-1); offset <= 1; offset++ {
		if totpCode(key, uint64(int64(counter)+offset)) == code {
			return true
		}
	}
	return false
}

func totpCode(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000)
}

// handleCredentialV2Alias 处理 /api/v2/:id/credential 与 /send（避免 ServeMux 冲突）。
func (s *Server) handleCredentialV2Alias(w http.ResponseWriter, r *http.Request) bool {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/")
	segments := strings.Split(path, "/")
	if len(segments) < 2 || segments[0] == "verify" {
		return false
	}
	id, err := uuid.Parse(segments[0])
	if err != nil {
		return false
	}
	switch {
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "credential":
		r.SetPathValue("id", id.String())
		s.withAPIAuth(s.handleCredentialIssue)(w, r)
		return true
	case r.Method == http.MethodPost && len(segments) == 3 && segments[1] == "credential" && segments[2] == "send":
		r.SetPathValue("id", id.String())
		s.withAPIAuth(s.handleCredentialSend)(w, r)
		return true
	}
	return false
}
