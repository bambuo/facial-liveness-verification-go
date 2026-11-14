package sec

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"facial-liveness-verification-go/internal/domain"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestAESGCMRoundtripAndTamper(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plain := []byte("特征向量数据")
	sealed, err := AESEncrypt(plain, key)
	if err != nil {
		t.Fatal(err)
	}
	if len(sealed) != 12+len(plain)+16 {
		t.Fatalf("cipher length = %d", len(sealed))
	}
	got, err := AESDecrypt(sealed, key)
	if err != nil || string(got) != string(plain) {
		t.Fatalf("roundtrip failed: %v", err)
	}
	sealed[len(sealed)-1] ^= 0xFF
	if _, err := AESDecrypt(sealed, key); err == nil {
		t.Fatal("tampered cipher should fail")
	}
}

func TestAESGCMShortKeyDerivation(t *testing.T) {
	key := []byte("short-key")
	plain := []byte("payload")
	sealed, err := AESEncrypt(plain, key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := AESDecrypt(sealed, key)
	if err != nil || string(got) != string(plain) {
		t.Fatalf("short key roundtrip failed: %v", err)
	}
}

func TestDevTokenIssueAndValidate(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	clk := fixedClock{t: time.Now().UTC()}
	issuer, err := NewIssuer(key, "result-key", 300, clk)
	if err != nil {
		t.Fatal(err)
	}
	validator, err := NewValidator(key)
	if err != nil {
		t.Fatal(err)
	}
	token, expires, err := issuer.IssueDev("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if expires.Before(clk.Now()) {
		t.Fatal("expires in the past")
	}
	claims, err := validator.ValidateEndpoint(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Tenant != "dev" || claims.Subject != "user-1" || claims.Scene != "kyc" || len(claims.Ops) == 0 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if _, err := validator.ValidateEndpoint(token[:len(token)-2] + "zz"); err == nil {
		t.Fatal("tampered token should fail")
	}
}

func TestResultTokenClaims(t *testing.T) {
	endpointKey := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	resultKey := "result-signing-key"
	clk := fixedClock{t: time.Now().UTC()}
	issuer, err := NewIssuer(endpointKey, resultKey, 300, clk)
	if err != nil {
		t.Fatal(err)
	}
	score := 0.81234
	liveness := domain.LivenessReal
	ref := "order-1"
	token, err := issuer.IssueResult(domain.ResultTokenPayload{
		TenantCode: "dev",
		Subject:    "user-1",
		Scene:      "kyc",
		Ref:        &ref,
		Result:     "passed",
		Liveness:   &liveness,
		Score:      &score,
		Jti:        "abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) { return []byte(resultKey), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("result token invalid: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["iss"] != "face-service" || claims["aud"] != "dev" || claims["result"] != "passed" || claims["ref"] != "order-1" {
		t.Fatalf("unexpected result claims: %v", claims)
	}
}
