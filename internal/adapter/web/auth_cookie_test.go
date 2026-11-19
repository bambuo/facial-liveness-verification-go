package web

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/config"
)

// TestDeveloperSessionCookie 校验开发者 Bearer token 可同时用于 Cookie 会话。
func TestDeveloperSessionCookie(t *testing.T) {
	s := &Server{cfg: &config.Config{Security: config.Security{APIKeySecret: "test-secret"}}}
	developerID := uuid.New()
	token, err := s.signDeveloperToken(developerID, "dev@example.com")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	principal, ok := s.authenticateDeveloperTokenValue(rec, httptest.NewRequest("GET", "/", nil), token)
	if !ok || principal.DeveloperID != developerID {
		t.Fatalf("cookie token rejected: ok=%v principal=%+v", ok, principal)
	}
	setAuthCookie(rec, httptest.NewRequest("GET", "/", nil), developerSessionCookie, token, 3600)
	header := rec.Header().Get("Set-Cookie")
	if !strings.Contains(header, developerSessionCookie+"=") || !strings.Contains(header, "HttpOnly") {
		t.Fatalf("unexpected cookie: %q", header)
	}
}

// TestNormalizedAdminStats 校验审核台统计键与状态折叠。
func TestNormalizedAdminStats(t *testing.T) {
	stats := normalizedAdminStats(map[string]int64{
		"AWAITING_FRONT": 2,
		"COMPLETE":       3,
		"HARD_REJECTED":  1,
		"manual_review":  4,
	})
	if stats["total"] != 10 || stats["pending"] != 2 || stats["verified"] != 3 || stats["failed"] != 1 || stats["manual_review"] != 4 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

// TestGitHubURLDisabled 校验未配置 OAuth 时返回可消费的禁用状态。
func TestGitHubURLDisabled(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).handleGitHubURL(rec, httptest.NewRequest("GET", "/api/auth/developer/github/url", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"configured":false`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
