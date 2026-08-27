package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type EmailVerificationCodeRepo struct {
	db *gorm.DB
}

func NewEmailVerificationCodeRepo(db *gorm.DB) *EmailVerificationCodeRepo {
	return &EmailVerificationCodeRepo{db: db}
}

// Create 写入新验证码记录
func (r *EmailVerificationCodeRepo) Create(c *model.EmailVerificationCode) error {
	return r.db.Create(c).Error
}

// FindValid 查询匹配 (user_id, email, purpose) 且未使用、未过期的最新一条记录
//
// 返回值：
//   - 找到：返回记录指针
//   - 找不到：返回 (nil, nil)，由 service 层判 ErrRecordNotFound
func (r *EmailVerificationCodeRepo) FindValid(userID uint, email, purpose, code string) (*model.EmailVerificationCode, error) {
	var rec model.EmailVerificationCode
	err := r.db.Where("user_id = ? AND email = ? AND purpose = ? AND code = ? AND used = 0 AND expires_at > ?",
		userID, email, purpose, code, time.Now()).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindLatest 查询同用户同 purpose 最近一条记录（用于 60s 冷却检查）
func (r *EmailVerificationCodeRepo) FindLatest(userID uint, purpose string) (*model.EmailVerificationCode, error) {
	var rec model.EmailVerificationCode
	err := r.db.Where("user_id = ? AND purpose = ?", userID, purpose).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindValidByEmail 按 (email, purpose, code) 查有效记录（未登录重置密码场景，user_id 为 NULL）
func (r *EmailVerificationCodeRepo) FindValidByEmail(email, purpose, code string) (*model.EmailVerificationCode, error) {
	var rec model.EmailVerificationCode
	err := r.db.Where("email = ? AND purpose = ? AND code = ? AND used = 0 AND expires_at > ?",
		email, purpose, code, time.Now()).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindLatestByEmail 按 (email, purpose) 查最近一条（未登录场景 60s 冷却）
func (r *EmailVerificationCodeRepo) FindLatestByEmail(email, purpose string) (*model.EmailVerificationCode, error) {
	var rec model.EmailVerificationCode
	err := r.db.Where("email = ? AND purpose = ?", email, purpose).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// CountTodayByEmail 按 email 维度统计今日发送次数（未登录重置密码场景）
func (r *EmailVerificationCodeRepo) CountTodayByEmail(email string) (int64, error) {
	var n int64
	startOfDay := startOfToday()
	err := r.db.Model(&model.EmailVerificationCode{}).
		Where("email = ? AND created_at >= ?", email, startOfDay).
		Count(&n).Error
	return n, err
}

// InvalidateActiveByEmail 失效同 (email, purpose) 的所有「未使用且未过期」记录（未登录场景发新码前清理）
func (r *EmailVerificationCodeRepo) InvalidateActiveByEmail(email, purpose string) error {
	return r.db.Model(&model.EmailVerificationCode{}).
		Where("email = ? AND purpose = ? AND used = 0 AND expires_at > ?", email, purpose, time.Now()).
		Update("used", true).Error
}

// CountToday 统计用户今日发送次数（purpose 不区分）
func (r *EmailVerificationCodeRepo) CountToday(userID uint) (int64, error) {
	var n int64
	startOfDay := startOfToday()
	err := r.db.Model(&model.EmailVerificationCode{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfDay).
		Count(&n).Error
	return n, err
}

// MarkUsed 标记验证码已使用
func (r *EmailVerificationCodeRepo) MarkUsed(id uint) error {
	return r.db.Model(&model.EmailVerificationCode{}).
		Where("id = ?", id).
		Update("used", true).Error
}

// InvalidateActive 将同用户同 purpose 的所有「未使用且未过期」记录标记为已使用
// 用于发送新码前清理旧码
func (r *EmailVerificationCodeRepo) InvalidateActive(userID uint, purpose string) error {
	return r.db.Model(&model.EmailVerificationCode{}).
		Where("user_id = ? AND purpose = ? AND used = 0 AND expires_at > ?", userID, purpose, time.Now()).
		Update("used", true).Error
}

// startOfToday 返回今天 0 点（本地时区）
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
