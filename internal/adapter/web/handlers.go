package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"facial-liveness-verification-go/internal/app"
	"facial-liveness-verification-go/internal/audio"
	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/deepfake"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
	"facial-liveness-verification-go/internal/vc"
)

//go:embed static/face-test.html
var staticFS embed.FS

// Server 组装 HTTP 依赖。
type Server struct {
	cfg           *config.Config
	auth          *app.Authenticator
	challenges    *app.ChallengeService
	actions       *app.ActionService
	enroll        *app.EnrollService
	verify        *app.VerifyService
	subjects      *app.SubjectService
	support       *app.VerificationSupport
	liveness      *app.LivenessService
	issuer        port.TokenIssuer
	ocr           OCRService
	voice         *audio.VoiceEngine
	face          port.FaceEngine
	idswyft       idswyftStore
	storage       ObjectStore
	webhooks      WebhookDispatcher
	sms           SMSSender
	mail          Mailer
	encryptor     DataEncryptor
	vc            *vc.Issuer
	thresholdMu   sync.RWMutex
	thresholds    map[string]map[string]float64
	deepfake      *deepfake.Engine
	engineVersion string
	buildVersion  string
	buildCommit   string
	modelHash     string
	configHash    string
	dbPing        func(context.Context) error
	redisPing     func(context.Context) error
	limiter       *ipLimiter
}

// Deps 是 HTTP 服务的构造依赖。
type Deps struct {
	Cfg           *config.Config
	Auth          *app.Authenticator
	Challenges    *app.ChallengeService
	Actions       *app.ActionService
	Enroll        *app.EnrollService
	Verify        *app.VerifyService
	Subjects      *app.SubjectService
	Support       *app.VerificationSupport
	Liveness      *app.LivenessService
	Issuer        port.TokenIssuer
	OCR           OCRService
	Voice         *audio.VoiceEngine
	Face          port.FaceEngine
	IdswyftStore  idswyftStore
	Storage       ObjectStore
	Webhooks      WebhookDispatcher
	SMS           SMSSender
	Mail          Mailer
	Encryptor     DataEncryptor
	VC            *vc.Issuer
	Deepfake      *deepfake.Engine
	EngineVersion string
	BuildVersion  string
	BuildCommit   string
	ModelHash     string
	DBPing        func(context.Context) error
	RedisPing     func(context.Context) error
}

// NewServer 构造 HTTP 服务。
func NewServer(d Deps) *Server {
	return &Server{
		cfg: d.Cfg, auth: d.Auth, challenges: d.Challenges, actions: d.Actions,
		enroll: d.Enroll, verify: d.Verify, subjects: d.Subjects, support: d.Support,
		liveness: d.Liveness, issuer: d.Issuer, ocr: d.OCR, voice: d.Voice, face: d.Face,
		idswyft: d.IdswyftStore, storage: d.Storage, webhooks: d.Webhooks, sms: d.SMS, mail: d.Mail, encryptor: d.Encryptor, engineVersion: d.EngineVersion,
		buildVersion: d.BuildVersion, buildCommit: d.BuildCommit,
		modelHash: d.ModelHash, configHash: d.Cfg.Fingerprint(),
		dbPing: d.DBPing, redisPing: d.RedisPing,
		limiter: newIPLimiter(d.Cfg.RateLimit.PermitPerMinute),
	}
}

