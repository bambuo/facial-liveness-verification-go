// Package vc 提供 W3C JWT-VC 的 Ed25519 签发。
package vc

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Issuer 使用 Ed25519 seed 签发 JWT-VC。
type Issuer struct {
	did     string
	key     ed25519.PrivateKey
	ttlDays int
}

// New 构造签发器；privateKeyHex 必须是 64 位十六进制 seed。
func New(did, privateKeyHex string, ttlDays int) (*Issuer, error) {
	seed, err := hex.DecodeString(privateKeyHex)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("VC_ISSUER_PRIVATE_KEY 必须是 %d 位十六进制 seed", ed25519.SeedSize*2)
	}
	if ttlDays <= 0 {
		ttlDays = 730
	}
	return &Issuer{did: did, key: ed25519.NewKeyFromSeed(seed), ttlDays: ttlDays}, nil
}

// PublicKeyJWK 返回 DID 文档使用的 Ed25519 公钥 JWK。
func (i *Issuer) PublicKeyJWK() map[string]string {
	publicKey := i.key.Public().(ed25519.PublicKey)
	return map[string]string{
		"kty": "OKP", "crv": "Ed25519",
		"x": base64.RawURLEncoding.EncodeToString(publicKey),
	}
}

// Credential 是一次签发结果的引用。
type Credential struct {
	JWT       string
	JTI       string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Issue 为已验证主体签发 JWT-VC。
func (i *Issuer) Issue(subject map[string]any) (Credential, error) {
	if subject == nil {
		return Credential{}, fmt.Errorf("缺少 credentialSubject")
	}
	now := time.Now().UTC()
	expires := now.AddDate(0, 0, i.ttlDays)
	jti := "urn:uuid:" + uuid.NewString()
	header := map[string]any{"alg": "EdDSA", "typ": "JWT"}
	payload := map[string]any{
		"iss": i.did,
		"jti": jti,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": expires.Unix(),
		"vc": map[string]any{
			"@context":          []string{"https://www.w3.org/2018/credentials/v1"},
			"type":              []string{"VerifiableCredential", "IdentityCredential"},
			"credentialSubject": subject,
		},
	}
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)
	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := ed25519.Sign(i.key, []byte(signingInput))
	return Credential{
		JWT: signingInput + "." + base64.RawURLEncoding.EncodeToString(signature),
		JTI: jti, IssuedAt: now, ExpiresAt: expires,
	}, nil
}
