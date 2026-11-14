// Package domain 定义人脸核验的实体、值对象、领域规则与业务错误。
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantStatus 表示租户是否可用。
type TenantStatus string

const (
	TenantActive   TenantStatus = "active"
	TenantDisabled TenantStatus = "disabled"
)

// SubjectStatus 表示主体是否可用。
type SubjectStatus string

const (
	SubjectActive  SubjectStatus = "active"
	SubjectDeleted SubjectStatus = "deleted"
)

// ProfileStatus 表示档案生命周期状态。
type ProfileStatus string

const (
	ProfileEnrolling ProfileStatus = "enrolling"
	ProfileActive    ProfileStatus = "active"
	ProfileDeleting  ProfileStatus = "deleting"
	ProfileDeleted   ProfileStatus = "deleted"
)

// SessionStatus 表示核验会话状态。
type SessionStatus string

const (
	SessionCreated     SessionStatus = "created"
	SessionChallenging SessionStatus = "challenging"
	SessionCollecting  SessionStatus = "collecting"
	SessionProcessing  SessionStatus = "processing"
	SessionPassed      SessionStatus = "passed"
	SessionFailed      SessionStatus = "failed"
	SessionExpired     SessionStatus = "expired"
)

// LivenessStatus 表示静默活体结论。
type LivenessStatus string

const (
	LivenessReal      LivenessStatus = "real"
	LivenessSpoof     LivenessStatus = "spoof"
	LivenessFuzzy     LivenessStatus = "fuzzy"
	LivenessDetecting LivenessStatus = "detecting"
	LivenessError     LivenessStatus = "error"
)

// MatchMode 表示核验模式，与 JWT ops 和路由对应。
type MatchMode string

const (
	ModeMatch    MatchMode = "match"
	ModeLiveness MatchMode = "liveness"
	ModeEnroll   MatchMode = "enroll"
	ModeVerify   MatchMode = "verify"
	ModeIdentify MatchMode = "identify"
	ModeDocument MatchMode = "document"
)

// ParseMatchMode 不区分大小写解析核验模式。
func ParseMatchMode(s string) (MatchMode, bool) {
	for _, m := range []MatchMode{ModeMatch, ModeLiveness, ModeEnroll, ModeVerify, ModeIdentify, ModeDocument} {
		if string(m) == lower(s) {
			return m, true
		}
	}
	return "", false
}

// Tenant 对应 face_tenants。
type Tenant struct {
	ID         uuid.UUID
	Code       string
	Name       string
	Status     TenantStatus
	PolicyJSON []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Subject 对应 face_subjects，只保存业务用户标识。
type Subject struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ExternalID string
	Status     SubjectStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// Profile 对应 face_profiles，只保存加密特征。
type Profile struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	SubjectID       uuid.UUID
	EmbeddingCipher []byte
	KeyVersion      int
	ModelVersion    string
	Status          ProfileStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Session 对应 face_verification_sessions。
type Session struct {
	ID          uuid.UUID
	ChallengeID string
	TenantID    uuid.UUID
	SubjectID   *uuid.UUID
	Scene       string
	Ref         *string
	Mode        MatchMode
	Status      SessionStatus
	Score       *float64
	Liveness    *LivenessStatus
	Clarity     *float64
	Reality     *float64
	DeviceID    *string
	RequestID   *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// AuditLog 对应 face_audit_logs，只记录非敏感元数据。
type AuditLog struct {
	ID         int64
	TenantID   uuid.UUID
	Actor      string
	Action     string
	Resource   string
	Result     string
	RequestID  *string
	DetailJSON []byte
	CreatedAt  time.Time
}

// Document 对应 face_documents，仅用于留存清理。
type Document struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	RetentionUntil *time.Time
}

// Challenge 是 Redis 中的一次性采集凭证。
type Challenge struct {
	ChallengeID     string    `json:"ChallengeId"`
	TenantID        uuid.UUID `json:"TenantId"`
	ExternalSubject string    `json:"ExternalSubjectId"`
	Scene           string    `json:"Scene"`
	Ref             *string   `json:"Ref"`
	Mode            MatchMode `json:"Mode"`
	Gallery         *string   `json:"Gallery"`
	DeviceID        string    `json:"DeviceId"`
	Nonce           string    `json:"Nonce"`
	Actions         []string  `json:"Actions"`
	FramesPerAction int       `json:"FramesPerAction"`
	FrameCount      int       `json:"FrameCount"`
	ExpiresAt       time.Time `json:"ExpiresAt"`
}

// FaceBox 是检测到的人脸框，Landmarks 为 YuNet 5 点。
type FaceBox struct {
	X1, Y1, X2, Y2 float64
	Score          float64
	Landmarks      [5][2]float64
	HasLandmarks   bool
}

// FrameMetric 是单帧关键点几何指标。
type FrameMetric struct {
	FaceFound  bool
	YawRatio   float64
	MouthRatio float64
}

// LivenessVote 是多帧活体投票结果。
type LivenessVote struct {
	Status    LivenessStatus
	MinScore  float64
	LiveRatio float64
}

// EndpointClaims 是端上 JWT 的解析结果。
type EndpointClaims struct {
	Tenant  string
	Subject string
	Scene   string
	Ref     *string
	Ops     []string
	Gallery *string
	Device  *string
	Jti     string
}

// ResultTokenPayload 是结果令牌的签发内容。
type ResultTokenPayload struct {
	TenantCode string
	Subject    string
	Scene      string
	Ref        *string
	Result     string
	Liveness   *LivenessStatus
	Score      *float64
	Jti        string
}

// lower 是 ASCII 小写转换，避免依赖 strings 的大小写规则。
func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}
