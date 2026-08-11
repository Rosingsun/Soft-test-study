package model

import "time"

// EmailVerificationCode 邮箱验证码记录
//
// 业务说明：
//   - 用于个人中心「邮箱验证 / 换绑」流程
//   - purpose=verify：验证当前邮箱
//   - purpose=change：换绑新邮箱（验证通过后由 service 层把 users.email 改成此处的 email）
//   - 同一用户同 purpose 同时刻最多 1 条有效记录（发送时先失效旧记录）
type EmailVerificationCode struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"column:user_id;not null"`
	Email     string    `gorm:"column:email;type:varchar(100);not null"`
	Code      string    `gorm:"column:code;type:varchar(6);not null"`
	Purpose   string    `gorm:"column:purpose;type:varchar(20);not null"`
	Used      bool      `gorm:"column:used;type:tinyint(1);default:0"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (EmailVerificationCode) TableName() string {
	return "email_verification_codes"
}
