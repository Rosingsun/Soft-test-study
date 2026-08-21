package repository

import (
	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/model"
)

// NotificationRepo 通知数据访问层（OPT-18：从 NotificationService 拆出）
type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Create 创建一条通知
func (r *NotificationRepo) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

// ListByUser 列某用户的最近 limit 条通知（按时间倒序）
func (r *NotificationRepo) ListByUser(userID uint, limit int) ([]model.Notification, error) {
	var list []model.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CountUnread 未读通知数
// 注意：read 是 MySQL 保留字，必须用反引号转义，否则会报 1064 语法错误
func (r *NotificationRepo) CountUnread(userID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND `read` = ?", userID, false).
		Count(&n).Error
	return n, err
}

// MarkRead 标记单条已读（按 userID + id 双重定位，避免越权）
func (r *NotificationRepo) MarkRead(userID, id uint) error {
	return r.db.Model(&model.Notification{}).
		Where("user_id = ? AND id = ?", userID, id).
		Update("`read`", true).Error
}

// MarkAllRead 全部标记已读
func (r *NotificationRepo) MarkAllRead(userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("user_id = ? AND `read` = ?", userID, false).
		Update("`read`", true).Error
}
