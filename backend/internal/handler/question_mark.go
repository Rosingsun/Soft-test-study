package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type QuestionMarkHandler struct {
	svc *service.QuestionMarkService
}

func NewQuestionMarkHandler(svc *service.QuestionMarkService) *QuestionMarkHandler {
	return &QuestionMarkHandler{svc: svc}
}

func (h *QuestionMarkHandler) AddMark(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	if err := h.svc.Mark(userID.(uint), uint(questionID)); err != nil {
		response.Error(c, 10001, "标记失败")
		return
	}
	response.Success(c, gin.H{"marked": true})
}

func (h *QuestionMarkHandler) RemoveMark(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	if err := h.svc.Unmark(userID.(uint), uint(questionID)); err != nil {
		response.Error(c, 10001, "取消标记失败")
		return
	}
	response.Success(c, gin.H{"marked": false})
}

func (h *QuestionMarkHandler) ListMarks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	list, err := h.svc.List(userID.(uint))
	if err != nil {
		response.Error(c, 10001, "获取标记列表失败")
		return
	}
	response.Success(c, list)
}

func (h *QuestionMarkHandler) CheckMarked(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)

	marked, err := h.svc.IsMarked(userID.(uint), uint(questionID))
	if err != nil {
		response.Error(c, 10001, "查询失败")
		return
	}
	response.Success(c, gin.H{"marked": marked})
}
