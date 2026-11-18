// Package smtp 提供标准 SMTP 邮件发送适配器。
package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"facial-liveness-verification-go/internal/config"
)

// Sender 使用标准 SMTP 发送纯文本或 HTML 邮件。
type Sender struct {
	cfg config.SMTP
}

// New 构造 SMTP 发送器；未启用时 Send 返回明确错误。
func New(cfg config.SMTP) *Sender { return &Sender{cfg: cfg} }

// Send 发送一封邮件。
func (s *Sender) Send(to, subject, body string) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("SMTP 未启用")
	}
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("邮件头包含非法换行")
	}
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprint(s.cfg.Port))
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP 连接失败: %w", err)
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("SMTP STARTTLS 失败: %w", err)
		}
	}
	if s.cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}
	if err := client.Mail(s.cfg.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	message := "From: " + s.cfg.From + "\r\nTo: " + to + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body + "\r\n"
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
