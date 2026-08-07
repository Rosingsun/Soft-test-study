package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type CheckInHandler struct {
	svc *service.CheckInService
}

func NewCheckInHandler(svc *service.CheckInService) *CheckInHandler {
	return &CheckInHandler{svc: svc}
}

func (h *CheckInHandler) Today(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.Today(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取今日打卡失败")
		return
	}
	response.Success(c, data)
}

func (h *CheckInHandler) Answer(c *gin.Context) {
	var req dto.CheckInAnswerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "请求参数错误")
		return
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.Answer(uid.(uint), req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, data)
}

func (h *CheckInHandler) Complete(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.Complete(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, data)
}

func (h *CheckInHandler) History(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		month = nowMonth()
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.History(uid.(uint), month)
	if err != nil {
		response.Error(c, config.CodeParamError, "月份参数格式错误，应为 YYYY-MM")
		return
	}
	response.Success(c, data)
}

func (h *CheckInHandler) Reset(c *gin.Context) {
	uid, _ := c.Get("user_id")
	if err := h.svc.Reset(uid.(uint)); err != nil {
		response.Error(c, config.CodeBadRequest, "重置今日打卡失败")
		return
	}
	response.Success(c, nil)
}

func (h *CheckInHandler) Stats(c *gin.Context) {
	uid, _ := c.Get("user_id")
	data, err := h.svc.Stats(uid.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取打卡统计失败")
		return
	}
	response.Success(c, data)
}
