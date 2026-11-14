// Package seed 提供 Development 专用的幂等种子数据。
package seed

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureDevTenant 确保 code=dev 的 active 租户存在。
func EnsureDevTenant(ctx context.Context, pool *pgxpool.Pool) error {
	var id string
	err := pool.QueryRow(ctx, "SELECT id::text FROM face_tenants WHERE code='dev'").Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		INSERT INTO face_tenants (id, code, name, status, policy_json, created_at, updated_at)
		VALUES ($1::uuid,'dev','本地开发租户','active','{}'::jsonb,$2,$2)
		ON CONFLICT (code) DO NOTHING`,
		uuid.New().String(), now)
	return err
}
