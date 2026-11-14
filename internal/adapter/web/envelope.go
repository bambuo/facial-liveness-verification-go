// Package web 提供 HTTP 路由、中间件与响应信封。
package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"facial-liveness-verification-go/internal/domain"
)

type ctxKey int

const ctxRequestID ctxKey = 1

type envelope struct {
	Code      any    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

func okEnvelope(data any, requestID string) envelope {
	return envelope{Code: 0, Message: "ok", Data: data, RequestID: requestID}
}

func failEnvelope(code domain.Code, message, requestID string) envelope {
	return envelope{Code: string(code), Message: message, Data: nil, RequestID: requestID}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	raw, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := requestIDFrom(r)
	var de *domain.Error
	if !errors.As(err, &de) {
		de = domain.New(domain.CodeServiceUnavailable, "服务暂不可用")
	}
	writeJSON(w, de.Status, failEnvelope(de.Code, de.Message, requestID))
}

func writeOK(w http.ResponseWriter, r *http.Request, data any) {
	writeJSON(w, http.StatusOK, okEnvelope(data, requestIDFrom(r)))
}

func requestIDFrom(r *http.Request) string {
	if id, ok := r.Context().Value(ctxRequestID).(string); ok && id != "" {
		return id
	}
	return newRequestID()
}

func newRequestID() string {
	buf := make([]byte, 10)
	if _, err := rand.Read(buf); err != nil {
		return "req_00000000000000000000"
	}
	return "req_" + hex.EncodeToString(buf)
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxRequestID, id)
}

// formatExpires 输出与 .NET "o" 对齐的 ISO 8601。
func formatExpires(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.0000000-07:00")
}
