package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

// nowMonth 返回当前月份 YYYY-MM
func nowMonth() string {
	return time.Now().Format("2006-01")
}

// respondError 将 service 层错误映射为统一错误码响应
func respondError(c *gin.Context, err error, defaultMsg string) {
	switch {
	case errors.Is(err, service.ErrInvalidFilePath):
		// OPT-03: 路径穿越尝试，返回 403
		response.Error(c, config.CodeForbidden, "非法文件路径")
	case errors.Is(err, service.ErrExamInProgress):
		// OPT-07: 并发重复创建考试，DB 唯一约束命中 → 409
		c.JSON(http.StatusConflict, gin.H{
			"code":    10009,
			"message": "已有进行中的考试",
		})
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, config.CodeForbidden, "无权操作")
	case errors.Is(err, service.ErrNotFound):
		response.Error(c, config.CodeNotFound, "资源不存在")
	case errors.Is(err, service.ErrConflict):
		response.Error(c, config.CodeBadRequest, err.Error())
	default:
		response.Error(c, config.CodeBadRequest, defaultMsg)
	}
}
