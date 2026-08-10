package service

import (
	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
)

// NotificationService 通知服务（CRUD + 主动推送）
type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Push 主动推送一条通知
func (s *NotificationService) Push(userID uint, ntype, title, content, link string) error {
	n := &model.Notification{
		UserID:  userID,
		Type:    ntype,
		Title:   title,
		Content: content,
		Link:    link,
		Read:    false,
	}
	return s.db.Create(n).Error
}

// List 列出某用户的通知（最近 limit 条），按时间倒序
func (s *NotificationService) List(userID uint, limit int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var unread int64
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.Model(&model.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&unread).Error; err != nil {
		return list, 0, nil
	}
	return list, unread, nil
}

// UnreadCount 未读数
func (s *NotificationService) UnreadCount(userID uint) int64 {
	var n int64
	s.db.Model(&model.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&n)
	return n
}

// MarkRead 标记单条已读
func (s *NotificationService) MarkRead(userID, id uint) error {
	return s.db.Model(&model.Notification{}).
		Where("user_id = ? AND id = ?", userID, id).
		Update("read", true).Error
}

// MarkAllRead 全部标记已读
func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.db.Model(&model.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error
}

// ToResp 模型转 DTO
func ToNotificationResp(n model.Notification) dto.NotificationResp {
	return dto.NotificationResp{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Content:   n.Content,
		Link:      n.Link,
		Read:      n.Read,
		CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToRespList 列表转换
func ToNotificationRespList(list []model.Notification) []dto.NotificationResp {
	out := make([]dto.NotificationResp, 0, len(list))
	for _, n := range list {
		out = append(out, ToNotificationResp(n))
	}
	return out
}
