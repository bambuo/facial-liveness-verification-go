package web

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/domain"
)

// handleAPIHealth 返回公开健康摘要。
func (s *Server) handleAPIHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "idswyft"})
}

// handleAPIHealthDetailed 返回组件级健康状态，不输出敏感配置。
func (s *Server) handleAPIHealthDetailed(w http.ResponseWriter, r *http.Request) {
	dbStatus, redisStatus := "ok", "ok"
	if s.dbPing != nil {
		if err := s.dbPing(r.Context()); err != nil {
			dbStatus = "error"
		}
	}
	if s.redisPing != nil {
		if err := s.redisPing(r.Context()); err != nil {
			redisStatus = "error"
		}
	}
	status := "ok"
	if dbStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     status,
		"components": map[string]string{"database": dbStatus, "redis": redisStatus},
	})
}

// handleAPIHealthReady 以依赖可用性作为就绪信号。
func (s *Server) handleAPIHealthReady(w http.ResponseWriter, r *http.Request) {
	if s.dbPing != nil {
		if err := s.dbPing(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready", "component": "database"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

// handleAPIHealthLive 返回进程存活信号。
func (s *Server) handleAPIHealthLive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "live"})
}

// handleAPIStatus 返回公开服务状态。
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "operational", "version": s.buildVersion})
}

// handleSetupStatus 返回是否已完成首跑初始化。
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	initialized := false
	if s.idswyft != nil {
		if count, err := s.idswyft.AdminDeveloperCount(r.Context()); err == nil && count > 0 {
			initialized = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "initialized": initialized, "needs_setup": !initialized})
}

// handleSetupInitialize 创建首个开发者与 API Key，并返回开发者会话。
func (s *Server) handleSetupInitialize(w http.ResponseWriter, r *http.Request) {
	if s.idswyft == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "setup storage unavailable"))
		return
	}
	var req struct {
		Name    string  `json:"name"`
		Email   string  `json:"email"`
		Company *string `json:"company"`
		Locale  string  `json:"locale"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(req.Email); err != nil || len(req.Name) < 2 {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "name 或 email 无效"))
		return
	}
	if count, err := s.idswyft.AdminDeveloperCount(r.Context()); err != nil {
		writeError(w, r, err)
		return
	} else if count > 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "message": "Setup already completed"})
		return
	}
	now := time.Now().UTC()
	verified := true
	developer := &gormmodel.Developer{
		ID: uuid.New(), Email: req.Email, Name: req.Name, Company: req.Company,
		IsVerified: &verified, Status: "active", Locale: firstNonEmpty(req.Locale, "en"),
		CreatedAt: &now, UpdatedAt: &now,
	}
	if err := s.idswyft.CreateDeveloper(r.Context(), developer); err != nil {
		writeError(w, r, err)
		return
	}
	raw, _, err := newSessionToken(s.apiKeySecret())
	if err != nil {
		writeError(w, r, err)
		return
	}
	keyValue := "ik_" + raw
	key := &gormmodel.APIKey{
		ID: uuid.New(), DeveloperID: &developer.ID, Name: "Default key",
		KeyHash: hashToken(s.apiKeySecret(), keyValue), KeyPrefix: keyValue[:8],
		IsActive: boolPtr(true), IsSandbox: boolPtr(false), CreatedAt: &now,
	}
	if err := s.idswyft.CreateDeveloperAPIKey(r.Context(), key); err != nil {
		writeError(w, r, err)
		return
	}
	token, err := s.signDeveloperToken(developer.ID, developer.Email)
	if err != nil {
		writeError(w, r, err)
		return
	}
	setAuthCookie(w, r, developerSessionCookie, token, 12*60*60)
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true, "token": token,
		"developer": map[string]any{"id": developer.ID, "email": developer.Email, "name": developer.Name, "company": developer.Company},
		"api_key":   map[string]any{"key": keyValue, "id": key.ID, "name": key.Name, "is_sandbox": false, "created_at": now},
	})
}

// handleDIDDocument 返回最小 did:web 文档。
func (s *Server) handleDIDDocument(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	writeJSON(w, http.StatusOK, map[string]any{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id":       "did:web:" + host,
	})
}

// handlePageConfigSlug 返回公开页面配置占位；真实配置由页面模块接管。
func (s *Server) handlePageConfigSlug(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "slug": r.PathValue("slug"), "config": map[string]any{}})
}

// handleDeveloperRegister 创建未验证开发者并发送邮箱 OTP。
func (s *Server) handleDeveloperRegister(w http.ResponseWriter, r *http.Request) {
	if s.idswyft == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "注册存储未启用"))
		return
	}
	var req struct {
		Email      string  `json:"email"`
		Name       string  `json:"name"`
		Company    *string `json:"company"`
		WebhookURL *string `json:"webhook_url"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	if _, err := mail.ParseAddress(req.Email); err != nil || len(req.Name) < 2 || len(req.Name) > 100 {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "email 或 name 无效"))
		return
	}
	if _, err := s.idswyft.FindDeveloperByEmail(r.Context(), req.Email); err == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Developer with this email already exists"})
		return
	}
	now := time.Now().UTC()
	verified := false
	developer := &gormmodel.Developer{
		ID: uuid.New(), Email: req.Email, Name: req.Name, Company: req.Company,
		WebhookURL: req.WebhookURL, IsVerified: &verified, Status: "active", Locale: "en",
		CreatedAt: &now, UpdatedAt: &now,
	}
	if err := s.idswyft.CreateDeveloper(r.Context(), developer); err != nil {
		writeError(w, r, err)
		return
	}
	code, _ := randomDigits(6)
	_ = s.idswyft.CreateDeveloperOTP(r.Context(), req.Email, hashToken(s.apiKeySecret(), code), 10*time.Minute)
	if s.cfg.SMTP.Enabled && s.mail != nil {
		_ = s.mail.Send(req.Email, "Verify your idswyft account", "Verification code: "+code)
	}
	response := map[string]any{
		"developer": map[string]any{
			"id": developer.ID, "email": developer.Email, "name": developer.Name,
			"company": developer.Company, "is_verified": false, "created_at": developer.CreatedAt,
		},
		"message": "Verification email sent. Complete email verification to activate your account.",
	}
	if !s.cfg.SMTP.Enabled && s.cfg.Env == "development" {
		response["code"] = code
	}
	writeJSON(w, http.StatusCreated, response)
}

