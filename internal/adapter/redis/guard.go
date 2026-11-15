package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// LivenessGuard 使用 Redis 计数连续活体失败，并在达到上限时短时锁定。
type LivenessGuard struct {
	client  *goredis.Client
	limit   int
	window  time.Duration
	lockTTL time.Duration
}

// NewLivenessGuard 构造活体失败锁定器。
func NewLivenessGuard(client *goredis.Client, limit, windowSeconds, lockSeconds int) *LivenessGuard {
	if limit <= 0 {
		limit = 5
	}
	return &LivenessGuard{client: client, limit: limit, window: time.Duration(windowSeconds) * time.Second, lockTTL: time.Duration(lockSeconds) * time.Second}
}

// IsLocked 查询主体与设备是否处于锁定状态。
func (g *LivenessGuard) IsLocked(ctx context.Context, tenantID uuid.UUID, subject, device string) (bool, error) {
	n, err := g.client.Exists(ctx, g.lockKey(tenantID, subject, device)).Result()
	return n > 0, err
}

// RecordFailure 记录一次活体失败；达到上限时写入锁定键。
func (g *LivenessGuard) RecordFailure(ctx context.Context, tenantID uuid.UUID, subject, device string) (bool, error) {
	key := g.countKey(tenantID, subject, device)
	n, err := g.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = g.client.Expire(ctx, key, g.window).Err()
	}
	if int(n) < g.limit {
		return false, nil
	}
	if err := g.client.Set(ctx, g.lockKey(tenantID, subject, device), "1", g.lockTTL).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// Reset 清除连续失败计数。
func (g *LivenessGuard) Reset(ctx context.Context, tenantID uuid.UUID, subject, device string) error {
	return g.client.Del(ctx, g.countKey(tenantID, subject, device)).Err()
}

func (g *LivenessGuard) lockKey(tenantID uuid.UUID, subject, device string) string {
	return "face:liveness:lock:" + guardHash(tenantID, subject, device)
}

func (g *LivenessGuard) countKey(tenantID uuid.UUID, subject, device string) string {
	return "face:liveness:fail:" + guardHash(tenantID, subject, device)
}

func guardHash(tenantID uuid.UUID, subject, device string) string {
	sum := sha256.Sum256([]byte(tenantID.String() + "|" + subject + "|" + device))
	return hex.EncodeToString(sum[:])
}
