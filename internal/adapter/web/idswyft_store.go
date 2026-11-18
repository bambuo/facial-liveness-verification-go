package web

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/port"
	"facial-liveness-verification-go/internal/store"
)

// idswyftStore 是 P1 鉴权与核心验证接口所需的仓储。
// ObjectStore 是 P1 所需的 S3 兼容对象存储子集。
type ObjectStore interface {
	Put(ctx context.Context, key string, content []byte, contentType string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type idswyftStore interface {
	port.IdswyftAuthStore
	CreateVerification(ctx context.Context, verification *gormmodel.VerificationRequest) error
	GetVerification(ctx context.Context, id uuid.UUID) (*gormmodel.VerificationRequest, error)
	UpdateVerification(ctx context.Context, id uuid.UUID, updates map[string]any) error
	CreateDocument(ctx context.Context, document *gormmodel.Document) error
	GetDocument(ctx context.Context, id uuid.UUID) (*gormmodel.Document, error)
	GetBackDocument(ctx context.Context, verificationID uuid.UUID) (*gormmodel.Document, error)
	CreateSelfie(ctx context.Context, selfie *gormmodel.Selfie) error
	GetSelfie(ctx context.Context, id uuid.UUID) (*gormmodel.Selfie, error)
	UpsertContext(ctx context.Context, verificationID uuid.UUID, payload []byte) error
	GetContext(ctx context.Context, verificationID uuid.UUID) ([]byte, error)
	CheckPhoneOTPSend(ctx context.Context, verificationID uuid.UUID, maxSends, windowSeconds int) (bool, error)
	CreatePhoneOTP(ctx context.Context, verificationID uuid.UUID, phone, codeHash string, ttl time.Duration) error
	VerifyPhoneOTP(ctx context.Context, verificationID uuid.UUID, codeHash string, maxAttempts int) (store.OTPResult, error)
	FindDeveloperByEmail(ctx context.Context, email string) (*gormmodel.Developer, error)
	CheckDeveloperOTPSend(ctx context.Context, email string, maxSends, windowSeconds int) (bool, error)
	CreateDeveloperOTP(ctx context.Context, email, codeHash string, ttl time.Duration) error
	VerifyDeveloperOTP(ctx context.Context, email, codeHash string, maxAttempts int) (store.OTPResult, error)
	WebhookForDeveloper(ctx context.Context, developerID uuid.UUID) (*port.WebhookTarget, error)
	CreateWebhookDelivery(ctx context.Context, delivery *gormmodel.WebhookDelivery) error
	UpdateWebhookDelivery(ctx context.Context, id uuid.UUID, updates map[string]any) error
}

// WebhookDispatcher 发送一次带签名的 Webhook。
type WebhookDispatcher interface {
	Deliver(ctx context.Context, target, secret string, payload any) (int, error)
}

// SMSSender 发送短信验证码。
type SMSSender interface {
	Send(ctx context.Context, phone, message string) error
}

// Mailer 发送事务邮件。
type Mailer interface {
	Send(to, subject, body string) error
}
