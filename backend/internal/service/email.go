package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// EmailService 邮箱验证码业务
//
// 关键规则：
//   - 6 位纯数字验证码，crypto/rand 生成（无偏置）
//   - 有效期 5 分钟
//   - 同邮箱 60 秒冷却（防刷）
//   - 每用户每日 10 条上限
//   - purpose=change 时校验新邮箱未被其他用户占用
type EmailService struct {
	repo    *repository.EmailVerificationCodeRepo
	userRepo *repository.UserRepo
	sender  EmailSender
}

func NewEmailService(
	repo *repository.EmailVerificationCodeRepo,
	userRepo *repository.UserRepo,
	sender EmailSender,
) *EmailService {
	return &EmailService{repo: repo, userRepo: userRepo, sender: sender}
}

// 配置常量
const (
	emailCodeTTL       = 5 * time.Minute
	emailCodeCooldown  = 60 * time.Second
	emailCodeDailyMax  = 10
	emailPurposeVerify         = "verify"
	emailPurposeChange         = "change"
	emailPurposeResetPassword  = "reset_password"
)

// SendCode 生成并发送 6 位验证码
//
// 入参：
//   - userID：发起请求的用户 ID
//   - email：接收验证码的邮箱
//   - purpose：verify=验证当前邮箱；change=换绑新邮箱
//
// 流程：
//  1. 校验邮箱格式
//  2. purpose=change 时，校验邮箱未被其他用户占用
//  3. 60s 冷却检查
//  4. 当日发送次数上限检查
//  5. 失效同用户同 purpose 的旧码
//  6. 生成 6 位码 + 入库
//  7. 调用 EmailSender 发送（未配 SMTP 走日志）
func (s *EmailService) SendCode(userID uint, email, purpose string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return errors.New("邮箱格式不正确")
	}
	if purpose != emailPurposeVerify && purpose != emailPurposeChange {
		return errors.New("参数错误")
	}

	// 换绑场景：新邮箱不能已被其他用户占用
	if purpose == emailPurposeChange {
		existing, err := s.userRepo.FindByEmail(email)
		if err == nil && existing != nil && existing.ID > 0 && existing.ID != userID {
			return errors.New("该邮箱已被其他账号绑定")
		}
	}

	// 60s 冷却
	latest, err := s.repo.FindLatest(userID, purpose)
	if err == nil && latest != nil {
		if time.Since(latest.CreatedAt) < emailCodeCooldown {
			remaining := emailCodeCooldown - time.Since(latest.CreatedAt)
			return fmt.Errorf("请等待 %d 秒后再试", int(remaining.Seconds())+1)
		}
	}

	// 每日上限
	today, _ := s.repo.CountToday(userID)
	if today >= emailCodeDailyMax {
		return errors.New("今日发送次数已达上限，请明天再试")
	}

	// 失效旧码
	if err := s.repo.InvalidateActive(userID, purpose); err != nil {
		return errors.New("发送失败，请稍后重试")
	}

	// 生成验证码
	code, err := generateCode()
	if err != nil {
		return errors.New("生成验证码失败")
	}

	// 入库
	uid := userID
	rec := &model.EmailVerificationCode{
		UserID:    &uid,
		Email:     email,
		Code:      code,
		Purpose:   purpose,
		Used:      false,
		ExpiresAt: time.Now().Add(emailCodeTTL),
	}
	if err := s.repo.Create(rec); err != nil {
		return errors.New("发送失败，请稍后重试")
	}

	// 发送邮件
	subject, body := renderEmail(email, code, purpose)
	if err := s.sender.Send(email, subject, body); err != nil {
		return fmt.Errorf("邮件发送失败：%v", err)
	}
	return nil
}

// VerifyAndBind 校验验证码并完成绑定
//
// 入参：
//   - userID：当前操作用户
//   - email：验证码对应的目标邮箱
//   - code：用户填写的 6 位码
//   - purpose：verify / change
//
// 行为：
//   - verify：校验通过后把 users.email_verified 置 1（email 字段不变）
//   - change：校验通过后把 users.email 和 email_verified 都更新
func (s *EmailService) VerifyAndBind(userID uint, email, code, purpose string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return errors.New("邮箱格式不正确")
	}
	if purpose != emailPurposeVerify && purpose != emailPurposeChange {
		return errors.New("参数错误")
	}
	if len(code) != 6 {
		return errors.New("请输入 6 位验证码")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	rec, err := s.repo.FindValid(userID, email, purpose, code)
	if err != nil {
		return errors.New("验证码错误或已失效")
	}
	if err := s.repo.MarkUsed(rec.ID); err != nil {
		return errors.New("绑定失败，请稍后重试")
	}

	if purpose == emailPurposeChange {
		user.Email = email
	}
	user.EmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("绑定失败，请稍后重试")
	}
	return nil
}

