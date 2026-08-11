package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type WrongQuestionHandler struct {
	svc *service.WrongQuestionService
}

func NewWrongQuestionHandler(svc *service.WrongQuestionService) *WrongQuestionHandler {
	return &WrongQuestionHandler{svc: svc}
}

func (h *WrongQuestionHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var subjectID *uint
	if s := c.Query("subject_id"); s != "" {
		if id, err := strconv.ParseUint(s, 10, 64); err == nil {
			v := uint(id)
			subjectID = &v
		}
	}
	source := c.Query("source")
	if source != "ai" && source != "real" {
		source = "all"
	}
	sort := c.Query("sort")
	if sort != "wrong_count" && sort != "last_wrong_at" {
		sort = "wrong_count"
	}
	list, err := h.svc.List(userID.(uint), subjectID, source, sort)
	if err != nil {
		response.Error(c, 10001, "获取错题列表失败")
		return
	}
	response.Success(c, list)
}

func (h *WrongQuestionHandler) Count(c *gin.Context) {
	userID, _ := c.Get("user_id")
	count, err := h.svc.Count(userID.(uint))
	if err != nil {
		response.Error(c, 10001, "获取错题数量失败")
		return
	}
	response.Success(c, gin.H{"count": count})
}

func (h *WrongQuestionHandler) PracticeSubmit(c *gin.Context) {
	var req dto.WrongPracticeSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, "参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	record, err := h.svc.PracticeSubmit(userID.(uint), req)
	if err != nil {
		response.Error(c, 10001, "提交失败")
		return
	}
	response.Success(c, record)
}

func (h *WrongQuestionHandler) Remove(c *gin.Context) {
	userID, _ := c.Get("user_id")
	questionIDStr := c.Param("id")
	questionID, _ := strconv.ParseUint(questionIDStr, 10, 64)
	if err := h.svc.Remove(userID.(uint), uint(questionID)); err != nil {
		response.Error(c, 10001, "移出失败")
		return
	}
	response.Success(c, nil)
}