// Handler 注册路由并套上中间件。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", s.healthLive)
	mux.HandleFunc("GET /health/ready", s.healthReady)
	mux.HandleFunc("GET /version", s.handleVersion)
	mux.HandleFunc("POST /v1/challenges", s.handleChallenge)
	mux.HandleFunc("POST /v1/action-check", s.handleActionCheck)
	mux.HandleFunc("POST /v1/enrollments", s.handleEnroll)
	mux.HandleFunc("POST /v1/verifications", s.handleVerify)
	mux.HandleFunc("GET /v1/subjects/{subject_id}", s.handleSubjectGet)
	mux.HandleFunc("DELETE /v1/subjects/{subject_id}", s.handleSubjectDelete)
	mux.HandleFunc("POST /extract/ocr", s.handleExtractOCR)
	mux.HandleFunc("POST /extract/front", s.handleExtractFront)
	mux.HandleFunc("POST /extract/back", s.handleExtractBack)
	mux.HandleFunc("POST /extract/live", s.handleExtractLive)
	mux.HandleFunc("POST /extract/voice-enroll", s.handleVoiceEnroll)
	mux.HandleFunc("POST /extract/voice-verify", s.handleVoiceVerify)

	mux.HandleFunc("POST /api/v2/compliance/rulesets", s.withAPIAuth(s.handleRulesetCreate))
	mux.HandleFunc("GET /api/v2/compliance/rulesets", s.withAPIAuth(s.handleRulesetList))
	mux.HandleFunc("GET /api/v2/compliance/rulesets/{id}", s.withAPIAuth(s.handleRulesetGet))
	mux.HandleFunc("PUT /api/v2/compliance/rulesets/{id}", s.withAPIAuth(s.handleRulesetUpdate))
	mux.HandleFunc("DELETE /api/v2/compliance/rulesets/{id}", s.withAPIAuth(s.handleRulesetDelete))
	mux.HandleFunc("POST /api/v2/compliance/rulesets/{id}/rules", s.withAPIAuth(s.handleRuleCreate))
	mux.HandleFunc("PUT /api/v2/compliance/rules/{id}", s.withAPIAuth(s.handleRuleUpdate))
	mux.HandleFunc("DELETE /api/v2/compliance/rules/{id}", s.withAPIAuth(s.handleRuleDelete))
	mux.HandleFunc("POST /api/v2/compliance/evaluate", s.withAPIAuth(s.handleComplianceEvaluate))

	mux.HandleFunc("GET /api/developer/webhooks", s.withAPIAuth(s.handleWebhookList))
	mux.HandleFunc("POST /api/developer/webhooks", s.withAPIAuth(s.handleWebhookRegister))
	mux.HandleFunc("GET /api/developer/webhooks/{webhookId}/secret", s.withAPIAuth(s.handleDeveloperWebhookSecret))
	mux.HandleFunc("GET /api/developer/webhooks/{webhookId}/deliveries", s.withAPIAuth(s.handleWebhookDeliveries))
	mux.HandleFunc("POST /api/developer/webhooks/{webhookId}/test", s.withAPIAuth(s.handleWebhookTest))
	mux.HandleFunc("POST /api/developer/webhooks/{webhookId}/deliveries/{deliveryId}/resend", s.withAPIAuth(s.handleDeveloperWebhookResend))
	mux.HandleFunc("DELETE /api/developer/webhooks/{webhookId}", s.withAPIAuth(s.handleWebhookDelete))
	mux.HandleFunc("GET /api/system/version", s.withAPIAuth(s.handleSystemVersion))

	mux.HandleFunc("GET /api/admin/thresholds", s.withAdminAuth(s.handleThresholdsGet))
	mux.HandleFunc("PUT /api/admin/thresholds", s.withAdminAuth(s.handleThresholdsUpdate))
	mux.HandleFunc("POST /api/admin/thresholds/preview", s.withAdminAuth(s.handleThresholdsPreview))
	mux.HandleFunc("POST /api/admin/thresholds/reset", s.withAdminAuth(s.handleThresholdsReset))

	mux.HandleFunc("GET /api/admin/dashboard", s.withAdminAuth(s.handleAdminDashboard))
	mux.HandleFunc("GET /api/admin/verifications", s.withAdminAuth(s.handleAdminVerifications))
	mux.HandleFunc("GET /api/admin/verification/{id}", s.withAdminAuth(s.handleAdminVerification))
	mux.HandleFunc("GET /api/admin/verification/{id}/duplicates", s.withAdminAuth(s.handleAdminDuplicates))
	mux.HandleFunc("PUT /api/admin/verification/{id}/review", s.withAdminAuth(s.handleAdminReview))
	mux.HandleFunc("GET /api/admin/developers", s.withAdminAuth(s.handleAdminDevelopers))
	mux.HandleFunc("GET /api/admin/analytics", s.withAdminAuth(s.handleAdminAnalytics))
	mux.HandleFunc("GET /api/admin/health", s.withAdminAuth(s.handleAdminHealth))
	mux.HandleFunc("GET /api/admin/analytics/funnel", s.withAdminAuth(s.handleAdminFunnel))
	mux.HandleFunc("GET /api/admin/analytics/rejections", s.withAdminAuth(s.handleAdminRejections))
	mux.HandleFunc("GET /api/admin/analytics/fraud-patterns", s.withAdminAuth(s.handleAdminFraudPatterns))
	mux.HandleFunc("GET /api/admin/analytics/risk-distribution", s.withAdminAuth(s.handleAdminRiskDistribution))
	mux.HandleFunc("GET /api/admin/provider-metrics", s.withAdminAuth(s.handleAdminProviderMetrics))
	mux.HandleFunc("GET /api/admin/audit/export", s.withAdminAuth(s.handleAdminAuditExport))
	mux.HandleFunc("DELETE /api/admin/user/{userId}/data", s.withAdminAuth(s.handleAdminDeleteUserData))

	mux.HandleFunc("POST /api/developer/reviewers/invite", s.withAPIAuth(s.handleReviewerInvite))
	mux.HandleFunc("GET /api/developer/reviewers", s.withAPIAuth(s.handleReviewerList))
	mux.HandleFunc("DELETE /api/developer/reviewers/{id}", s.withAPIAuth(s.handleReviewerDelete))

	mux.HandleFunc("GET /api/v2/verify/page-config", s.handlePageConfig)
	mux.HandleFunc("GET /api/v2/verify/session-info", s.handleSessionInfo)
	mux.HandleFunc("GET /api/health/api-key-diagnostic", s.handleAPIKeyDiagnostic)
	mux.HandleFunc("GET /api/docs/markdown", s.handleDocsMarkdown)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	mux.HandleFunc("GET /api/auth/csrf-token", s.handleCSRFToken)

	mux.HandleFunc("GET /api/developer/settings/llm", s.withAPIAuth(s.handleSettingsLLMGet))
	mux.HandleFunc("PUT /api/developer/settings/llm", s.withAPIAuth(s.handleSettingsLLMPut))
	mux.HandleFunc("GET /api/developer/settings/sms", s.withAPIAuth(s.handleSettingsSMSGet))
	mux.HandleFunc("PUT /api/developer/settings/sms", s.withAPIAuth(s.handleSettingsSMSPut))
	mux.HandleFunc("GET /api/developer/settings/branding", s.withAPIAuth(s.handleSettingsBrandingGet))
	mux.HandleFunc("PUT /api/developer/settings/branding", s.withAPIAuth(s.handleSettingsBrandingPut))
	mux.HandleFunc("GET /api/developer/settings/page-builder", s.withAPIAuth(s.handleSettingsPageBuilderGet))
	mux.HandleFunc("PUT /api/developer/settings/page-builder", s.withAPIAuth(s.handleSettingsPageBuilderPut))
	mux.HandleFunc("PUT /api/developer/settings/page-builder/slug", s.withAPIAuth(s.handleSettingsPageBuilderSlug))
	mux.HandleFunc("GET /api/developer/settings/aml", s.withAPIAuth(s.handleSettingsAMLGet))
	mux.HandleFunc("PUT /api/developer/settings/aml", s.withAPIAuth(s.handleSettingsAMLPut))
	mux.HandleFunc("GET /api/developer/settings/voice-auth", s.withAPIAuth(s.handleSettingsVoiceGet))
	mux.HandleFunc("PUT /api/developer/settings/voice-auth", s.withAPIAuth(s.handleSettingsVoicePut))
	mux.HandleFunc("GET /api/developer/settings/vc", s.withAPIAuth(s.handleSettingsVCGet))
	mux.HandleFunc("PUT /api/developer/settings/vc", s.withAPIAuth(s.handleSettingsVCPut))
	mux.HandleFunc("GET /api/developer/settings/duplicate-detection", s.withAPIAuth(s.handleSettingsDuplicateGet))
	mux.HandleFunc("PUT /api/developer/settings/duplicate-detection", s.withAPIAuth(s.handleSettingsDuplicatePut))
	mux.HandleFunc("GET /api/developer/settings/vault", s.withAPIAuth(s.handleSettingsVaultGet))
	mux.HandleFunc("PUT /api/developer/settings/vault", s.withAPIAuth(s.handleSettingsVaultPut))

	mux.HandleFunc("GET /api/developer/stats", s.withAPIAuth(s.handleDeveloperStats))
	mux.HandleFunc("GET /api/developer/activity", s.withAPIAuth(s.handleDeveloperActivity))
	mux.HandleFunc("GET /api/developer/verifications/{verificationId}", s.withAPIAuth(s.handleDeveloperVerification))
	mux.HandleFunc("GET /api/developer/analytics", s.withAPIAuth(s.handleDeveloperAnalytics))

	mux.HandleFunc("GET /api/developer/profile", s.withAPIAuth(s.handleDeveloperProfileGet))
	mux.HandleFunc("PUT /api/developer/profile", s.withAPIAuth(s.handleDeveloperProfileUpdate))
	mux.HandleFunc("POST /api/developer/avatar", s.withAPIAuth(s.handleDeveloperAvatar("avatar")))
	mux.HandleFunc("POST /api/developer/branding/logo", s.withAPIAuth(s.handleDeveloperAvatar("branding")))
	mux.HandleFunc("DELETE /api/developer/account", s.withAPIAuth(s.handleDeveloperAccountDelete))

	mux.HandleFunc("POST /api/developer/api-key", s.withAPIAuth(s.handleDeveloperKeyCreate))
	mux.HandleFunc("GET /api/developer/api-keys", s.withAPIAuth(s.handleDeveloperKeyList))
	mux.HandleFunc("DELETE /api/developer/api-key/{keyId}", s.withAPIAuth(s.handleDeveloperKeyRevoke))
	mux.HandleFunc("POST /api/developer/api-key/{id}/rotate", s.withAPIAuth(s.handleDeveloperKeyRotate))

	mux.HandleFunc("POST /api/vaas/users", s.handleVAASCreateUser)
	mux.HandleFunc("POST /api/vaas/verify", s.handleVAASVerify)
	mux.HandleFunc("GET /api/vaas/verify/{verification_id}/status", s.handleVAASStatus)
	mux.HandleFunc("GET /api/vaas/developers", s.handleVAASDevelopers)
	mux.HandleFunc("GET /api/vaas/developers/{id}", s.handleVAASDeveloper)
	mux.HandleFunc("POST /api/vaas/developers/{id}/suspend", s.handleVAASDeveloperStatus("suspended"))
	mux.HandleFunc("POST /api/vaas/developers/{id}/unsuspend", s.handleVAASDeveloperStatus("active"))
	mux.HandleFunc("GET /api/vaas/health", s.handleVAASHealth)

	mux.HandleFunc("POST /api/v2/monitoring/schedules", s.withAPIAuth(s.handleScheduleCreate))
	mux.HandleFunc("GET /api/v2/monitoring/schedules", s.withAPIAuth(s.handleScheduleList))
	mux.HandleFunc("GET /api/v2/monitoring/schedules/{schedule_id}", s.withAPIAuth(s.handleScheduleGet))
	mux.HandleFunc("DELETE /api/v2/monitoring/schedules/{schedule_id}", s.withAPIAuth(s.handleScheduleCancel))
	mux.HandleFunc("GET /api/v2/monitoring/expiring-documents", s.withAPIAuth(s.handleExpiringDocuments))

	mux.HandleFunc("POST /api/v2/verify/{verification_id}/address-document", s.withAPIAuth(s.handleAddressDocument))
	mux.HandleFunc("GET /api/v2/verify/{verification_id}/address-status", s.withAPIAuth(s.handleAddressStatus))

	mux.HandleFunc("POST /api/webhooks/register", s.withAPIAuth(s.handleWebhookRegister))
	mux.HandleFunc("GET /api/webhooks", s.withAPIAuth(s.handleWebhookList))
	mux.HandleFunc("PUT /api/webhooks/{webhookId}", s.withAPIAuth(s.handleWebhookUpdate))
	mux.HandleFunc("DELETE /api/webhooks/{webhookId}", s.withAPIAuth(s.handleWebhookDelete))
	mux.HandleFunc("GET /api/webhooks/{webhookId}/deliveries", s.withAPIAuth(s.handleWebhookDeliveries))
	mux.HandleFunc("POST /api/webhooks/{webhookId}/test", s.withAPIAuth(s.handleWebhookTest))

	mux.HandleFunc("GET /api/v2/verify/{id}/credential", s.withAPIAuth(s.handleCredentialIssue))
	mux.HandleFunc("POST /api/v2/verify/{id}/credential/send", s.withAPIAuth(s.handleCredentialSend))
	mux.HandleFunc("POST /api/v2/verify/credentials/{jti}/revoke", s.withAPIAuth(s.handleCredentialRevoke))
	mux.HandleFunc("GET /api/v2/verify/credentials/{jti}/status", s.handleCredentialStatus)
	mux.HandleFunc("POST /api/v2/credentials/{jti}/revoke", s.withAPIAuth(s.handleCredentialRevoke))
	mux.HandleFunc("GET /api/v2/credentials/{jti}/status", s.handleCredentialStatus)

	mux.HandleFunc("GET /api/v2/vault/share/{shareToken}", s.handleVaultShareGet)
	mux.HandleFunc("POST /api/v2/vault/store", s.withAPIAuth(s.handleVaultStore))
	mux.HandleFunc("GET /api/v2/vault", s.withAPIAuth(s.handleVaultList))
	mux.HandleFunc("GET /api/v2/vault/{token}", s.withAPIAuth(s.handleVaultGet))
	mux.HandleFunc("GET /api/v2/vault/{token}/attributes/{attr}", s.withAPIAuth(s.handleVaultAttribute))
	mux.HandleFunc("DELETE /api/v2/vault/{token}", s.withAPIAuth(s.handleVaultDelete))
	mux.HandleFunc("POST /api/v2/vault/{token}/share", s.withAPIAuth(s.handleVaultShareCreate))

	mux.HandleFunc("POST /api/v2/batch/upload", s.withAPIAuth(s.handleBatchUpload))
	mux.HandleFunc("GET /api/v2/batch", s.withAPIAuth(s.handleBatchList))
	mux.HandleFunc("GET /api/v2/batch/{id}/status", s.withAPIAuth(s.handleBatchStatus))
	mux.HandleFunc("GET /api/v2/batch/{id}/results", s.withAPIAuth(s.handleBatchResults))
	mux.HandleFunc("POST /api/v2/batch/{id}/cancel", s.withAPIAuth(s.handleBatchCancel))

	mux.HandleFunc("POST /api/verify/handoff/create", s.handleHandoffCreate)
	mux.HandleFunc("GET /api/verify/handoff/{token}/session", s.handleHandoffSession)
	mux.HandleFunc("PATCH /api/verify/handoff/{token}/link", s.handleHandoffLink)
	mux.HandleFunc("PATCH /api/verify/handoff/{token}/complete", s.handleHandoffComplete)
	mux.HandleFunc("GET /api/verify/handoff/{token}/status", s.handleHandoffStatus)

	mux.HandleFunc("POST /api/v2/verify/initialize", s.withAPIAuth(s.handleVerifyInitialize))
	mux.HandleFunc("POST /api/v2/verify/re-verify", s.withAPIAuth(s.handleVerifyReverify))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/front-document", s.withAPIAuth(s.handleFrontDocument))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/back-document", s.withAPIAuth(s.handleBackDocument))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/cross-validation", s.withAPIAuth(s.handleCrossValidation))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/live-capture", s.withAPIAuth(s.handleLiveCapture))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/voice-challenge", s.withAPIAuth(s.handleVoiceChallenge))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/voice-capture", s.withAPIAuth(s.handleVoiceCapture))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/finalize", s.withAPIAuth(s.handleFinalize))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/restart", s.withAPIAuth(s.handleRestart))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/internal/session", s.handleInternalSession)
	mux.HandleFunc("GET /api/v2/verify/{verification_id}/status", s.withAPIAuth(s.handleVerificationStatus))
	mux.HandleFunc("GET /api/v2/verify/{verification_id}/selfie", s.withAPIAuth(s.handleVerificationSelfie))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/phone-otp/send", s.withAPIAuth(s.handlePhoneOTPSend))
	mux.HandleFunc("POST /api/v2/verify/{verification_id}/phone-otp/verify", s.withAPIAuth(s.handlePhoneOTPVerify))

	mux.HandleFunc("GET /api/health", s.handleAPIHealth)
	mux.HandleFunc("GET /api/health/detailed", s.handleAPIHealthDetailed)
	mux.HandleFunc("GET /api/health/ready", s.handleAPIHealthReady)
	mux.HandleFunc("GET /api/health/live", s.handleAPIHealthLive)
	mux.HandleFunc("GET /api/status", s.handleAPIStatus)
	mux.HandleFunc("GET /api/setup/status", s.handleSetupStatus)
	mux.HandleFunc("POST /api/setup/initialize", s.handleSetupInitialize)
	mux.HandleFunc("GET /.well-known/did.json", s.handleDIDDocument)
	mux.HandleFunc("GET /api/v2/verify/page-config/slug/{slug}", s.handlePageConfigSlug)
	mux.HandleFunc("POST /api/developer/register", s.handleDeveloperRegister)
	mux.HandleFunc("POST /api/auth/admin/login", s.handleAdminLogin)
	mux.HandleFunc("POST /api/auth/developer/otp/complete-registration", s.handleOTPCompleteRegistration)
	mux.HandleFunc("POST /api/auth/totp/setup", s.withAPIAuth(s.handleTOTPSetup))
	mux.HandleFunc("POST /api/auth/totp/verify", s.withAPIAuth(s.handleTOTPVerify))
	mux.HandleFunc("POST /api/auth/admin/escalate", s.handleAdminEscalate)
	mux.HandleFunc("POST /api/auth/developer/otp/send", s.handleOTPSend)
	mux.HandleFunc("POST /api/auth/developer/otp/verify", s.handleOTPVerify)
	mux.HandleFunc("POST /api/auth/reviewer/otp/send", s.handleOTPSend)
	mux.HandleFunc("POST /api/auth/reviewer/otp/verify", s.handleOTPVerify)
	mux.HandleFunc("GET /api/auth/developer/github/url", s.handleGitHubURL)
	mux.HandleFunc("POST /api/auth/developer/github/callback", s.handleGitHubCallback)
	if s.cfg.LabMode {
		mux.HandleFunc("POST /v1/liveness-checks", s.handleLivenessCheck)
	}
	if s.cfg.IsDevelopment() && s.cfg.Auth.EnableDevToken {
		mux.HandleFunc("POST /dev/token", s.handleDevToken)
		mux.HandleFunc("GET /dev/face-test.html", s.handleFaceTest)
	}
	base := s.withMiddleware(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.handleCredentialV2Alias(w, r) {
			return
		}
		base.ServeHTTP(w, r)
	})
}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (*app.AuthContext, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "缺少 Bearer token"))
		return nil, false
	}
	auth, err := s.auth.Authenticate(r.Context(), token)
	if err != nil {
		writeError(w, r, err)
		return nil, false
	}
	return auth, true
}

