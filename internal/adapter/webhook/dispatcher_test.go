package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestSignPayload 验证 Webhook HMAC 签名。
func TestSignPayload(t *testing.T) {
	body := []byte(`{"event":"test"}`)
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(body)
	if got, want := signPayload("secret", body), hex.EncodeToString(mac.Sum(nil)); got != want {
		t.Fatalf("signature=%q want=%q", got, want)
	}
	if signPayload("", body) != "" {
		t.Fatal("empty secret should not sign")
	}
}

// TestDeliverRejectsLoopback 验证回环地址被阻断。
func TestDeliverRejectsLoopback(t *testing.T) {
	if _, err := New().Deliver(context.Background(), "http://127.0.0.1:9/hook", "", map[string]any{}); err == nil {
		t.Fatal("expected loopback rejection")
	}
}