// generateCode 用 crypto/rand 生成 6 位数字验证码（无偏置）
func generateCode() (string, error) {
	const max = 1000000
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// SendResetPasswordCode 未登录场景发送重置密码验证码
//
// 安全策略：
//   - 邮箱不存在 / 邮箱未验证：静默返回 nil（防止通过接口枚举已注册邮箱）
//   - 60s 冷却 / 单邮箱每日 10 条上限（复用常量）
//   - 验证码 user_id 写 NULL，service 层走 email 维度查
func (s *EmailService) SendResetPasswordCode(rawEmail string) error {
	email := strings.TrimSpace(strings.ToLower(rawEmail))
	if !isValidEmail(email) {
		return errors.New("邮箱格式不正确")
	}

	// 查用户：找不到 / 未验证 → 静默成功
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil || user.ID == 0 {
		return nil
	}
	if !user.EmailVerified {
		return nil
	}

	// 60s 冷却
	latest, err := s.repo.FindLatestByEmail(email, emailPurposeResetPassword)
	if err == nil && latest != nil {
		if time.Since(latest.CreatedAt) < emailCodeCooldown {
			remaining := emailCodeCooldown - time.Since(latest.CreatedAt)
			return fmt.Errorf("请等待 %d 秒后再试", int(remaining.Seconds())+1)
		}
	}

	// 每日上限
	today, _ := s.repo.CountTodayByEmail(email)
	if today >= emailCodeDailyMax {
		return errors.New("今日发送次数已达上限，请明天再试")
	}

	// 失效旧码
	if err := s.repo.InvalidateActiveByEmail(email, emailPurposeResetPassword); err != nil {
		return errors.New("发送失败，请稍后重试")
	}

	// 生成验证码
	code, err := generateCode()
	if err != nil {
		return errors.New("生成验证码失败")
	}

	// 入库（user_id 为 NULL）
	rec := &model.EmailVerificationCode{
		UserID:    nil,
		Email:     email,
		Code:      code,
		Purpose:   emailPurposeResetPassword,
		Used:      false,
		ExpiresAt: time.Now().Add(emailCodeTTL),
	}
	if err := s.repo.Create(rec); err != nil {
		return errors.New("发送失败，请稍后重试")
	}

	// 发送邮件
	subject, body := renderEmail(email, code, emailPurposeResetPassword)
	if err := s.sender.Send(email, subject, body); err != nil {
		return fmt.Errorf("邮件发送失败：%v", err)
	}
	return nil
}

// ResetPasswordByCode 校验重置密码验证码并写入新密码
//
// 入参：
//   - rawEmail：用户提交的邮箱
//   - code：6 位数字验证码
//   - newPassword：新密码（明文）
//
// 行为：
//   - 验证码错误 / 已失效：返回错误
//   - 校验通过后：标记 used → 写新密码 hash → 落库 → 清空登录失败计数与锁定状态
//   - 邮箱在 users 中查不到：返回「验证码错误或已失效」（防枚举）
func (s *EmailService) ResetPasswordByCode(rawEmail, code, newPassword string) error {
	email := strings.TrimSpace(strings.ToLower(rawEmail))
	if !isValidEmail(email) {
		return errors.New("邮箱格式不正确")
	}
	if len(code) != 6 {
		return errors.New("请输入 6 位验证码")
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// 按 email 查用户
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil || user.ID == 0 {
		return errors.New("验证码错误或已失效")
	}

	// 查有效码
	rec, err := s.repo.FindValidByEmail(email, emailPurposeResetPassword, code)
	if err != nil {
		return errors.New("验证码错误或已失效")
	}
	if err := s.repo.MarkUsed(rec.ID); err != nil {
		return errors.New("重置失败，请稍后重试")
	}

	// 写新密码 hash
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("重置失败，请稍后重试")
	}
	user.PasswordHash = string(hash)
	// 改密成功后清空登录失败计数与锁定状态，避免被锁定的账号改密后仍不可用
	user.FailedAttempts = 0
	user.LockedUntil = nil

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("重置失败，请稍后重试")
	}
	return nil
}

// isValidEmail 校验邮箱格式（使用标准库 net/mail）
func isValidEmail(s string) bool {
	if len(s) > 100 {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

// renderEmail 渲染邮件标题与正文
func renderEmail(email, code, purpose string) (subject, body string) {
	var action string
	switch purpose {
	case emailPurposeChange:
		action = "换绑"
	case emailPurposeResetPassword:
		action = "重置密码"
	default:
		action = "验证"
	}
	subject = fmt.Sprintf("【软考学系】邮箱%s验证码", action)
	body = fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>%s</title></head>
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#f6f7fb;padding:24px;">
  <div style="max-width:480px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.06);">
    <div style="background:linear-gradient(135deg,#4f46e5,#7c3aed);padding:24px 32px;color:#fff;">
      <h1 style="margin:0;font-size:20px;">软考学系</h1>
      <p style="margin:4px 0 0;opacity:0.85;font-size:13px;">邮箱%s</p>
    </div>
    <div style="padding:32px;">
      <p style="margin:0 0 16px;color:#374151;font-size:14px;">您好，您正在进行邮箱%s操作，验证码为：</p>
      <div style="background:#f3f4f6;border-radius:8px;padding:20px;text-align:center;margin:16px 0;">
        <span style="font-family:'SF Mono',Consolas,monospace;font-size:32px;font-weight:700;letter-spacing:8px;color:#4f46e5;">%s</span>
      </div>
      <p style="margin:16px 0 0;color:#6b7280;font-size:13px;line-height:1.6;">
        · 验证码有效期为 <b>5 分钟</b>，请尽快使用<br>
        · 如非本人操作，请忽略此邮件<br>
        · 接收邮箱：%s
      </p>
    </div>
    <div style="border-top:1px solid #f3f4f6;padding:16px 32px;color:#9ca3af;font-size:12px;text-align:center;">
      本邮件由系统自动发送，请勿直接回复
    </div>
  </div>
</body></html>`, subject, action, action, code, email)
	return
}