type challengeRequest struct {
	Scene    string `json:"scene"`
	Mode     string `json:"mode"`
	Gallery  string `json:"gallery"`
	DeviceID string `json:"device_id"`
}

type challengeResponse struct {
	ChallengeID string   `json:"challenge_id"`
	Actions     []string `json:"actions"`
	FrameCount  int      `json:"frame_count"`
	ExpiresAt   string   `json:"expires_at"`
	Nonce       string   `json:"nonce"`
}

func (s *Server) handleChallenge(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	var req challengeRequest
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Security.MaxUploadBytes)
	ct := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, uploadError(err))
			return
		}
	case strings.HasPrefix(ct, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeError(w, r, uploadError(err))
			return
		}
		req = challengeRequest{Scene: r.FormValue("scene"), Mode: r.FormValue("mode"), Gallery: r.FormValue("gallery"), DeviceID: r.FormValue("device_id")}
	default:
		if err := r.ParseForm(); err != nil {
			writeError(w, r, uploadError(err))
			return
		}
		req = challengeRequest{Scene: r.FormValue("scene"), Mode: r.FormValue("mode"), Gallery: r.FormValue("gallery"), DeviceID: r.FormValue("device_id")}
	}
	challenge, err := s.challenges.Create(r.Context(), auth, req.Scene, req.Mode, req.Gallery, req.DeviceID, requestIDFrom(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, challengeResponse{
		ChallengeID: challenge.ChallengeID,
		Actions:     challenge.Actions,
		FrameCount:  challenge.FrameCount,
		ExpiresAt:   formatExpires(challenge.ExpiresAt),
		Nonce:       challenge.Nonce,
	})
}

