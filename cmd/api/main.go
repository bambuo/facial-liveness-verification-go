// API 组装根：加载配置、连接依赖、注册适配器并启动 HTTP 服务。
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	"facial-liveness-verification-go/internal/adapter/sms"
	"facial-liveness-verification-go/internal/adapter/smtp"
	"facial-liveness-verification-go/internal/adapter/web"
	"facial-liveness-verification-go/internal/adapter/webhook"
	"facial-liveness-verification-go/internal/app"
	"facial-liveness-verification-go/internal/audio"
	"facial-liveness-verification-go/internal/config"
	"facial-liveness-verification-go/internal/deepfake"
	"facial-liveness-verification-go/internal/ocr"
	"facial-liveness-verification-go/internal/port"
	"facial-liveness-verification-go/internal/vc"
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

	gormDB, err := pg.OpenGORM(cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("打开 GORM: %w", err)
	}
	defer func() { _ = pg.CloseGORM(gormDB) }()
	if cfg.DB.AutoMigrate {
		if err := pg.AutoMigrate(ctx, gormDB,
			time.Duration(cfg.DB.MigrationLockSeconds)*time.Second,
			time.Duration(cfg.DB.MigrationStatementSeconds)*time.Second); err != nil {
			return fmt.Errorf("启动迁移: %w", err)
		}
		slog.Info("数据库迁移完成")
	}
	identityStore := pg.NewIdentityStore(gormDB)
	var vcIssuer *vc.Issuer
	if cfg.VC.PrivateKey != "" {
		vcIssuer, err = vc.New(cfg.VC.IssuerDID, cfg.VC.PrivateKey, cfg.VC.TTLDays)
		if err != nil {
			return fmt.Errorf("初始化 VC 签发器: %w", err)
		}
	}

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

	var ocrSvc web.OCRService
	if cfg.OCR.Enabled {
		det := cfg.OCR.DetModel
		if det == "" {
			det = filepath.Join(cfg.OCR.ModelDir, "det.onnx")
		}
		rec := cfg.OCR.RecModel
		if rec == "" {
			rec = filepath.Join(cfg.OCR.ModelDir, "rec.onnx")
		}
		dict := cfg.OCR.DictPath
		if dict == "" {
			dict = filepath.Join(cfg.OCR.ModelDir, "dict.txt")
		}
		ocrCfg := ocr.Config{DetModel: det, RecModel: rec, DictPath: dict, Orientation: cfg.OCR.Orientation, MaxSide: cfg.OCR.MaxSide}
		if cfg.OCR.WidthMode == "bucket" {
			ocrCfg.WidthBucket = 32
		}
		ocrEngine, err := ocr.New(ocrCfg)
		if err != nil {
			return fmt.Errorf("加载 OCR: %w", err)
		}
		defer ocrEngine.Close()
		ocrSvc = ocrEngine
		slog.Info("OCR 已启用", "det", det, "rec", rec, "dict", dict,
			"width_mode", cfg.OCR.WidthMode, "orientation", cfg.OCR.Orientation)
	}

	var deepfakeEngine *deepfake.Engine
	if cfg.Face.DeepfakeModel != "" {
		if _, statErr := os.Stat(cfg.Face.DeepfakeModel); statErr == nil {
			deepfakeEngine, err = deepfake.New(cfg.Face.DeepfakeModel)
			if err != nil {
				return fmt.Errorf("加载深伪模型: %w", err)
			}
			defer deepfakeEngine.Close()
			slog.Info("深伪模型已启用", "model", cfg.Face.DeepfakeModel)
		}
	}

	var voiceEngine *audio.VoiceEngine
	if cfg.Voice.Enabled {
		paths := audio.VoiceModelPaths(cfg.Voice.ModelDir)
		if cfg.Voice.SpeakerModel != "" {
			paths.SpeakerModel = cfg.Voice.SpeakerModel
		}
		if cfg.Voice.ASRModel != "" {
			paths.ASRModel = cfg.Voice.ASRModel
		}
		if cfg.Voice.TokensPath != "" {
			paths.TokensPath = cfg.Voice.TokensPath
		}
		voiceEngine, err = audio.NewVoiceEngine(paths)
		if err != nil {
			return fmt.Errorf("加载语音引擎: %w", err)
		}
		defer voiceEngine.Close()
		slog.Info("语音引擎已启用", "speaker", paths.SpeakerModel, "asr", paths.ASRModel)
	}

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
	subjectSvc := app.NewSubjectService(repo.Subjects(), repo.Profiles(), storage, repo.Audits(), clk)
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
		OCR:           ocrSvc,
		Voice:         voiceEngine,
		Face:          engine,
		Deepfake:      deepfakeEngine,
		IdswyftStore:  identityStore,
		Storage:       storage,
		Webhooks:      webhook.New(),
		SMS:           sms.New(cfg.SMS),
		Mail:          smtp.New(cfg.SMTP),
		Encryptor:     encryptor,
		VC:            vcIssuer,
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
