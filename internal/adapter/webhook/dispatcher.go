// Package webhook 提供带 HMAC 签名的 Webhook 投递适配器。
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Dispatcher 负责一次 Webhook HTTP 投递。
type Dispatcher struct {
	client *http.Client
}

// New 构造 Webhook 投递器；禁止重定向到其他主机，避免 SSRF 跳转。
func New() *Dispatcher {
	return &Dispatcher{client: &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

// Deliver 发送 JSON 载荷，X-Idswyft-Signature 为 HMAC-SHA256。
func (d *Dispatcher) Deliver(ctx context.Context, target, secret string, payload any) (int, error) {
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return 0, fmt.Errorf("webhook URL 无效")
	}
	if isBlockedHost(parsed.Hostname()) {
		return 0, fmt.Errorf("webhook 目标不允许")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if signature := signPayload(secret, body); signature != "" {
		req.Header.Set("X-Idswyft-Signature", signature)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("webhook 返回 %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

// signPayload 生成 HMAC-SHA256 十六进制签名；空密钥返回空串。
func signPayload(secret string, body []byte) string {
	if secret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// isBlockedHost 阻断回环、链路本地与云元数据地址；域名留给 HTTP 客户端解析。
func isBlockedHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