type actionCheckResponse struct {
	Satisfied bool   `json:"satisfied"`
	Detail    string `json:"detail"`
}

func (s *Server) handleActionCheck(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	if err := s.parseMultipart(w, r); err != nil {
		writeError(w, r, err)
		return
	}
	frames, err := readFrames(r.MultipartForm.File["frames"], s.cfg.Face.MaxFrameBytes)
	if err != nil {
		writeError(w, r, err)
		return
	}
	baseline, err := readFrames(r.MultipartForm.File["baseline_frames"], s.cfg.Face.MaxFrameBytes)
	if err != nil {
		writeError(w, r, err)
		return
	}
	satisfied, detail, err := s.actions.CheckAction(r.Context(), r.FormValue("action"), frames, baseline, auth.Policy)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, actionCheckResponse{Satisfied: satisfied, Detail: detail})
}

type enrollResponse struct {
	ProfileID    string `json:"profile_id"`
	Status       string `json:"status"`
	ModelVersion string `json:"model_version"`
}

func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	challengeID, frames, deviceID, err := s.parseUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	challenge, session, err := s.support.BeginUpload(r.Context(), auth, challengeID, deviceID, domain.ModeEnroll)
	if err != nil {
		writeError(w, r, err)
		return
	}
	result, err := s.enroll.Enroll(r.Context(), auth, challenge, session, frames, strPtr(requestIDFrom(r)))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, enrollResponse{
		ProfileID:    strings.ReplaceAll(result.ProfileID.String(), "-", ""),
		Status:       result.Status,
		ModelVersion: result.ModelVersion,
	})
}

