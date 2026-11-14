// Package config 从环境变量加载服务配置并做启动期安全校验。
package config

import (
	"encoding/base64"
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
	Auth      Auth
	Face      Face
	DB        DB
	Redis     Redis
	Security  Security
	Retention Retention
	RateLimit RateLimit
	Storage   Storage
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

// DB 是 PostgreSQL 配置。
type DB struct {
	URL       string
	VectorDim int
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

// RateLimit 是限流配置。
type RateLimit struct {
	PermitPerMinute int
}

// Storage 是 S3 兼容对象存储配置。
type Storage struct {
	Enabled             bool
	Endpoint            string
	Bucket              string
	Region              string
	AccessKey           string
	SecretKey           string
	ForcePathStyle      bool
	SignedURLTTLSeconds int
}

// Load 读取环境变量并返回配置；生产环境启用 dev token 时失败。
func Load() (*Config, error) {
	c := &Config{
		Env:      getString("APP_ENV", "production"),
		HTTPAddr: getString("HTTP_ADDR", ":8080"),
		Auth: Auth{
			EnableDevToken:   getBool("AUTH_ENABLE_DEV_TOKEN", false),
			DevSigningKey:    getString("AUTH_DEV_SIGNING_KEY", ""),
			ResultSigningKey: getString("AUTH_RESULT_TOKEN_SIGNING_KEY", ""),
			TokenTTLSeconds:  getInt("AUTH_TOKEN_TTL_SECONDS", 300),
		},
		Face: Face{
			ModelDir:            getString("FACE_MODEL_DIR", "models"),
			DetectionThreshold:  getFloat("FACE_DETECTION_THRESHOLD", 0.9),
			NMSThreshold:        getFloat("FACE_NMS_THRESHOLD", 0.3),
			AnalysisThreshold:   getFloat("FACE_ANALYSIS_THRESHOLD", 0.6),
			MatchThreshold:      getFloat("FACE_MATCH_THRESHOLD", 0.35),
			LivenessThreshold:   getFloat("FACE_LIVENESS_THRESHOLD", 0.85),
			LiveRatioThreshold:  getFloat("FACE_LIVE_RATIO_THRESHOLD", 0.5),
			YawThreshold:        getFloat("FACE_YAW_THRESHOLD", 0.20),
			StraightYaw:         getFloat("FACE_STRAIGHT_YAW", 0.15),
			MouthOpenRatio:      getFloat("FACE_MOUTH_OPEN_RATIO", 1.08),
			ActionPassRatio:     getFloat("FACE_ACTION_PASS_RATIO", 0.5),
			FramesPerAction:     getInt("FACE_FRAMES_PER_ACTION", 5),
			ActionsCount:        getInt("FACE_ACTIONS_COUNT", 3),
			ActionVerifyEnabled: getBool("FACE_ACTION_VERIFY_ENABLED", true),
			LivenessCropScale:   getFloat("FACE_LIVENESS_CROP_SCALE", 2.7),
			MaxFrameBytes:       int64(getInt("FACE_MAX_FRAME_BYTES", 2*1024*1024)),
		},
		DB: DB{
			URL:       getString("DATABASE_URL", ""),
			VectorDim: getInt("DB_VECTOR_DIMENSION", 128),
		},
		Redis: Redis{
			Addr:                getString("REDIS_ADDR", ""),
			Password:            getString("REDIS_PASSWORD", ""),
			DB:                  getInt("REDIS_DB", 0),
			ChallengeTTLSeconds: getInt("REDIS_CHALLENGE_TTL_SECONDS", 300),
		},
		Security: Security{
			MasterKey:      getString("SECURITY_MASTER_KEY", ""),
			AllowedOrigins: splitOrigins(getString("SECURITY_ALLOWED_ORIGINS", "")),
			MaxUploadBytes: int64(getInt("SECURITY_MAX_UPLOAD_BYTES", 20*1024*1024)),
		},
		Retention: Retention{ProfileDays: getInt("RETENTION_PROFILE_DAYS", 90)},
		RateLimit: RateLimit{PermitPerMinute: getInt("RATE_LIMIT_PERMIT_PER_MINUTE", 60)},
		Storage: Storage{
			Enabled:             getBool("STORAGE_ENABLED", false),
			Endpoint:            getString("STORAGE_ENDPOINT", ""),
			Bucket:              getString("STORAGE_BUCKET", ""),
			Region:              getString("STORAGE_REGION", "auto"),
			AccessKey:           getString("STORAGE_ACCESS_KEY", ""),
			SecretKey:           getString("STORAGE_SECRET_KEY", ""),
			ForcePathStyle:      getBool("STORAGE_FORCE_PATH_STYLE", true),
			SignedURLTTLSeconds: getInt("STORAGE_SIGNED_URL_TTL_SECONDS", 900),
		},
	}

	if c.Env != "development" && c.Auth.EnableDevToken {
		return nil, fmt.Errorf("生产配置不得启用 AUTH_ENABLE_DEV_TOKEN")
	}
	if c.Auth.DevSigningKey != "" {
		if _, err := base64.StdEncoding.DecodeString(c.Auth.DevSigningKey); err != nil {
			return nil, fmt.Errorf("AUTH_DEV_SIGNING_KEY 必须是 Base64: %w", err)
		}
	}
	if c.Security.MasterKey != "" {
		if _, err := base64.StdEncoding.DecodeString(c.Security.MasterKey); err != nil {
			return nil, fmt.Errorf("SECURITY_MASTER_KEY 必须是 Base64: %w", err)
		}
	}
	if c.DB.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL 不能为空")
	}
	if c.Redis.Addr == "" {
		return nil, fmt.Errorf("REDIS_ADDR 不能为空")
	}
	return c, nil
}

// IsDevelopment 返回是否为本地开发环境。
func (c *Config) IsDevelopment() bool { return c.Env == "development" }

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

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
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
