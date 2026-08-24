package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type RankingHandler struct {
	svc *service.RankingService
}

func NewRankingHandler(svc *service.RankingService) *RankingHandler {
	return &RankingHandler{svc: svc}
}

func (h *RankingHandler) Get(c *gin.Context) {
	category := c.Query("category")
	metric := c.Query("metric")
	if category == "" || metric == "" {
		response.Error(c, config.CodeParamError, "缺少 category 或 metric 参数")
		return
	}
	subjectID := uint(0)
	if v := c.Query("subject_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			subjectID = uint(n)
		}
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.Get(uid.(uint), category, metric, subjectID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取排行失败")
		return
	}
	response.Success(c, data)
}

// GetEstimatedScores 分项预估分（选择题 / 案例分析 / 论文）
func (h *RankingHandler) GetEstimatedScores(c *gin.Context) {
	subjectID := uint(0)
	if v := c.Query("subject_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			subjectID = uint(n)
		}
	}
	uid, _ := c.Get("user_id")
	data, err := h.svc.GetEstimatedSectionScores(uid.(uint), subjectID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取预估分失败")
		return
	}
	response.Success(c, data)
}
