package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// resendAPIEndpoint Resend 邮件发送接口
const resendAPIEndpoint = "https://api.resend.com/emails"

// ResendEmailSender 通过 Resend HTTP API 发送邮件
//
// 与 SMTPEmailSender 的区别：
//   - 不需要邮箱授权码，只需一个 Resend API Key（服务器直接 HTTP 调用）
//   - 送达率好（Resend 负责 SPF/DKIM 与投递）
//
// 配置：
//   - RESEND_API_KEY：必填，Resend 控制台生成的 key
//   - RESEND_FROM：发件人，形如 "软考学系 <no-reply@your-domain.com>"
//     未验证域名时可先用 "onboarding@resend.dev"，但只能发给 Resend 账号本人的邮箱
//
// 接口文档：https://resend.com/docs/api-reference/emails/send-email
type ResendEmailSender struct {
	apiKey   string
	from     string
	endpoint string
	client   *http.Client
}

// NewResendEmailSender 创建 Resend 发送器
func NewResendEmailSender(apiKey, from string) *ResendEmailSender {
	if strings.TrimSpace(from) == "" {
		from = "软考学系 <onboarding@resend.dev>"
	}
	return &ResendEmailSender{
		apiKey:   apiKey,
		from:     from,
		endpoint: resendAPIEndpoint,
		// 避免网络异常时请求长时间挂起，拖慢验证码接口响应
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// resendSendRequest Resend 发送邮件请求体
type resendSendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// Send 发送一封 HTML 邮件
func (s *ResendEmailSender) Send(to, subject, htmlBody string) error {
	if s.apiKey == "" {
		return errors.New("Resend API Key 未配置")
	}
	if to == "" {
		return errors.New("收件邮箱不能为空")
	}

	payload, err := json.Marshal(resendSendRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	})
	if err != nil {
		return fmt.Errorf("构造请求失败：%w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败：%w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求 Resend 失败：%w", err)
	}
	defer resp.Body.Close()

	// 限制读取长度，避免异常响应撑爆内存
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Resend 返回 %d：%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
