// Package config 从环境变量加载服务配置并做启动期安全校验。
package config

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"facial-liveness-verification-go/internal/domain"
)

// Config 是进程级配置。
type Config struct {
	Env       string
	HTTPAddr  string
	LabMode   bool
	Auth      Auth
	Face      Face
	DB        DB
	Redis     Redis
	Security  Security
	Retention Retention
	RateLimit RateLimit
	Storage   Storage
	Guard     Guard
	OCR       OCR
}

// Auth 是 JWT 与 dev 令牌配置。
type Auth struct {
	EnableDevToken   bool
	DevSigningKey    string
	ResultSigningKey string
	TokenTTLSeconds  int
}

// Face 是模型与阈值配置。
type Face struct {
	ModelDir            string
	DetectionThreshold  float64
	NMSThreshold        float64
	AnalysisThreshold   float64
	MatchThreshold      float64
	LivenessThreshold   float64
	LiveRatioThreshold  float64
	YawThreshold        float64
	StraightYaw         float64
	MouthOpenRatio      float64
	ActionPassRatio     float64
	FramesPerAction     int
	ActionsCount        int
	ActionVerifyEnabled bool
	LivenessCropScale   float64
	MaxFrameBytes       int64
}

// OCR 是 PaddleOCR 引擎配置。
type OCR struct {
	Enabled     bool
	ModelDir    string
	MaxSide     int
	DetModel    string
	RecModel    string
	DictPath    string
	WidthMode   string
	Orientation string
}

// DB 是 PostgreSQL 配置。
type DB struct {
	URL                       string
	AutoMigrate               bool
	MigrationLockSeconds      int
	MigrationStatementSeconds int
}

// Redis 是缓存配置。
type Redis struct {
	Addr                string
	Password            string
	DB                  int
	ChallengeTTLSeconds int
}

// Security 是主密钥与 HTTP 安全配置。
type Security struct {
	MasterKey      string
	AllowedOrigins []string
	MaxUploadBytes int64
}

// Retention 是 Worker 留存配置。
type Retention struct {
	ProfileDays int
}

// Guard 是活体失败锁定配置。
type Guard struct {
	Enabled       bool
	FailureLimit  int
	WindowSeconds int
	LockSeconds   int
}

// RateLimit 是限流配置。
type RateLimit struct {
	PermitPerMinute int
}

