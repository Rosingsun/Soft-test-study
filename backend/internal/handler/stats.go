package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type StatsHandler struct {
	svc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	data, err := h.svc.GetOverview(userID.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取统计数据失败")
		return
	}
	response.Success(c, data)
}

func (h *StatsHandler) Daily(c *gin.Context) {
	userID, _ := c.Get("user_id")
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	data, err := h.svc.GetDailyStats(userID.(uint), days)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取每日统计失败")
		return
	}
	response.Success(c, data)
}

func (h *StatsHandler) SubjectProgress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	data, err := h.svc.GetSubjectProgress(userID.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取科目进度失败")
		return
	}
	response.Success(c, data)
}

func (h *StatsHandler) Calendar(c *gin.Context) {
	userID, _ := c.Get("user_id")
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	data, err := h.svc.GetCalendarStats(userID.(uint), month)
	if err != nil {
		response.Error(c, config.CodeParamError, "月份参数格式错误，应为 YYYY-MM")
		return
	}
	response.Success(c, data)
}

func (h *StatsHandler) ChapterProgress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	data, err := h.svc.GetChapterProgress(userID.(uint))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取章节掌握度失败")
		return
	}
	response.Success(c, data)
}
