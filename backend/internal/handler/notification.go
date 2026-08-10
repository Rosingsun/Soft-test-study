package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func currentUserID(c *gin.Context) (uint, bool) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	return userID, ok && userID != 0
}

// List 列出当前用户通知（默认 limit=20）
func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	list, unread, err := h.svc.List(userID, limit)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取通知失败: "+err.Error())
		return
	}
	response.Success(c, dto.NotificationListResp{
		List:        service.ToNotificationRespList(list),
		UnreadCount: unread,
	})
}

// UnreadCount 仅查未读数（轮询用）
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	response.Success(c, dto.UnreadCountResp{UnreadCount: h.svc.UnreadCount(userID)})
}

// MarkRead 标记单条已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "id 必须为数字")
		return
	}
	if err := h.svc.MarkRead(userID, uint(id64)); err != nil {
		response.Error(c, config.CodeBadRequest, "标记已读失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// MarkAllRead 全部标记已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	if err := h.svc.MarkAllRead(userID); err != nil {
		response.Error(c, config.CodeBadRequest, "标记全部已读失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}
