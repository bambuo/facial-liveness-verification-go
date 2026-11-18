package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// APIKeyAuth 是 API Key 鉴权所需的最小字段。
type APIKeyAuth struct {
	ID              uuid.UUID
	DeveloperID     uuid.UUID
	IsSandbox       bool
	IsService       bool
	ExpiresAt       *time.Time
	DeveloperStatus string
}

// HandoffAuth 是 handoff token 解析结果。
type HandoffAuth struct {
	APIKey         APIKeyAuth
	ExpiresAt      time.Time
	Status         string
	VerificationID *string
}

// SessionAuth 是 session token 解析结果。
type SessionAuth struct {
	VerificationID uuid.UUID
	DeveloperID    uuid.UUID
	APIKeyID       *uuid.UUID
	ExpiresAt      *time.Time
}

// WebhookTarget 是开发者 Webhook 投递目标。
type WebhookTarget struct {
	ID     uuid.UUID
	URL    string
	Secret string
}

// IdswyftAuthStore 提供 idswyft 认证矩阵所需的查询。
type IdswyftAuthStore interface {
	FindAPIKeyByHash(ctx context.Context, hash string) (*APIKeyAuth, error)
	FindHandoffByToken(ctx context.Context, hash string) (*HandoffAuth, error)
	FindSessionByToken(ctx context.Context, hash string) (*SessionAuth, error)
	TouchAPIKey(ctx context.Context, id uuid.UUID) error
}
