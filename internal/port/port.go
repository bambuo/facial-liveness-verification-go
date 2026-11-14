// Package port 定义应用层依赖的外部能力接口，由 adapter 实现。
package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
)

// Clock 提供可替换的当前时间。
type Clock interface {
	Now() time.Time
}

// FaceEngine 是人脸检测、特征与活体能力的唯一入口。
type FaceEngine interface {
	ModelVersion() string
	DetectAll(image []byte) ([]domain.FaceBox, error)
	ExtractLargest(frames [][]byte) ([]float32, error)
	CheckLiveness(frames [][]byte, threshold, liveRatioThreshold float64) (domain.LivenessVote, error)
	Analyze(frames [][]byte) ([]domain.FrameMetric, error)
	Compare(left, right []float32) (float64, error)
}

// TenantRepo 按 ID 或 code 解析租户。
type TenantRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetByCode(ctx context.Context, code string) (*domain.Tenant, error)
}

// SubjectRepo 读写主体。
type SubjectRepo interface {
	Get(ctx context.Context, tenantID, subjectID uuid.UUID) (*domain.Subject, error)
	GetByExternal(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Subject, error)
	GetOrCreate(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Subject, error)
	Update(ctx context.Context, subject *domain.Subject) error
}

// ProfileRepo 读写人脸档案。
type ProfileRepo interface {
	GetActive(ctx context.Context, tenantID, subjectID uuid.UUID) (*domain.Profile, error)
	Upsert(ctx context.Context, profile *domain.Profile) (*domain.Profile, error)
	Update(ctx context.Context, profile *domain.Profile) error
	ListDeletedExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Profile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionRepo 读写核验会话。
type SessionRepo interface {
	Add(ctx context.Context, session *domain.Session) error
	GetByChallengeID(ctx context.Context, tenantID uuid.UUID, challengeID string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	ListExpired(ctx context.Context, now time.Time, limit int) ([]*domain.Session, error)
}

// AuditRepo 读写审计。
type AuditRepo interface {
	Add(ctx context.Context, log *domain.AuditLog) error
	ListByResource(ctx context.Context, tenantID uuid.UUID, resource string) ([]*domain.AuditLog, error)
	Update(ctx context.Context, log *domain.AuditLog) error
}

// DocumentRepo 仅供 Worker 清理过期证件记录。
type DocumentRepo interface {
	ListExpired(ctx context.Context, now time.Time, limit int) ([]*domain.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// KeyRepo 提供租户密钥密文的持久化。
type KeyRepo interface {
	FindActive(ctx context.Context, tenantID uuid.UUID) (cipher []byte, version int, err error)
	MaxVersion(ctx context.Context, tenantID uuid.UUID) (int, error)
	Insert(ctx context.Context, tenantID uuid.UUID, version int, cipher []byte) error
}

// ChallengeStore 是一次性 challenge 存储。
type ChallengeStore interface {
	Create(ctx context.Context, challenge *domain.Challenge, ttl time.Duration) error
	Consume(ctx context.Context, challengeID string) (*domain.Challenge, error)
}

// QuotaStore 是租户分钟配额计数器。
type QuotaStore interface {
	TryIncrement(ctx context.Context, tenantID uuid.UUID, limit int) (bool, error)
}

// Encryptor 提供租户级 AES-GCM 加解密。
type Encryptor interface {
	Encrypt(ctx context.Context, tenantID uuid.UUID, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, tenantID uuid.UUID, cipher []byte) ([]byte, error)
}

// KeyStore 提供租户当前数据密钥。
type KeyStore interface {
	CurrentKey(ctx context.Context, tenantID uuid.UUID) ([]byte, error)
}

// TokenValidator 校验端上 JWT。
type TokenValidator interface {
	ValidateEndpoint(token string) (domain.EndpointClaims, error)
}

// TokenIssuer 签发结果令牌与 dev 令牌。
type TokenIssuer interface {
	IssueResult(payload domain.ResultTokenPayload) (string, error)
	IssueDev(sub string) (token string, expiresAt time.Time, err error)
}

// FaceIndex 提供向量成员的合规清理。
type FaceIndex interface {
	DeleteSubject(ctx context.Context, tenantID, subjectID uuid.UUID) error
}

// ObjectStorage 是 S3 兼容对象存储。
type ObjectStorage interface {
	Put(ctx context.Context, key string, content []byte, contentType string) error
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}
