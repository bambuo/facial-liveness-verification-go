// Package redis 提供 challenge 一次性存储与租户分钟配额。
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"facial-liveness-verification-go/internal/domain"
)

// ChallengeStore 使用 SET NX + TTL 保存，GETDEL 原子消费。
type ChallengeStore struct {
	client *goredis.Client
	ttl    time.Duration
}

// NewChallengeStore 构造 challenge 存储。
func NewChallengeStore(client *goredis.Client, ttl time.Duration) *ChallengeStore {
	return &ChallengeStore{client: client, ttl: ttl}
}

// Create 写入 challenge；ID 冲突时返回错误。
func (s *ChallengeStore) Create(ctx context.Context, c *domain.Challenge, ttl time.Duration) error {
	payload, err := json.Marshal(c)
	if err != nil {
		return err
	}
	ok, err := s.client.SetNX(ctx, key(c.ChallengeID), payload, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("challenge_id 冲突")
	}
	return nil
}

// Consume 原子取删；不存在或已过期返回 nil。
func (s *ChallengeStore) Consume(ctx context.Context, challengeID string) (*domain.Challenge, error) {
	payload, err := s.client.GetDel(ctx, key(challengeID)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var c domain.Challenge
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("challenge 数据损坏: %w", err)
	}
	return &c, nil
}

func key(id string) string { return "face:challenge:" + id }

// QuotaStore 按租户与分钟分桶计数。
type QuotaStore struct {
	client *goredis.Client
}

// NewQuotaStore 构造租户配额存储。
func NewQuotaStore(client *goredis.Client) *QuotaStore { return &QuotaStore{client: client} }

// TryIncrement 计数加一并判断是否仍在配额内。
func (s *QuotaStore) TryIncrement(ctx context.Context, tenantID uuid.UUID, limit int) (bool, error) {
	bucket := time.Now().UTC().Format("200601021504")
	k := "face:quota:" + tenantID.String() + ":" + bucket
	n, err := s.client.Incr(ctx, k).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = s.client.Expire(ctx, k, 70*time.Second).Err()
	}
	return n <= int64(limit), nil
}
