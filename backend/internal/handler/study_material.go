package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type StudyMaterialHandler struct {
	svc *service.StudyMaterialService
}

func NewStudyMaterialHandler(svc *service.StudyMaterialService) *StudyMaterialHandler {
	return &StudyMaterialHandler{svc: svc}
}

func (h *StudyMaterialHandler) List(c *gin.Context) {
	var subjectID uint
	if s := c.Query("subject_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			response.Error(c, config.CodeParamError, "subject_id 参数格式错误")
			return
		}
		subjectID = uint(id)
	}
	list, err := h.svc.List(subjectID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取学习资料失败")
		return
	}
	response.Success(c, list)
}

func (h *StudyMaterialHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	resp, err := h.svc.Detail(uint(id))
	if err != nil {
		respondError(c, err, "获取学习资料详情失败")
		return
	}
	response.Success(c, resp)
}

func (h *StudyMaterialHandler) Content(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	nodes, err := h.svc.Content(uint(id))
	if err != nil {
		respondError(c, err, "解析思维导图失败，请下载 xmind 查看")
		return
	}
	response.Success(c, nodes)
}

func (h *StudyMaterialHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	absPath, fileName, err := h.svc.Download(uint(id))
	if err != nil {
		respondError(c, err, "下载失败")
		return
	}
	c.FileAttachment(absPath, fileName)
}
