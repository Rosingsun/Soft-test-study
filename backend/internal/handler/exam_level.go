package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type ExamLevelHandler struct {
	svc *service.ExamLevelService
}

func NewExamLevelHandler(svc *service.ExamLevelService) *ExamLevelHandler {
	return &ExamLevelHandler{svc: svc}
}

func (h *ExamLevelHandler) List(c *gin.Context) {
	list, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, 10001, "获取等级列表失败")
		return
	}
	response.Success(c, list)
}
