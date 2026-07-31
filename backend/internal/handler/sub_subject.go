package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type SubSubjectHandler struct {
	svc *service.SubSubjectService
}

func NewSubSubjectHandler(svc *service.SubSubjectService) *SubSubjectHandler {
	return &SubSubjectHandler{svc: svc}
}

func (h *SubSubjectHandler) ListBySubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, 10002, "参数格式错误")
		return
	}
	list, err := h.svc.GetBySubjectID(uint(id))
	if err != nil {
		response.Error(c, 10001, "获取子科目列表失败")
		return
	}
	response.Success(c, list)
}
