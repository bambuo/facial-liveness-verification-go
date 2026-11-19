package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDocsMarkdown 校验 Markdown 端点返回可渲染文本且不泄露参考品牌。
func TestDocsMarkdown(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).handleDocsMarkdown(rec, httptest.NewRequest("GET", "/api/docs/markdown", nil))

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/markdown") {
		t.Fatalf("content-type = %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# Identity Verification API Documentation") {
		t.Fatal("missing markdown title")
	}
	if strings.Contains(strings.ToLower(body), "idswyft") {
		t.Fatal("reference brand leaked into markdown")
	}
}