type verifyResponse struct {
	Result            string  `json:"result"`
	Score             float64 `json:"score"`
	Liveness          string  `json:"liveness"`
	VerificationToken string  `json:"verification_token,omitempty"`
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	challengeID, frames, deviceID, err := s.parseUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	challenge, session, err := s.support.BeginUpload(r.Context(), auth, challengeID, deviceID, domain.ModeVerify)
	if err != nil {
		writeError(w, r, err)
		return
	}
	result, err := s.verify.Verify(r.Context(), auth, challenge, session, frames, strPtr(requestIDFrom(r)))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, verifyResponse{Result: result.Result, Score: result.Score, Liveness: result.Liveness, VerificationToken: result.Token})
}

type livenessCheckResponse struct {
	Result         string    `json:"result"`
	Status         string    `json:"status"`
	PadScore       float64   `json:"pad_score"`
	PadThreshold   float64   `json:"pad_threshold"`
	PadDecision    string    `json:"pad_decision"`
	MinFrameScore  float64   `json:"min_frame_score"`
	MeanFrameScore float64   `json:"mean_frame_score"`
	FrameScores    []float64 `json:"frame_scores"`
	LiveFrames     int       `json:"live_frames"`
	AttackFrames   int       `json:"attack_frames"`
	LiveRatio      float64   `json:"live_ratio"`
	ActionChecked  bool      `json:"action_checked"`
	ActionPassed   bool      `json:"action_passed"`
	ActionCode     string    `json:"action_code,omitempty"`
	ActionDetail   string    `json:"action_detail,omitempty"`
	ModelVersion   string    `json:"model_version"`
	ConfigVersion  string    `json:"config_version"`
	DeviceID       string    `json:"device_id"`
	RequestID      string    `json:"request_id"`
}

