package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

// NotificationService 通知服务（CRUD + 主动推送）
// OPT-18：持有 *repository.NotificationRepo，遵循 AGENTS.md 分层规则（repository 层做数据库查询）
type NotificationService struct {
	repo *repository.NotificationRepo
}

func NewNotificationService(repo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
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
	return s.repo.Create(n)
}

// List 列出某用户的通知（最近 limit 条），按时间倒序
func (s *NotificationService) List(userID uint, limit int) ([]model.Notification, int64, error) {
	list, err := s.repo.ListByUser(userID, limit)
	if err != nil {
		return nil, 0, err
	}
	unread, err := s.repo.CountUnread(userID)
	if err != nil {
		return list, 0, nil
	}
	return list, unread, nil
}

// UnreadCount 未读数
func (s *NotificationService) UnreadCount(userID uint) int64 {
	n, err := s.repo.CountUnread(userID)
	if err != nil {
		return 0
	}
	return n
}

// MarkRead 标记单条已读
func (s *NotificationService) MarkRead(userID, id uint) error {
	return s.repo.MarkRead(userID, id)
}

// MarkAllRead 全部标记已读
func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
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
