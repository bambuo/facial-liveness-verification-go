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
	"time"

	"facial-liveness-verification-go/internal/app"
	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
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
	face          port.FaceEngine
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
	Face          port.FaceEngine
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
		liveness: d.Liveness, issuer: d.Issuer, ocr: d.OCR, face: d.Face, engineVersion: d.EngineVersion,
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
	if s.cfg.LabMode {
		mux.HandleFunc("POST /v1/liveness-checks", s.handleLivenessCheck)
	}
	if s.cfg.IsDevelopment() && s.cfg.Auth.EnableDevToken {
		mux.HandleFunc("POST /dev/token", s.handleDevToken)
		mux.HandleFunc("GET /dev/face-test.html", s.handleFaceTest)
	}
	return s.withMiddleware(mux)
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