func (s *Server) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	challengeID, frames, deviceID, err := s.parseUpload(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	challenge, session, err := s.support.BeginUpload(r.Context(), auth, challengeID, deviceID, domain.ModeLiveness)
	if err != nil {
		writeError(w, r, err)
		return
	}
	verifyActions := r.FormValue("verify_actions") == "true"
	result, err := s.liveness.Check(r.Context(), auth, challenge, session, frames, verifyActions, strPtr(requestIDFrom(r)))
	if err != nil {
		writeError(w, r, err)
		return
	}
	mean := 0.0
	for _, score := range result.FrameScores {
		mean += score
	}
	if len(result.FrameScores) > 0 {
		mean /= float64(len(result.FrameScores))
	}
	writeOK(w, r, livenessCheckResponse{
		Result:         result.Result,
		Status:         string(result.Status),
		PadScore:       result.MinScore,
		PadThreshold:   auth.Policy.LivenessThreshold,
		PadDecision:    string(result.Status),
		MinFrameScore:  result.MinScore,
		MeanFrameScore: mean,
		FrameScores:    result.FrameScores,
		LiveFrames:     result.LiveFrames,
		AttackFrames:   result.AttackFrames,
		LiveRatio:      result.LiveRatio,
		ActionChecked:  result.ActionChecked,
		ActionPassed:   result.ActionPassed,
		ActionCode:     result.ActionCode,
		ActionDetail:   result.ActionDetail,
		ModelVersion:   s.engineVersion,
		ConfigVersion:  s.configHash,
		DeviceID:       deviceID,
		RequestID:      requestIDFrom(r),
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version":               s.buildVersion,
		"commit":                s.buildCommit,
		"env":                   s.cfg.Env,
		"engine":                s.engineVersion,
		"model_sha256":          s.modelHash,
		"config_sha256":         s.configHash,
		"liveness_threshold":    s.cfg.Face.LivenessThreshold,
		"live_ratio_threshold":  s.cfg.Face.LiveRatioThreshold,
		"actions_count":         s.cfg.Face.ActionsCount,
		"frames_per_action":     s.cfg.Face.FramesPerAction,
		"action_verify_enabled": s.cfg.Face.ActionVerifyEnabled,
		"lab_mode":              s.cfg.LabMode,
	})
}

