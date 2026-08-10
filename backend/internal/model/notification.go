package model

import "time"

// Notification 系统通知（AI 出题完成、错误、其他系统消息等）
type Notification struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"column:user_id;not null;index"`
	Type      string    `gorm:"column:type;type:varchar(40);default:''"` // ai_generate_done / ai_generate_failed / system ...
	Title     string    `gorm:"column:title;type:varchar(200);not null"`
	Content   string    `gorm:"column:content;type:text"`
	Link      string    `gorm:"column:link;type:varchar(500);default:''"` // 前端跳转链接
	Read      bool      `gorm:"column:read;default:false;index"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