// handleAdminEscalate 按参考实现固定返回 410。
func (s *Server) handleAdminEscalate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusGone, map[string]any{"success": false, "message": "admin escalate 已下线"})
}

// handleOTPSend 发送邮箱 OTP；未配置 SMTP 时开发环境回显验证码。
func (s *Server) handleOTPSend(w http.ResponseWriter, r *http.Request) {
	if s.idswyft == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OTP 存储未启用"))
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"message": "If this email is valid, a verification code has been sent."})
		return
	}
	allowed, err := s.idswyft.CheckDeveloperOTPSend(r.Context(), req.Email, 5, 3600)
	if err != nil || !allowed {
		writeJSON(w, http.StatusOK, map[string]any{"message": "If this email is valid, a verification code has been sent."})
		return
	}
	code, err := randomDigits(6)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := s.idswyft.CreateDeveloperOTP(r.Context(), req.Email, hashToken(s.apiKeySecret(), code), 10*time.Minute); err != nil {
		writeError(w, r, err)
		return
	}
	if s.cfg.SMTP.Enabled && s.mail != nil {
		_ = s.mail.Send(req.Email, "Your idswyft verification code", "Verification code: "+code)
	}
	resp := map[string]any{"message": "If this email is valid, a verification code has been sent."}
	if !s.cfg.SMTP.Enabled && s.cfg.Env == "development" {
		resp["code"], resp["self_hosted"] = code, true
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleOTPVerify 校验邮箱 OTP；成功后返回开发者摘要，登录令牌由后续 auth 模块接管。
func (s *Server) handleOTPVerify(w http.ResponseWriter, r *http.Request) {
	if s.idswyft == nil {
		writeError(w, r, domain.New(domain.CodeServiceUnavailable, "OTP 存储未启用"))
		return
	}
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil || !otpPattern.MatchString(req.Code) {
		writeError(w, r, domain.New(domain.CodeDocumentInvalid, "email 或 code 格式无效"))
		return
	}
	result, err := s.idswyft.VerifyDeveloperOTP(r.Context(), req.Email, hashToken(s.apiKeySecret(), req.Code), 3)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if result.Status != "valid" {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "验证码无效"))
		return
	}
	developer, err := s.idswyft.FindDeveloperByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "verified": true})
		return
	}
	var token string
	if strings.Contains(r.URL.Path, "/reviewer/") {
		token, err = s.signAdminToken(developer.Email)
		if err == nil {
			setAuthCookie(w, r, adminSessionCookie, token, 8*60*60)
		}
	} else {
		token, err = s.signDeveloperToken(developer.ID, developer.Email)
		if err == nil {
			setAuthCookie(w, r, developerSessionCookie, token, 12*60*60)
		}
	}
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verified": true, "token": token,
		"developer": map[string]any{"id": developer.ID.String(), "email": developer.Email, "name": developer.Name, "company": developer.Company},
	})
}

// handleGitHubURL 只在配置 GitHub OAuth 时返回跳转地址。
func (s *Server) handleGitHubURL(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"configured": false})
}

// handleGitHubCallback 是公开 OAuth 回调入口；P1 先返回明确的未实现信号。
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "message": "GitHub OAuth 尚未实现"})
}