type subjectResponse struct {
	SubjectID    string  `json:"subject_id"`
	Status       string  `json:"status"`
	ModelVersion *string `json:"model_version,omitempty"`
	RequestID    string  `json:"request_id"`
}

func (s *Server) handleSubjectGet(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	external := r.PathValue("subject_id")
	subject, modelVersion, err := s.subjects.Get(r.Context(), auth.Tenant.ID, external)
	if err != nil {
		writeError(w, r, err)
		return
	}
	status := "pending"
	if subject.Status == domain.SubjectActive && modelVersion != nil {
		status = "active"
	}
	writeOK(w, r, subjectResponse{SubjectID: external, Status: status, ModelVersion: modelVersion, RequestID: requestIDFrom(r)})
}

type deleteSubjectResponse struct {
	SubjectID string          `json:"subject_id"`
	Deleted   string          `json:"deleted"`
	Cleaned   map[string]bool `json:"cleaned"`
	RequestID string          `json:"request_id"`
}

func (s *Server) handleSubjectDelete(w http.ResponseWriter, r *http.Request) {
	auth, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	external := r.PathValue("subject_id")
	_, cleaned, err := s.subjects.Delete(r.Context(), auth.Tenant.ID, external, auth.Claims.Subject, strPtr(requestIDFrom(r)))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, deleteSubjectResponse{SubjectID: external, Deleted: "deleted", Cleaned: cleaned, RequestID: requestIDFrom(r)})
}

