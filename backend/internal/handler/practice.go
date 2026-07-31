package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type PracticeHandler struct {
	svc *service.PracticeService
}

func NewPracticeHandler(svc *service.PracticeService) *PracticeHandler {
	return &PracticeHandler{svc: svc}
}

func (h *PracticeHandler) Submit(c *gin.Context) {
	var req dto.PracticeSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, "请求参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	record, err := h.svc.Submit(uid, req)
	if err != nil {
		response.Error(c, 10001, "提交答案失败")
		return
	}
	response.Success(c, record)
}
