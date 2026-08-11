package model

import "time"

// InvitationCode 邀请码记录
//
// 业务说明：
//   - 一次性使用：用完即焚（used_by 不为空即视为已用）
//   - 注册流程强制要求邀请码；管理员白名单（默认 ross）可绕过
//   - created_by 记录生成者 user_id；CLI 默认传 0
//   - used_by / used_at 在注册事务里由 service 层回填
//   - status=0 时即便未使用也视为失效（管理员紧急停用）
type InvitationCode struct {
	ID        uint       `gorm:"primarykey"`
	Code      string     `gorm:"column:code;type:varchar(32);uniqueIndex;not null"`
	CreatedBy uint       `gorm:"column:created_by;not null;default:0"`
	UsedBy    *uint      `gorm:"column:used_by"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	Note      string     `gorm:"column:note;type:varchar(200);default:''"`
	Status    int        `gorm:"column:status;type:tinyint;not null;default:1"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (InvitationCode) TableName() string {
	return "invitation_codes"
}
