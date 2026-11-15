// Package app 编排人脸核验用例，只依赖领域层与端口。
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	mrand "math/rand"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// AuthContext 是一次端上请求的认证结果。
type AuthContext struct {
	Tenant *domain.Tenant
	Claims domain.EndpointClaims
	Policy domain.Policy
}

// Authenticator 执行签名、租户与配额校验。
type Authenticator struct {
	validator port.TokenValidator
	tenants   port.TenantRepo
	quota     port.QuotaStore
	base      domain.Policy
}

// NewAuthenticator 构造认证链。
func NewAuthenticator(v port.TokenValidator, tenants port.TenantRepo, quota port.QuotaStore, base domain.Policy) *Authenticator {
	return &Authenticator{validator: v, tenants: tenants, quota: quota, base: base}
}

// Authenticate 校验 JWT、解析租户并执行分钟配额。
func (a *Authenticator) Authenticate(ctx context.Context, bearer string) (*AuthContext, error) {
	claims, err := a.validator.ValidateEndpoint(bearer)
	if err != nil {
		return nil, err
	}
	var tenant *domain.Tenant
	if id, parseErr := uuid.Parse(claims.Tenant); parseErr == nil {
		tenant, err = a.tenants.GetByID(ctx, id)
	} else {
		tenant, err = a.tenants.GetByCode(ctx, claims.Tenant)
	}
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "租户查询失败")
	}
	if tenant == nil || tenant.Status != domain.TenantActive {
		return nil, domain.Of(domain.CodeTenantDisabled)
	}
	policy, err := domain.ApplyPolicy(a.base, tenant.PolicyJSON)
	if err != nil {
		slog.Warn("租户策略配置非法", "tenant", tenant.Code, "err", err)
		return nil, domain.New(domain.CodeServiceUnavailable, "租户策略配置非法")
	}
	ok, err := a.quota.TryIncrement(ctx, tenant.ID, policy.MaxRequestsPerMinute)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "配额服务不可用")
	}
	if !ok {
		return nil, domain.Of(domain.CodeQuotaExceeded)
	}
	return &AuthContext{Tenant: tenant, Claims: claims, Policy: policy}, nil
}

// newToken 生成 16 字节随机串的十六进制形式。
func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

// newRand 返回仅用于动作洗牌的伪随机源。
func newRand() *mrand.Rand { return mrand.New(mrand.NewSource(mrand.Int63())) }
