package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) Today(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.Today(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取今日复习失败")
		return
	}
	response.Success(c, data)
}

func (h *ReviewHandler) Answer(c *gin.Context) {
	var req dto.ReviewAnswerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "请求参数错误")
		return
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.Answer(uid.(uint), req)
	if err != nil {
		respondError(c, err, "提交复习失败")
		return
	}
	response.Success(c, data)
}

func (h *ReviewHandler) Overview(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.Overview(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取复习概览失败")
		return
	}
	response.Success(c, data)
}
