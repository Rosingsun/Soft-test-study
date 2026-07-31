package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type ChapterHandler struct {
	svc *service.ChapterService
}

func NewChapterHandler(svc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{svc: svc}
}

func (h *ChapterHandler) ListBySubSubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 10002, "参数格式错误")
		return
	}
	list, err := h.svc.GetBySubSubjectID(uint(id))
	if err != nil {
		response.Error(c, 10001, "获取章节列表失败")
		return
	}
	response.Success(c, list)
}
