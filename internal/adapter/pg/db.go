// Package pg 提供 PostgreSQL 仓储。
package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo 持有 pgx 连接池并实现全部仓储与索引端口。
type Repo struct {
	Pool *pgxpool.Pool
}

// New 创建连接池并验证连通性。
func New(ctx context.Context, url string) (*Repo, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Repo{Pool: pool}, nil
}

// Close 关闭连接池。
func (r *Repo) Close() { r.Pool.Close() }

// Tenants 返回租户仓储。
func (r *Repo) Tenants() *TenantRepo { return &TenantRepo{r: r} }

// Subjects 返回主体仓储。
func (r *Repo) Subjects() *SubjectRepo { return &SubjectRepo{r: r} }

// Profiles 返回档案仓储。
func (r *Repo) Profiles() *ProfileRepo { return &ProfileRepo{r: r} }

// Sessions 返回会话仓储。
func (r *Repo) Sessions() *SessionRepo { return &SessionRepo{r: r} }

// Audits 返回审计仓储。
func (r *Repo) Audits() *AuditRepo { return &AuditRepo{r: r} }

// Documents 返回证件记录仓储。
func (r *Repo) Documents() *DocumentRepo { return &DocumentRepo{r: r} }

// Keys 返回租户密钥仓储。
func (r *Repo) Keys() *KeyRepo { return &KeyRepo{r: r} }
