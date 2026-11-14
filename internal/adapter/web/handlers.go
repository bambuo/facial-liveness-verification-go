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
	issuer        port.TokenIssuer
	engineVersion string
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
	Issuer        port.TokenIssuer
	EngineVersion string
	DBPing        func(context.Context) error
	RedisPing     func(context.Context) error
}

// NewServer 构造 HTTP 服务。
func NewServer(d Deps) *Server {
	return &Server{
		cfg: d.Cfg, auth: d.Auth, challenges: d.Challenges, actions: d.Actions,
		enroll: d.Enroll, verify: d.Verify, subjects: d.Subjects, support: d.Support,
		issuer: d.Issuer, engineVersion: d.EngineVersion, dbPing: d.DBPing, redisPing: d.RedisPing,
		limiter: newIPLimiter(d.Cfg.RateLimit.PermitPerMinute),
	}
}

// Handler 注册路由并套上中间件。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", s.healthLive)
	mux.HandleFunc("GET /health/ready", s.healthReady)
	mux.HandleFunc("POST /v1/challenges", s.handleChallenge)
	mux.HandleFunc("POST /v1/action-check", s.handleActionCheck)
	mux.HandleFunc("POST /v1/enrollments", s.handleEnroll)
	mux.HandleFunc("POST /v1/verifications", s.handleVerify)
	mux.HandleFunc("GET /v1/subjects/{subject_id}", s.handleSubjectGet)
	mux.HandleFunc("DELETE /v1/subjects/{subject_id}", s.handleSubjectDelete)
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
	ct := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, domain.New(domain.CodeAuthInvalid, "请求体格式不正确"))
			return
		}
	case strings.HasPrefix(ct, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeError(w, r, domain.New(domain.CodeAuthInvalid, "请求体格式不正确"))
			return
		}
		req = challengeRequest{Scene: r.FormValue("scene"), Mode: r.FormValue("mode"), Gallery: r.FormValue("gallery"), DeviceID: r.FormValue("device_id")}
	default:
		if err := r.ParseForm(); err != nil {
			writeError(w, r, domain.New(domain.CodeAuthInvalid, "请求体格式不正确"))
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
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, r, domain.New(domain.CodeAuthInvalid, "请求体格式不正确"))
		return
	}
	frames, err := readFrames(r.MultipartForm.File["frames"])
	if err != nil {
		writeError(w, r, err)
		return
	}
	baseline, err := readFrames(r.MultipartForm.File["baseline_frames"])
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
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Security.MaxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return "", nil, "", domain.New(domain.CodeServiceUnavailable, "请求体过大")
		}
		return "", nil, "", domain.New(domain.CodeAuthInvalid, "请求体格式不正确")
	}
	frames, err := readFrames(r.MultipartForm.File["frames"])
	if err != nil {
		return "", nil, "", err
	}
	return r.FormValue("challenge_id"), frames, r.FormValue("device_id"), nil
}

func readFrames(files []*multipart.FileHeader) ([][]byte, error) {
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
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return nil, domain.New(domain.CodeLowQuality, "帧读取失败")
		}
		out = append(out, data)
	}
	return out, nil
}

func strPtr(s string) *string { return &s }
