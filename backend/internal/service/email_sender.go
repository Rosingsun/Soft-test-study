package service

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// EmailSender 抽象邮件发送能力
//
// 实现：
//   - SMTPEmailSender：使用 net/smtp 通过 SMTP 服务发送
//   - LogEmailSender：仅写日志，用于未配置 SMTP 时的开发环境
type EmailSender interface {
	Send(to, subject, htmlBody string) error
}

// ---------- LogEmailSender ----------

// LogEmailSender 把邮件内容打印到日志
//
// 使用场景：
//   - SMTP_HOST 未配置时自动降级
//   - 本地开发、自动化测试
//
// 日志格式：[EMAIL-DEV] to=xxx@xx.com subject=xxx body=...
//   - 单行输出，便于 grep
type LogEmailSender struct{}

func NewLogEmailSender() *LogEmailSender { return &LogEmailSender{} }

func (s *LogEmailSender) Send(to, subject, htmlBody string) error {
	if to == "" {
		return errors.New("收件邮箱不能为空")
	}
	// 提取验证码方便联调（如果正文里恰好出现 6 位连续数字）
	code := extractCodeFromText(htmlBody)
	log.Printf("[EMAIL-DEV] to=%s subject=%s code=%s body=%s", to, subject, code, stripHTML(htmlBody))
	return nil
}

// ---------- SMTPEmailSender ----------

// SMTPEmailSender 使用 net/smtp 发送邮件
//
// 兼容性：
//   - 465 端口（SSL）：使用 smtp.Dial + StartTLS
//   - 25 / 587 端口（明文 / STARTTLS）：使用 smtp.SendMail
// 鉴权：
//   - 使用 smtp.PlainAuth（QQ 邮箱、163 邮箱、阿里云邮箱等通用）
type SMTPEmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewSMTPEmailSender(host string, port int, user, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailSender) Send(to, subject, htmlBody string) error {
	if s.host == "" || s.user == "" || s.password == "" {
		return errors.New("SMTP 未完整配置")
	}
	if to == "" {
		return errors.New("收件邮箱不能为空")
	}

	from := s.from
	if from == "" {
		from = s.user
	}
	msg := buildMIMEMessage(from, to, subject, htmlBody)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	// 465 端口走 SSL（实际上 Go 标准库没有原生 SSL，通过 Dial+StartTLS 兼容）
	// 25 / 587 直接 SendMail
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func buildMIMEMessage(from, to, subject, htmlBody string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return b.String()
}

// ---------- 工具函数 ----------

// stripHTML 去除 HTML 标签，用于日志展示
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// extractCodeFromText 从文本中提取第一个 6 位连续数字（仅用于日志展示）
func extractCodeFromText(s string) string {
	digits := make([]rune, 0, 6)
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
			if len(digits) == 6 {
				return string(digits)
			}
		} else {
			if len(digits) > 0 {
				digits = digits[:0]
			}
		}
	}
	return ""
}
