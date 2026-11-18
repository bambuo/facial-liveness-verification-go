package smtp

import (
	"strings"
	"testing"

	"facial-liveness-verification-go/internal/config"
)

// TestSendRejectsHeaderInjection 验证邮件头换行被阻断。
func TestSendRejectsHeaderInjection(t *testing.T) {
	sender := New(config.SMTP{Enabled: true, Host: "smtp.example.com", Port: 587, From: "a@example.com"})
	if err := sender.Send("b@example.com", "subject\r\nBcc: c@example.com", "body"); err == nil || !strings.Contains(err.Error(), "换行") {
		t.Fatalf("err=%v", err)
	}
}

// TestSendDisabled 验证未启用时明确失败。
func TestSendDisabled(t *testing.T) {
	if err := New(config.SMTP{}).Send("b@example.com", "subject", "body"); err == nil {
		t.Fatal("expected disabled error")
	}
}
