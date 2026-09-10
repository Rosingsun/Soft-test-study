package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
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
	host      string
	port      int
	user      string
	password  string
	fromEmail string
	fromName  string
}

// NewSMTPEmailSender 创建 SMTP 发送器。
//   - fromEmail：信封发件人，必须是合法邮箱地址（缺省回退到 user）；QQ 等要求与登录账号一致
//   - fromName：MIME 头部显示名（如“软考学系”），可选，会自动 RFC2047 编码
func NewSMTPEmailSender(host string, port int, user, password, fromEmail, fromName string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:      host,
		port:      port,
		user:      user,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (s *SMTPEmailSender) Send(to, subject, htmlBody string) error {
	if s.host == "" || s.user == "" || s.password == "" {
		return errors.New("SMTP 未完整配置")
	}
	if to == "" {
		return errors.New("收件邮箱不能为空")
	}

	// 信封发件人（MAIL FROM）必须是合法邮箱地址，缺省回退到登录账号。
	// 之前把显示名“软考学系”当作信封发件人发给 QQ，导致 502 Invalid input。
	envelopeFrom := s.fromEmail
	if envelopeFrom == "" {
		envelopeFrom = s.user
	}
	// MIME 头部 From：显示名 + 邮箱，非 ASCII 显示名由 mail.Address 自动 RFC2047 编码
	mailFrom := (&mail.Address{Name: s.fromName, Address: envelopeFrom}).String()
	msg := buildMIMEMessage(mailFrom, to, subject, htmlBody)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	// 465 端口为 implicit TLS（SMTPS），标准库 smtp.SendMail 不支持，需手动 TLS 握手；
	// 25 / 587 走 smtp.SendMail（587 会自动协商 STARTTLS）
	if s.port == 465 {
		return sendMailSSL(addr, s.host, auth, envelopeFrom, to, []byte(msg))
	}
	return sendMailPlain(addr, auth, envelopeFrom, []string{to}, []byte(msg))
}

// sendMailSSL 通过 implicit TLS（SMTPS，通常为 465 端口）发送邮件
//
// 强制使用 IPv4 连接：QQ 邮箱等国内 SMTP 服务器对 IPv6 来源的发信会返回
// "502 Invalid input"，因此 dialer 限制网络层为 IPv4。
func sendMailSSL(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	dialer := &net.Dialer{}
	rawConn, err := dialer.Dial("tcp4", addr)
	if err != nil {
		return fmt.Errorf("TCP 连接失败：%w", err)
	}
	tlsConfig := &tls.Config{ServerName: host}
	conn := tls.Client(rawConn, tlsConfig)
	if err := conn.Handshake(); err != nil {
		conn.Close()
		return fmt.Errorf("TLS 握手失败：%w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败：%w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败：%w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM 失败：%w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO 失败：%w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA 失败：%w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP 写入正文失败：%w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP 提交正文失败：%w", err)
	}
	return client.Quit()
}

// sendMailPlain 通过 25/587 端口发送邮件（强制 IPv4）
//
// 与 smtp.SendMail 功能一致，但 dialer 限制为 tcp4，
// 避免 IPv6 来源被 QQ/163 等 SMTP 拒绝（502 Invalid input）。
func sendMailPlain(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	dialer := &net.Dialer{}
	conn, err := dialer.Dial("tcp4", addr)
	if err != nil {
		return fmt.Errorf("TCP 连接失败：%w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败：%w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: strings.Split(addr, ":")[0]}); err != nil {
			return fmt.Errorf("STARTTLS 失败：%w", err)
		}
	}
	if a != nil {
		if err := c.Auth(a); err != nil {
			return fmt.Errorf("SMTP 认证失败：%w", err)
		}
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM 失败：%w", err)
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP RCPT TO 失败：%w", err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA 失败：%w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("SMTP 写入正文失败：%w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP 提交正文失败：%w", err)
	}
	return c.Quit()
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