// Storage 是 S3 兼容对象存储配置。
type Storage struct {
	Enabled        bool
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

// Load 读取 API 环境变量并返回配置；非法值或生产安全缺失直接失败。
func Load() (*Config, error) { return load(false) }

// LoadWorker 读取 Worker 环境变量；Worker 不需要 JWT、模型与存储配置。
func LoadWorker() (*Config, error) { return load(true) }

func load(worker bool) (*Config, error) {
	var errs []error
	intVal := func(key string, def int) int {
		v, err := getInt(key, def)
		if err != nil {
			errs = append(errs, err)
		}
		return v
	}
	floatVal := func(key string, def float64) float64 {
		v, err := getFloat(key, def)
		if err != nil {
			errs = append(errs, err)
		}
		return v
	}
	boolVal := func(key string, def bool) bool {
		v, err := getBool(key, def)
		if err != nil {
			errs = append(errs, err)
		}
		return v
	}

	c := &Config{
		Env:      getString("APP_ENV", "production"),
		HTTPAddr: getString("HTTP_ADDR", ":8080"),
		LabMode:  boolVal("LAB_MODE", false),
		Auth: Auth{
			EnableDevToken:   boolVal("AUTH_ENABLE_DEV_TOKEN", false),
			DevSigningKey:    getString("AUTH_DEV_SIGNING_KEY", ""),
			ResultSigningKey: getString("AUTH_RESULT_TOKEN_SIGNING_KEY", ""),
			TokenTTLSeconds:  intVal("AUTH_TOKEN_TTL_SECONDS", 300),
		},
		Face: Face{
			ModelDir:            getString("FACE_MODEL_DIR", "models"),
			DetectionThreshold:  floatVal("FACE_DETECTION_THRESHOLD", 0.9),
			NMSThreshold:        floatVal("FACE_NMS_THRESHOLD", 0.3),
			AnalysisThreshold:   floatVal("FACE_ANALYSIS_THRESHOLD", 0.6),
			MatchThreshold:      floatVal("FACE_MATCH_THRESHOLD", 0.35),
			LivenessThreshold:   floatVal("FACE_LIVENESS_THRESHOLD", 0.85),
			LiveRatioThreshold:  floatVal("FACE_LIVE_RATIO_THRESHOLD", 0.5),
			YawThreshold:        floatVal("FACE_YAW_THRESHOLD", 0.20),
			StraightYaw:         floatVal("FACE_STRAIGHT_YAW", 0.15),
			MouthOpenRatio:      floatVal("FACE_MOUTH_OPEN_RATIO", 1.08),
			ActionPassRatio:     floatVal("FACE_ACTION_PASS_RATIO", 0.5),
			FramesPerAction:     intVal("FACE_FRAMES_PER_ACTION", 5),
			ActionsCount:        intVal("FACE_ACTIONS_COUNT", 3),
			ActionVerifyEnabled: boolVal("FACE_ACTION_VERIFY_ENABLED", true),
			LivenessCropScale:   floatVal("FACE_LIVENESS_CROP_SCALE", 2.7),
			MaxFrameBytes:       int64(intVal("FACE_MAX_FRAME_BYTES", 2*1024*1024)),
		},
		DB: DB{
			URL:                       getString("DATABASE_URL", ""),
			AutoMigrate:               boolVal("DB_AUTO_MIGRATE", true),
			MigrationLockSeconds:      intVal("DB_MIGRATION_LOCK_SECONDS", 30),
			MigrationStatementSeconds: intVal("DB_MIGRATION_STATEMENT_SECONDS", 120),
		},
		Redis: Redis{
			Addr:                getString("REDIS_ADDR", ""),
			Password:            getString("REDIS_PASSWORD", ""),
			DB:                  intVal("REDIS_DB", 0),
			ChallengeTTLSeconds: intVal("REDIS_CHALLENGE_TTL_SECONDS", 300),
		},
		Security: Security{
			MasterKey:      getString("SECURITY_MASTER_KEY", ""),
			AllowedOrigins: splitOrigins(getString("SECURITY_ALLOWED_ORIGINS", "")),
			MaxUploadBytes: int64(intVal("SECURITY_MAX_UPLOAD_BYTES", 20*1024*1024)),
		},
		Retention: Retention{ProfileDays: intVal("RETENTION_PROFILE_DAYS", 90)},
		RateLimit: RateLimit{PermitPerMinute: intVal("RATE_LIMIT_PERMIT_PER_MINUTE", 60)},
		Guard: Guard{
			Enabled:       boolVal("LIVENESS_GUARD_ENABLED", true),
			FailureLimit:  intVal("LIVENESS_GUARD_FAILURE_LIMIT", 5),
			WindowSeconds: intVal("LIVENESS_GUARD_WINDOW_SECONDS", 600),
			LockSeconds:   intVal("LIVENESS_GUARD_LOCK_SECONDS", 300),
		},
		Storage: Storage{
			Enabled:        boolVal("STORAGE_ENABLED", false),
			Endpoint:       getString("STORAGE_ENDPOINT", ""),
			Bucket:         getString("STORAGE_BUCKET", ""),
			Region:         getString("STORAGE_REGION", "auto"),
			AccessKey:      getString("STORAGE_ACCESS_KEY", ""),
			SecretKey:      getString("STORAGE_SECRET_KEY", ""),
			ForcePathStyle: boolVal("STORAGE_FORCE_PATH_STYLE", true),
		},
		OCR: OCR{
			Enabled:     boolVal("OCR_ENABLED", false),
			ModelDir:    getString("OCR_MODEL_DIR", "models/ocr"),
			MaxSide:     intVal("OCR_MAX_SIDE", 640),
			DetModel:    getString("OCR_DET_MODEL", ""),
			RecModel:    getString("OCR_REC_MODEL", ""),
			DictPath:    getString("OCR_DICT", ""),
			WidthMode:   getString("OCR_WIDTH_MODE", "exact"),
			Orientation: getString("OCR_ORIENTATION", "0"),
		},
	}

	errs = append(errs, validate(c, worker)...)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return c, nil
}

// validate 汇集启动期安全与范围校验；worker 跳过 API 专用配置。
func validate(c *Config, worker bool) []error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	switch c.Env {
	case "development", "staging", "production":
	default:
		add("APP_ENV 必须是 development/staging/production，当前 %q", c.Env)
	}
	if c.Env != "development" && c.Auth.EnableDevToken {
		add("非 development 环境不得启用 AUTH_ENABLE_DEV_TOKEN")
	}
	if c.Env == "production" && c.LabMode {
		add("生产配置不得启用 LAB_MODE")
	}
	if c.OCR.Enabled {
		if c.OCR.MaxSide <= 0 {
			add("OCR_MAX_SIDE 必须大于 0")
		}
		if c.OCR.ModelDir == "" && c.OCR.DetModel == "" {
			add("OCR_ENABLED 时 OCR_MODEL_DIR 或 OCR_DET_MODEL 不能为空")
		}
		switch c.OCR.WidthMode {
		case "bucket", "exact":
		default:
			add("OCR_WIDTH_MODE 必须是 bucket 或 exact，当前 %q", c.OCR.WidthMode)
		}
		switch c.OCR.Orientation {
		case "0", "auto", "90", "180", "270":
		default:
			add("OCR_ORIENTATION 必须是 0/auto/90/180/270，当前 %q", c.OCR.Orientation)
		}
	}
	if c.DB.URL == "" {
		add("DATABASE_URL 不能为空")
	}
	if c.DB.AutoMigrate {
		if c.DB.MigrationLockSeconds <= 0 {
			add("DB_MIGRATION_LOCK_SECONDS 必须大于 0")
		}
		if c.DB.MigrationStatementSeconds <= 0 {
			add("DB_MIGRATION_STATEMENT_SECONDS 必须大于 0")
		}
	}
	if c.Redis.Addr == "" {
		add("REDIS_ADDR 不能为空")
	}
	if c.Redis.ChallengeTTLSeconds <= 0 {
		add("REDIS_CHALLENGE_TTL_SECONDS 必须大于 0")
	}
	if c.Retention.ProfileDays < 0 {
		add("RETENTION_PROFILE_DAYS 不能为负数")
	}
	if worker {
		return errs
	}

	if c.Auth.DevSigningKey == "" {
		add("AUTH_DEV_SIGNING_KEY 不能为空")
	} else if _, err := base64.StdEncoding.DecodeString(c.Auth.DevSigningKey); err != nil {
		add("AUTH_DEV_SIGNING_KEY 必须是 Base64: %v", err)
	}
	if c.Auth.ResultSigningKey == "" {
		add("AUTH_RESULT_TOKEN_SIGNING_KEY 不能为空")
	}
	if c.Auth.TokenTTLSeconds <= 0 {
		add("AUTH_TOKEN_TTL_SECONDS 必须大于 0")
	}
	if c.Security.MasterKey != "" {
		if _, err := base64.StdEncoding.DecodeString(c.Security.MasterKey); err != nil {
			add("SECURITY_MASTER_KEY 必须是 Base64: %v", err)
		}
	} else if c.Env == "production" {
		add("生产配置必须设置 SECURITY_MASTER_KEY")
	}
	if c.RateLimit.PermitPerMinute <= 0 {
		add("RATE_LIMIT_PERMIT_PER_MINUTE 必须大于 0")
	}
	if c.Guard.Enabled {
		if c.Guard.FailureLimit <= 0 {
			add("LIVENESS_GUARD_FAILURE_LIMIT 必须大于 0")
		}
		if c.Guard.WindowSeconds <= 0 {
			add("LIVENESS_GUARD_WINDOW_SECONDS 必须大于 0")
		}
		if c.Guard.LockSeconds <= 0 {
			add("LIVENESS_GUARD_LOCK_SECONDS 必须大于 0")
		}
	}
	if c.Storage.Enabled {
		if c.Storage.Endpoint == "" || c.Storage.Bucket == "" || c.Storage.AccessKey == "" || c.Storage.SecretKey == "" {
			add("启用 STORAGE 时必须配置 ENDPOINT、BUCKET、ACCESS_KEY、SECRET_KEY")
		}
	}
	if c.Face.MaxFrameBytes <= 0 {
		add("FACE_MAX_FRAME_BYTES 必须大于 0")
	}
	if c.Security.MaxUploadBytes <= 0 {
		add("SECURITY_MAX_UPLOAD_BYTES 必须大于 0")
	} else if c.Security.MaxUploadBytes < c.Face.MaxFrameBytes {
		add("SECURITY_MAX_UPLOAD_BYTES 不能小于 FACE_MAX_FRAME_BYTES")
	}

	errs = append(errs, validateFace(c.Face)...)
	return errs
}

