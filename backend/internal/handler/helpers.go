package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

// respondError 将 service 层错误映射为统一错误码响应
func respondError(c *gin.Context, err error, defaultMsg string) {
	switch {
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
