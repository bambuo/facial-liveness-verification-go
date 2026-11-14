package app

import (
	"context"
	"time"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// CleanupService 执行会话过期与留存清理。
type CleanupService struct {
	sessions    port.SessionRepo
	documents   port.DocumentRepo
	profiles    port.ProfileRepo
	clock       port.Clock
	profileDays int
}

// NewCleanupService 构造清理用例。
func NewCleanupService(sessions port.SessionRepo, documents port.DocumentRepo, profiles port.ProfileRepo, clock port.Clock, profileDays int) *CleanupService {
	return &CleanupService{sessions: sessions, documents: documents, profiles: profiles, clock: clock, profileDays: profileDays}
}

// ExpireSessions 将超时未完成会话标记为 expired。
func (s *CleanupService) ExpireSessions(ctx context.Context) (int, error) {
	expired, err := s.sessions.ListExpired(ctx, s.clock.Now().UTC(), 200)
	if err != nil {
		return 0, err
	}
	now := s.clock.Now().UTC()
	for _, session := range expired {
		session.Status = domain.SessionExpired
		session.CompletedAt = &now
		if err := s.sessions.Update(ctx, session); err != nil {
			return 0, err
		}
	}
	return len(expired), nil
}

// Retention 删除过期证件记录并硬删超保留期档案。
func (s *CleanupService) Retention(ctx context.Context) (docs int, profiles int, err error) {
	now := s.clock.Now().UTC()
	expiredDocs, err := s.documents.ListExpired(ctx, now, 200)
	if err != nil {
		return 0, 0, err
	}
	for _, doc := range expiredDocs {
		if err := s.documents.Delete(ctx, doc.ID); err != nil {
			return 0, 0, err
		}
	}
	before := now.Add(-time.Duration(s.profileDays) * 24 * time.Hour)
	expiredProfiles, err := s.profiles.ListDeletedExpired(ctx, before, 200)
	if err != nil {
		return len(expiredDocs), 0, err
	}
	for _, profile := range expiredProfiles {
		if err := s.profiles.Delete(ctx, profile.ID); err != nil {
			return len(expiredDocs), 0, err
		}
	}
	return len(expiredDocs), len(expiredProfiles), nil
}
