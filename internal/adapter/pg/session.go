package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"facial-liveness-verification-go/internal/domain"
)

const sessionColumns = "id::text, challenge_id, tenant_id::text, subject_id::text, scene, ref, mode, status, score, liveness_status, clarity, reality, device_id, request_id, expires_at, created_at, completed_at"

// SessionRepo 是会话仓储适配器。
type SessionRepo struct{ r *Repo }

// Add 新增会话。
func (sr *SessionRepo) Add(ctx context.Context, s *domain.Session) error {
	_, err := sr.r.Pool.Exec(ctx, `
		INSERT INTO face_verification_sessions
		(id, challenge_id, tenant_id, subject_id, scene, ref, mode, status, score, liveness_status, clarity, reality, device_id, request_id, expires_at, created_at, completed_at)
		VALUES ($1::uuid,$2,$3::uuid,$4::uuid,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		s.ID.String(), s.ChallengeID, s.TenantID.String(), uuidPtr(s.SubjectID), s.Scene, s.Ref, string(s.Mode), string(s.Status),
		s.Score, livenessPtr(s.Liveness), s.Clarity, s.Reality, s.DeviceID, s.RequestID, s.ExpiresAt, s.CreatedAt, s.CompletedAt)
	return err
}

// GetByChallengeID 按 challenge 查询会话。
func (sr *SessionRepo) GetByChallengeID(ctx context.Context, tenantID uuid.UUID, challengeID string) (*domain.Session, error) {
	row := sr.r.Pool.QueryRow(ctx, "SELECT "+sessionColumns+" FROM face_verification_sessions WHERE tenant_id=$1::uuid AND challenge_id=$2", tenantID.String(), challengeID)
	return scanSession(row)
}

// Update 更新会话。
func (sr *SessionRepo) Update(ctx context.Context, s *domain.Session) error {
	_, err := sr.r.Pool.Exec(ctx, `
		UPDATE face_verification_sessions SET subject_id=$3::uuid, status=$4, score=$5, liveness_status=$6,
		       clarity=$7, reality=$8, device_id=$9, request_id=$10, expires_at=$11, completed_at=$12
		WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		s.TenantID.String(), s.ID.String(), uuidPtr(s.SubjectID), string(s.Status), s.Score, livenessPtr(s.Liveness),
		s.Clarity, s.Reality, s.DeviceID, s.RequestID, s.ExpiresAt, s.CompletedAt)
	return err
}

// ListExpired 列出超时且未完成的会话。
func (sr *SessionRepo) ListExpired(ctx context.Context, now time.Time, limit int) ([]*domain.Session, error) {
	rows, err := sr.r.Pool.Query(ctx,
		"SELECT "+sessionColumns+" FROM face_verification_sessions WHERE expires_at < $1 AND status NOT IN ('expired','passed','failed') ORDER BY expires_at LIMIT $2",
		now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanSession(row scanner) (*domain.Session, error) {
	var (
		id, challenge, tenant, scene, mode, status string
		subject, ref, liveness, device, requestID  *string
		score, clarity, reality                    *float64
		expires, created                           time.Time
		completed                                  *time.Time
	)
	err := row.Scan(&id, &challenge, &tenant, &subject, &scene, &ref, &mode, &status, &score, &liveness,
		&clarity, &reality, &device, &requestID, &expires, &created, &completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	tid, err := uuid.Parse(tenant)
	if err != nil {
		return nil, err
	}
	s := &domain.Session{
		ID: uid, ChallengeID: challenge, TenantID: tid, Scene: scene, Ref: ref,
		Mode: domain.MatchMode(mode), Status: domain.SessionStatus(status),
		Score: score, Clarity: clarity, Reality: reality, DeviceID: device, RequestID: requestID,
		ExpiresAt: expires, CreatedAt: created, CompletedAt: completed,
	}
	if subject != nil && *subject != "" {
		sid, err := uuid.Parse(*subject)
		if err != nil {
			return nil, err
		}
		s.SubjectID = &sid
	}
	if liveness != nil {
		lv := domain.LivenessStatus(*liveness)
		s.Liveness = &lv
	}
	return s, nil
}

func uuidPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func livenessPtr(v *domain.LivenessStatus) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}
