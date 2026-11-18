package vc

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// TestIssueJWTVC 验证 EdDSA 签名与核心声明。
func TestIssueJWTVC(t *testing.T) {
	seed := hex.EncodeToString(make([]byte, ed25519.SeedSize))
	issuer, err := New("did:web:example.test", seed, 30)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := issuer.Issue(map[string]any{"name": "JOANA SILVA COSTA"})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(credential.JWT, ".")
	if len(parts) != 3 {
		t.Fatalf("jwt parts=%d", len(parts))
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(issuer.key.Public().(ed25519.PublicKey), []byte(parts[0]+"."+parts[1]), signature) {
		t.Fatal("signature verification failed")
	}
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims["iss"] != "did:web:example.test" || claims["jti"] != credential.JTI {
		t.Fatalf("claims=%v", claims)
	}
}
