package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type SubjectHandler struct {
	svc *service.SubjectService
}

func NewSubjectHandler(svc *service.SubjectService) *SubjectHandler {
	return &SubjectHandler{svc: svc}
}

func (h *SubjectHandler) ListByLevel(c *gin.Context) {
	levelIDStr := c.Query("level_id")
	if levelIDStr == "" {
		response.Error(c, 10002, "缺少 level_id 参数")
		return
	}
	levelID, err := strconv.ParseUint(levelIDStr, 10, 64)
	if err != nil {
		response.Error(c, 10002, "level_id 参数格式错误")
		return
	}
	list, err := h.svc.GetByLevelID(uint(levelID))
	if err != nil {
		response.Error(c, 10001, "获取科目列表失败")
		return
	}
	response.Success(c, list)
}

func (h *SubjectHandler) ListAll(c *gin.Context) {
	list, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, 10001, "获取科目列表失败")
		return
	}
	response.Success(c, list)
}