// validateFace 校验人脸阈值与动作参数范围。
func validateFace(f Face) []error {
	var errs []error
	check := func(key string, v, min, max float64) {
		if v < min || v > max {
			errs = append(errs, fmt.Errorf("%s 必须在 [%g, %g]，当前 %g", key, min, max, v))
		}
	}
	positive := func(key string, v float64) {
		if v <= 0 {
			errs = append(errs, fmt.Errorf("%s 必须大于 0", key))
		}
	}
	check("FACE_DETECTION_THRESHOLD", f.DetectionThreshold, 0, 1)
	check("FACE_NMS_THRESHOLD", f.NMSThreshold, 0, 1)
	check("FACE_ANALYSIS_THRESHOLD", f.AnalysisThreshold, 0, 1)
	check("FACE_MATCH_THRESHOLD", f.MatchThreshold, 0, 1)
	check("FACE_LIVENESS_THRESHOLD", f.LivenessThreshold, 0, 1)
	check("FACE_LIVE_RATIO_THRESHOLD", f.LiveRatioThreshold, 0, 1)
	check("FACE_ACTION_PASS_RATIO", f.ActionPassRatio, 0, 1)
	positive("FACE_YAW_THRESHOLD", f.YawThreshold)
	positive("FACE_STRAIGHT_YAW", f.StraightYaw)
	positive("FACE_MOUTH_OPEN_RATIO", f.MouthOpenRatio)
	positive("FACE_LIVENESS_CROP_SCALE", f.LivenessCropScale)
	if f.FramesPerAction < 2 || f.FramesPerAction > 8 {
		errs = append(errs, fmt.Errorf("FACE_FRAMES_PER_ACTION 必须在 [2, 8]，当前 %d", f.FramesPerAction))
	}
	if f.ActionsCount < 1 || f.ActionsCount > 3 {
		errs = append(errs, fmt.Errorf("FACE_ACTIONS_COUNT 必须在 [1, 3]，当前 %d", f.ActionsCount))
	}
	actions := 1
	if f.ActionVerifyEnabled {
		actions = f.ActionsCount
	}
	if f.FramesPerAction*actions > 15 {
		errs = append(errs, fmt.Errorf("FACE_FRAMES_PER_ACTION × 动作数不能超过 15，当前 %d × %d", f.FramesPerAction, actions))
	}
	return errs
}

