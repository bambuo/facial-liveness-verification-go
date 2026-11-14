package sec

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

const audience = "face-service"

// Validator 校验端上 JWT：HS256、aud=face-service、exp/nbf，容差 30 秒。
type Validator struct {
	key []byte
}

// NewValidator 解析 Base64 端上签名密钥。
func NewValidator(base64Key string) (*Validator, error) {
	if base64Key == "" {
		return nil, domain.New(domain.CodeServiceUnavailable, "未配置 AUTH_DEV_SIGNING_KEY")
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "AUTH_DEV_SIGNING_KEY 必须是 Base64")
	}
	return &Validator{key: key}, nil
}

// ValidateEndpoint 验签并解析 claims；失败返回 AUTH_INVALID。
func (v *Validator) ValidateEndpoint(tokenString string) (domain.EndpointClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return v.key, nil
	}, jwt.WithAudience(audience), jwt.WithExpirationRequired(), jwt.WithLeeway(30*time.Second))
	if err != nil || !token.Valid {
		return domain.EndpointClaims{}, domain.Of(domain.CodeAuthInvalid)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return domain.EndpointClaims{}, domain.Of(domain.CodeAuthInvalid)
	}
	tenant, ok1 := claimString(claims, "tenant")
	sub, ok2 := claimString(claims, "sub")
	scene, ok3 := claimString(claims, "scene")
	jti, ok4 := claimString(claims, "jti")
	if !ok1 || !ok2 || !ok3 || !ok4 || tenant == "" || sub == "" || scene == "" || jti == "" {
		return domain.EndpointClaims{}, domain.New(domain.CodeAuthInvalid, "缺少 tenant/sub/scene/jti claim")
	}
	ops := parseOps(claims["ops"])
	if len(ops) == 0 {
		return domain.EndpointClaims{}, domain.New(domain.CodeAuthInvalid, "缺少 ops claim")
	}
	return domain.EndpointClaims{
		Tenant:  tenant,
		Subject: sub,
		Scene:   scene,
		Ref:     optionalString(claims, "ref"),
		Ops:     ops,
		Gallery: optionalString(claims, "gallery"),
		Device:  optionalString(claims, "device"),
		Jti:     jti,
	}, nil
}

// Issuer 签发结果令牌与 dev 令牌。
type Issuer struct {
	endpointKey []byte
	resultKey   []byte
	ttl         time.Duration
	clock       port.Clock
}

// NewIssuer 构造令牌签发器；结果令牌密钥按 UTF-8 原文使用。
func NewIssuer(endpointBase64Key, resultKey string, ttlSeconds int, clock port.Clock) (*Issuer, error) {
	key, err := base64.StdEncoding.DecodeString(endpointBase64Key)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "AUTH_DEV_SIGNING_KEY 必须是 Base64")
	}
	return &Issuer{
		endpointKey: key,
		resultKey:   []byte(resultKey),
		ttl:         time.Duration(ttlSeconds) * time.Second,
		clock:       clock,
	}, nil
}

// IssueDev 签发固定 tenant=dev 的 5 分钟令牌。
func (i *Issuer) IssueDev(sub string) (string, time.Time, error) {
	if sub == "" {
		sub = "dev-user"
	}
	now := i.clock.Now()
	expires := now.Add(i.ttl)
	claims := jwt.MapClaims{
		"iss":    "dev",
		"aud":    audience,
		"tenant": "dev",
		"sub":    sub,
		"scene":  "kyc",
		"ops":    []string{"match", "liveness", "enroll", "verify", "identify", "document", "delete"},
		"jti":    randomID(),
		"exp":    expires.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.endpointKey)
	if err != nil {
		return "", time.Time{}, domain.New(domain.CodeServiceUnavailable, "签发 dev 令牌失败")
	}
	return signed, expires, nil
}

// IssueResult 签发结果令牌：iss=face-service，aud=租户 code。
func (i *Issuer) IssueResult(p domain.ResultTokenPayload) (string, error) {
	if len(i.resultKey) == 0 {
		return "", domain.New(domain.CodeServiceUnavailable, "未配置 AUTH_RESULT_TOKEN_SIGNING_KEY")
	}
	claims := jwt.MapClaims{
		"iss":    "face-service",
		"aud":    p.TenantCode,
		"tenant": p.TenantCode,
		"sub":    p.Subject,
		"scene":  p.Scene,
		"result": p.Result,
		"jti":    p.Jti,
		"exp":    i.clock.Now().Add(i.ttl).Unix(),
	}
	if p.Ref != nil && *p.Ref != "" {
		claims["ref"] = *p.Ref
	}
	if p.Liveness != nil {
		claims["liveness"] = string(*p.Liveness)
	}
	if p.Score != nil {
		claims["score"] = domain.RoundAway(*p.Score, 5)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.resultKey)
	if err != nil {
		return "", domain.New(domain.CodeServiceUnavailable, "签发结果令牌失败")
	}
	return signed, nil
}

func claimString(claims jwt.MapClaims, key string) (string, bool) {
	v, ok := claims[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func optionalString(claims jwt.MapClaims, key string) *string {
	if s, ok := claimString(claims, key); ok && s != "" {
		return &s
	}
	return nil
}

func parseOps(raw any) []string {
	switch v := raw.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	case string:
		return strings.Fields(v)
	default:
		return nil
	}
}

func randomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
