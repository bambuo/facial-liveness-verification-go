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
// DataEncryptor 加密/解密保险库载荷。
type DataEncryptor interface {
	Encrypt(ctx context.Context, tenantID uuid.UUID, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, tenantID uuid.UUID, cipherText []byte) ([]byte, error)
}

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
	GetDeveloperByID(ctx context.Context, id uuid.UUID) (*gormmodel.Developer, error)
	GetDeveloperByAPIKey(ctx context.Context, apiKeyID uuid.UUID) (*gormmodel.Developer, error)
	CreateBatchJob(ctx context.Context, job *gormmodel.BatchJob, items []*gormmodel.BatchItem) error
	GetBatchJob(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.BatchJob, error)
	UpdateBatchJob(ctx context.Context, id uuid.UUID, updates map[string]any) error
	ListBatchJobs(ctx context.Context, developerID uuid.UUID, page, limit int) ([]gormmodel.BatchJob, int64, error)
	ListBatchItems(ctx context.Context, batchID uuid.UUID) ([]gormmodel.BatchItem, error)
	UpdateBatchItem(ctx context.Context, id uuid.UUID, updates map[string]any) error
	GetVaultByVerification(ctx context.Context, developerID, verificationID uuid.UUID) (*gormmodel.IdentityVault, error)
	CreateVaultEntry(ctx context.Context, entry *gormmodel.IdentityVault) error
	GetVaultByToken(ctx context.Context, developerID uuid.UUID, token string) (*gormmodel.IdentityVault, error)
	ListVault(ctx context.Context, developerID uuid.UUID, page, limit int) ([]gormmodel.IdentityVault, int64, error)
	DeleteVault(ctx context.Context, developerID uuid.UUID, token string) (int64, error)
	CreateShareLink(ctx context.Context, link *gormmodel.VaultShareLink) error
	GetShareLink(ctx context.Context, token string) (*gormmodel.VaultShareLink, error)
	IncrementShareAccess(ctx context.Context, id uuid.UUID) (int64, error)
	GetVaultByID(ctx context.Context, id uuid.UUID) (*gormmodel.IdentityVault, error)
	TouchVaultAccess(ctx context.Context, id uuid.UUID) error
	GetActiveCredential(ctx context.Context, developerID, verificationID uuid.UUID) (*gormmodel.VerifiableCredential, error)
	CreateCredential(ctx context.Context, credential *gormmodel.VerifiableCredential) error
	GetCredentialByJTI(ctx context.Context, jti string) (*gormmodel.VerifiableCredential, error)
	RevokeCredential(ctx context.Context, developerID uuid.UUID, jti, reason string) (int64, error)
	WebhooksByDeveloper(ctx context.Context, developerID uuid.UUID) ([]gormmodel.Webhook, error)
	WebhookByID(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.Webhook, error)
	WebhookByURL(ctx context.Context, developerID uuid.UUID, target string, sandbox bool) (*gormmodel.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *gormmodel.Webhook) error
	UpdateWebhook(ctx context.Context, developerID, id uuid.UUID, updates map[string]any) (int64, error)
	DeleteWebhook(ctx context.Context, developerID, id uuid.UUID) (int64, error)
	WebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, limit int) ([]gormmodel.WebhookDelivery, int64, error)
	CreateSchedule(ctx context.Context, schedule *gormmodel.ReverificationSchedule) error
	ListSchedules(ctx context.Context, developerID uuid.UUID, status string, page, limit int) ([]gormmodel.ReverificationSchedule, int64, error)
	GetSchedule(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.ReverificationSchedule, error)
	CancelSchedule(ctx context.Context, developerID, id uuid.UUID) (int64, error)
	ExpiringDocuments(ctx context.Context, developerID uuid.UUID, days, page, limit int) ([]gormmodel.ExpiryAlert, int64, error)
	CreateVAASUser(ctx context.Context, user *gormmodel.User) error
	ListDevelopers(ctx context.Context, search, status string, page, limit int) ([]gormmodel.Developer, int64, error)
	SetDeveloperStatus(ctx context.Context, id uuid.UUID, status string) (int64, error)
	CountAPIKeys(ctx context.Context, developerID uuid.UUID) (int64, error)
	CountVerifications(ctx context.Context, developerID uuid.UUID) (int64, error)
	DeveloperAPIKeys(ctx context.Context, developerID uuid.UUID) ([]gormmodel.APIKey, error)
	CreateDeveloperAPIKey(ctx context.Context, key *gormmodel.APIKey) error
	RevokeDeveloperAPIKey(ctx context.Context, developerID, id uuid.UUID) (int64, error)
	GetDeveloperAPIKey(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.APIKey, error)
	UpdateDeveloper(ctx context.Context, id uuid.UUID, updates map[string]any) error
	DeleteDeveloper(ctx context.Context, id uuid.UUID) (int64, error)
	CreateDeveloper(ctx context.Context, developer *gormmodel.Developer) error
	CreateReviewer(ctx context.Context, reviewer *gormmodel.VerificationReviewer) error
	Reviewers(ctx context.Context, developerID uuid.UUID) ([]gormmodel.VerificationReviewer, error)
	DeleteReviewer(ctx context.Context, developerID, id uuid.UUID) (int64, error)
	AdminVerifications(ctx context.Context, status string, developerID *uuid.UUID, page, limit int) ([]gormmodel.VerificationRequest, int64, error)
	AdminStatusCounts(ctx context.Context, since time.Time) (map[string]int64, error)
	AdminDeveloperCount(ctx context.Context) (int64, error)
	DuplicateVerifications(ctx context.Context, verification *gormmodel.VerificationRequest) ([]gormmodel.VerificationRequest, error)
	ReviewVerification(ctx context.Context, id uuid.UUID, status, reason, reviewer string) error
	ProviderMetricSummary(ctx context.Context, since time.Time) ([]map[string]any, error)
	AuditLogs(ctx context.Context, from, to time.Time, developerID *uuid.UUID, status *int32) ([]gormmodel.APIActivityLog, error)
	DeleteUserData(ctx context.Context, userID uuid.UUID) (int64, error)
	VerificationStatusCounts(ctx context.Context, developerID uuid.UUID, since time.Time) (map[string]int64, error)
	RecentActivity(ctx context.Context, developerID uuid.UUID, limit int) ([]gormmodel.APIActivityLog, error)
	VerificationByID(ctx context.Context, id uuid.UUID) (*gormmodel.VerificationRequest, error)
	CreateRuleset(ctx context.Context, ruleset *gormmodel.ComplianceRuleset) error
	ListRulesets(ctx context.Context, developerID uuid.UUID) ([]gormmodel.ComplianceRuleset, error)
	GetRuleset(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.ComplianceRuleset, error)
	UpdateRuleset(ctx context.Context, developerID, id uuid.UUID, updates map[string]any) (int64, error)
	DeleteRuleset(ctx context.Context, developerID, id uuid.UUID) (int64, error)
	RulesByRuleset(ctx context.Context, rulesetID uuid.UUID) ([]gormmodel.ComplianceRule, error)
	CreateRule(ctx context.Context, rule *gormmodel.ComplianceRule) error
	GetRule(ctx context.Context, id uuid.UUID) (*gormmodel.ComplianceRule, error)
	UpdateRule(ctx context.Context, id uuid.UUID, updates map[string]any) (int64, error)
	DeleteRule(ctx context.Context, id uuid.UUID) (int64, error)
	CreateHandoff(ctx context.Context, session *gormmodel.MobileHandoffSession) error
	GetHandoffByToken(ctx context.Context, hash string) (*gormmodel.MobileHandoffSession, error)
	UpdateHandoff(ctx context.Context, hash string, updates map[string]any) (int64, error)
	UpdatePendingHandoff(ctx context.Context, hash string, updates map[string]any) (int64, error)
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
