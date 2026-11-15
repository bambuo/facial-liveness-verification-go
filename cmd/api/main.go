// API 组装根：加载配置、连接依赖、注册适配器并启动 HTTP 服务。
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"facial-liveness-verification-go/internal/adapter/clock"
	"facial-liveness-verification-go/internal/adapter/onnx"
	"facial-liveness-verification-go/internal/adapter/pg"
	"facial-liveness-verification-go/internal/adapter/redis"
	"facial-liveness-verification-go/internal/adapter/s3"
	"facial-liveness-verification-go/internal/adapter/sec"
	"facial-liveness-verification-go/internal/adapter/seed"
	"facial-liveness-verification-go/internal/adapter/web"
	"facial-liveness-verification-go/internal/app"
	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/port"
)

var (
	buildVersion = "dev"
	buildCommit  = "unknown"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api 启动失败", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()

	repo, err := pg.New(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer repo.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	engine, err := onnx.New(onnx.Config{
		ModelDir:           cfg.Face.ModelDir,
		DetectionThreshold: cfg.Face.DetectionThreshold,
		NMSThreshold:       cfg.Face.NMSThreshold,
		AnalysisThreshold:  cfg.Face.AnalysisThreshold,
		LivenessCropScale:  cfg.Face.LivenessCropScale,
	})
	if err != nil {
		return err
	}
	defer engine.Close()

	modelHash, err := modelFingerprint(cfg.Face.ModelDir)
	if err != nil {
		return err
	}

	clk := clock.System{}
	validator, err := sec.NewValidator(cfg.Auth.DevSigningKey)
	if err != nil {
		return err
	}
	issuer, err := sec.NewIssuer(cfg.Auth.DevSigningKey, cfg.Auth.ResultSigningKey, cfg.Auth.TokenTTLSeconds, clk)
	if err != nil {
		return err
	}
	keyStore, err := sec.NewKeyStore(repo.Keys(), cfg.Security.MasterKey)
	if err != nil {
		return err
	}
	encryptor := sec.NewEncryptor(keyStore)
	challengeStore := redis.NewChallengeStore(rdb, time.Duration(cfg.Redis.ChallengeTTLSeconds)*time.Second)
	quotaStore := redis.NewQuotaStore(rdb)
	storage, err := s3.New(ctx, cfg.Storage)
	if err != nil {
		return err
	}

	auditor := app.NewAuditor(repo.Audits(), clk)
	var guard port.LivenessGuard
	if cfg.Guard.Enabled {
		guard = redis.NewLivenessGuard(rdb, cfg.Guard.FailureLimit, cfg.Guard.WindowSeconds, cfg.Guard.LockSeconds)
	}
	support := app.NewVerificationSupport(challengeStore, repo.Sessions(), repo.Subjects(), issuer, auditor, clk, guard)
	actionSvc := app.NewActionService(engine)
	challengeSvc := app.NewChallengeService(challengeStore, repo.Sessions(), auditor, clk, time.Duration(cfg.Redis.ChallengeTTLSeconds)*time.Second)
	enrollSvc := app.NewEnrollService(support, actionSvc, engine, encryptor, repo.Subjects(), repo.Profiles(), auditor, clk)
	verifySvc := app.NewVerifyService(support, actionSvc, engine, encryptor, repo.Subjects(), repo.Profiles())
	livenessSvc := app.NewLivenessService(support, actionSvc, engine)
	subjectSvc := app.NewSubjectService(repo.Subjects(), repo.Profiles(), repo.Index(), storage, repo.Audits(), clk)
	authenticator := app.NewAuthenticator(validator, repo.Tenants(), quotaStore, cfg.BasePolicy())

	if cfg.IsDevelopment() && cfg.Auth.EnableDevToken {
		if err := seed.EnsureDevTenant(ctx, repo.Pool); err != nil {
			return err
		}
		slog.Info("开发端点已启用", "routes", "/dev/token,/dev/face-test.html")
	}

	server := web.NewServer(web.Deps{
		Cfg:           cfg,
		Auth:          authenticator,
		Challenges:    challengeSvc,
		Actions:       actionSvc,
		Enroll:        enrollSvc,
		Verify:        verifySvc,
		Subjects:      subjectSvc,
		Support:       support,
		Liveness:      livenessSvc,
		Issuer:        issuer,
		EngineVersion: engine.ModelVersion(),
		BuildVersion:  buildVersion,
		BuildCommit:   buildCommit,
		ModelHash:     modelHash,
		DBPing:        func(ctx context.Context) error { return repo.Pool.Ping(ctx) },
		RedisPing:     func(ctx context.Context) error { return rdb.Ping(ctx).Err() },
	})

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP 服务已启动",
			"addr", cfg.HTTPAddr, "env", cfg.Env, "version", buildVersion, "commit", buildCommit,
			"engine", engine.ModelVersion(), "model_sha256", modelHash, "config_sha256", cfg.Fingerprint(),
			"liveness_threshold", cfg.Face.LivenessThreshold, "actions_count", cfg.Face.ActionsCount,
			"frames_per_action", cfg.Face.FramesPerAction, "lab_mode", cfg.LabMode)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-stop:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

func modelFingerprint(dir string) (string, error) {
	h := sha256.New()
	for _, name := range []string{
		"face_detection_yunet_2023mar.onnx",
		"face_recognition_sface_2021dec.onnx",
		"2.7_80x80_MiniFASNetV2.onnx",
	} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(name))
		_, _ = h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
