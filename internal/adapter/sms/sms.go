// Package sms 提供通用 HTTP 短信网关适配器。
package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"facial-liveness-verification-go/internal/config"
)

// Sender 通过配置的 HTTP 网关发送短信。
type Sender struct {
	cfg    config.SMS
	client *http.Client
}

// New 构造短信发送器。
func New(cfg config.SMS) *Sender {
	return &Sender{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

// Send 发送短信；网关返回 2xx 视为成功。
func (s *Sender) Send(ctx context.Context, phone, message string) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("短信未启用")
	}
	payload, _ := json.Marshal(map[string]string{
		"to": phone, "from": s.cfg.From, "message": message,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("短信网关请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("短信网关返回 %d", resp.StatusCode)
	}
	return nil
}
