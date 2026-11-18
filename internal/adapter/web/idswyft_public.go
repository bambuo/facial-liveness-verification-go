package web

import (
	"net/http"
	"net/mail"
	"time"

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
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "initialized": false})
}

// handleSetupInitialize 是首跑保护入口；P1 先返回明确的未实现信号。
func (s *Server) handleSetupInitialize(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "message": "setup initialize 尚未实现"})
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

// handleDeveloperRegister 是公开注册入口；P1 先返回明确的未实现信号。
func (s *Server) handleDeveloperRegister(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "message": "开发者注册尚未实现"})
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
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "verified": true,
		"developer": map[string]any{"id": developer.ID.String(), "email": developer.Email, "name": developer.Name, "company": developer.Company},
	})
}

// handleGitHubURL 只在配置 GitHub OAuth 时返回跳转地址。
func (s *Server) handleGitHubURL(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "message": "GitHub OAuth 未配置"})
}

// handleGitHubCallback 是公开 OAuth 回调入口；P1 先返回明确的未实现信号。
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "message": "GitHub OAuth 尚未实现"})
}