// IsDevelopment 返回是否为本地开发环境。
func (c *Config) IsDevelopment() bool { return c.Env == "development" }

// Fingerprint 返回不含密钥的配置指纹，用于冻结测试版本。
func (c *Config) Fingerprint() string {
	payload := struct {
		Face      Face      `json:"face"`
		RateLimit RateLimit `json:"rate_limit"`
		Retention Retention `json:"retention"`
		LabMode   bool      `json:"lab_mode"`
		TokenTTL  int       `json:"token_ttl_seconds"`
	}{
		Face:      c.Face,
		RateLimit: c.RateLimit,
		Retention: c.Retention,
		LabMode:   c.LabMode,
		TokenTTL:  c.Auth.TokenTTLSeconds,
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// BasePolicy 由全局配置构造领域策略，租户 policy_json 再覆盖。
func (c *Config) BasePolicy() domain.Policy {
	return domain.Policy{
		MatchThreshold:       c.Face.MatchThreshold,
		LivenessThreshold:    c.Face.LivenessThreshold,
		LiveRatioThreshold:   c.Face.LiveRatioThreshold,
		YawThreshold:         c.Face.YawThreshold,
		StraightYaw:          c.Face.StraightYaw,
		MouthOpenRatio:       c.Face.MouthOpenRatio,
		ActionPassRatio:      c.Face.ActionPassRatio,
		FramesPerAction:      c.Face.FramesPerAction,
		ActionsCount:         c.Face.ActionsCount,
		ActionVerifyEnabled:  c.Face.ActionVerifyEnabled,
		MaxRequestsPerMinute: c.RateLimit.PermitPerMinute,
	}
}

func getString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def, fmt.Errorf("%s 必须是整数: %w", key, err)
	}
	return n, nil
}

func getFloat(key string, def float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def, fmt.Errorf("%s 必须是数字: %w", key, err)
	}
	return f, nil
}

func getBool(key string, def bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def, fmt.Errorf("%s 必须是布尔值: %w", key, err)
	}
	return b, nil
}

func splitOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
