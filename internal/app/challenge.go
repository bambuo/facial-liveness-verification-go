package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// ChallengeService 负责下发一次性 challenge 与动作序列。
type ChallengeService struct {
	store    port.ChallengeStore
	sessions port.SessionRepo
	audit    *Auditor
	clock    port.Clock
	ttl      time.Duration
}

// NewChallengeService 构造 challenge 用例。
func NewChallengeService(store port.ChallengeStore, sessions port.SessionRepo, audit *Auditor, clock port.Clock, ttl time.Duration) *ChallengeService {
	return &ChallengeService{store: store, sessions: sessions, audit: audit, clock: clock, ttl: ttl}
}

// Create 创建 challenge 和对应会话。
func (s *ChallengeService) Create(ctx context.Context, auth *AuthContext, scene, mode, gallery, deviceID, requestID string) (*domain.Challenge, error) {
	if strings.TrimSpace(scene) == "" || strings.TrimSpace(mode) == "" || strings.TrimSpace(deviceID) == "" {
		return nil, domain.New(domain.CodeAuthInvalid, "scene/mode/device_id 必填")
	}
	if strings.TrimSpace(gallery) != "" {
		return nil, domain.New(domain.CodeGalleryNotFound, "1:N 检索本期未开放")
	}
	matchMode, ok := domain.ParseMatchMode(mode)
	if !ok {
		return nil, domain.New(domain.CodeAuthInvalid, "未知 mode="+mode)
	}
	now := s.clock.Now().UTC()
	actions := auth.Policy.BuildActions(newRand())
	framesPerAction := auth.Policy.FramesPerAction
	frameCount := framesPerAction * len(actions)
	if frameCount > 15 {
		frameCount = 15
	}
	ref := auth.Claims.Ref
	challenge := &domain.Challenge{
		ChallengeID:     "ch_" + randomHex(16),
		TenantID:        auth.Tenant.ID,
		ExternalSubject: auth.Claims.Subject,
		Scene:           scene,
		Ref:             ref,
		Mode:            matchMode,
		Gallery:         nil,
		DeviceID:        deviceID,
		Nonce:           randomHex(8),
		Actions:         actions,
		FramesPerAction: framesPerAction,
		FrameCount:      frameCount,
		ExpiresAt:       now.Add(300 * time.Second),
	}
	if err := s.store.Create(ctx, challenge, s.ttl); err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "challenge 存储不可用")
	}
	session := &domain.Session{
		ID:          uuid.New(),
		ChallengeID: challenge.ChallengeID,
		TenantID:    auth.Tenant.ID,
		Scene:       scene,
		Ref:         ref,
		Mode:        matchMode,
		Status:      domain.SessionChallenging,
		DeviceID:    &deviceID,
		RequestID:   &requestID,
		ExpiresAt:   challenge.ExpiresAt,
		CreatedAt:   now,
	}
	if err := s.sessions.Add(ctx, session); err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "会话写入失败")
	}
	if err := s.audit.Write(ctx, auth.Tenant.ID, auth.Claims.Subject, "challenge.created", "session:"+session.ID.String(), "ok", &requestID, map[string]any{
		"mode":        string(matchMode),
		"frame_count": frameCount,
		"actions":     actions,
	}); err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "审计写入失败")
	}
	return challenge, nil
}