type devTokenResponse struct {
	Token     string `json:"token"`
	Tenant    string `json:"tenant"`
	Sub       string `json:"sub"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *Server) handleDevToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sub string `json:"sub"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
	sub := strings.TrimSpace(body.Sub)
	if sub == "" {
		sub = "dev-user"
	}
	token, expires, err := s.issuer.IssueDev(sub)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeOK(w, r, devTokenResponse{Token: token, Tenant: "dev", Sub: sub, ExpiresAt: expires.Unix()})
}

func (s *Server) handleFaceTest(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/face-test.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write(data)
}

func (s *Server) healthLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

func (s *Server) healthReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := s.dbPing(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if err := s.redisPing(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "engine": s.engineVersion})
}

func (s *Server) parseUpload(w http.ResponseWriter, r *http.Request) (string, [][]byte, string, error) {
	if err := s.parseMultipart(w, r); err != nil {
		return "", nil, "", err
	}
	frames, err := readFrames(r.MultipartForm.File["frames"], s.cfg.Face.MaxFrameBytes)
	if err != nil {
		return "", nil, "", err
	}
	return r.FormValue("challenge_id"), frames, r.FormValue("device_id"), nil
}

// parseMultipart 统一限制请求体总大小并解析 multipart。
func (s *Server) parseMultipart(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Security.MaxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return uploadError(err)
	}
	return nil
}

// uploadError 将请求体超限映射为 503，其余解析错误映射为 401。
func uploadError(err error) error {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return domain.New(domain.CodeServiceUnavailable, "请求体过大")
	}
	return domain.New(domain.CodeAuthInvalid, "请求体格式不正确")
}

func readFrames(files []*multipart.FileHeader, maxFrameBytes int64) ([][]byte, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > 15 {
		return nil, domain.New(domain.CodeAuthInvalid, "帧数超过 15")
	}
	out := make([][]byte, 0, len(files))
	for _, header := range files {
		f, err := header.Open()
		if err != nil {
			return nil, domain.New(domain.CodeLowQuality, "帧读取失败")
		}
		data, err := io.ReadAll(io.LimitReader(f, maxFrameBytes+1))
		_ = f.Close()
		if err != nil {
			return nil, domain.New(domain.CodeLowQuality, "帧读取失败")
		}
		if int64(len(data)) > maxFrameBytes {
			return nil, domain.New(domain.CodeLowQuality, "单帧超过大小上限")
		}
		out = append(out, data)
	}
	return out, nil
}

func strPtr(s string) *string { return &s }
