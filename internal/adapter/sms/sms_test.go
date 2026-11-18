package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"facial-liveness-verification-go/internal/config"
)

// TestSendPostsMessage 验证短信网关请求体与鉴权头。
func TestSendPostsMessage(t *testing.T) {
	var gotAuth string
	var gotPhone string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotPhone = body["to"]
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	sender := New(config.SMS{Enabled: true, Endpoint: server.URL, APIKey: "key", From: "idswyft"})
	if err := sender.Send(context.Background(), "+15551234567", "code 123456"); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer key" || gotPhone != "+15551234567" {
		t.Fatalf("auth=%q phone=%q", gotAuth, gotPhone)
	}
}
