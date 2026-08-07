package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type StudyPlanHandler struct {
	svc *service.StudyPlanService
}

func NewStudyPlanHandler(svc *service.StudyPlanService) *StudyPlanHandler {
	return &StudyPlanHandler{svc: svc}
}

func (h *StudyPlanHandler) List(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.List(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取学习计划失败")
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) Create(c *gin.Context) {
	var req dto.StudyPlanCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "请求参数错误")
		return
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.Create(uid.(uint), req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "计划 ID 错误")
		return
	}
	var req dto.StudyPlanUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "请求参数错误")
		return
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.Update(uid.(uint), uint(id), req)
	if err != nil {
		respondError(c, err, "更新计划失败")
		return
	}
	response.Success(c, data)
}

func (h *StudyPlanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "计划 ID 错误")
		return
	}
	uid, _ := c.Get("user_id")
	if err := h.svc.Delete(uid.(uint), uint(id)); err != nil {
		respondError(c, err, "删除计划失败")
		return
	}
	response.Success(c, nil)
}
