package handler

import (
	"errors"
	"net/http"
	"strconv"
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

// normalizePageParams 归一化分页参数，保证响应里的 page/page_size 与实际取值一致。
//
// 说明：service 层同样会做兜底（避免越界)，此处再归一一次是为了让前端拿到的
// page_size 与真实分页口径一致，防止首页返回 size=0 导致 UI 计算出错。
func normalizePageParams(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// parseUserIDParam 解析路由路径中的 :id 参数为 uint。
// 缺失或非法时直接写 400 响应并返回 ok=false，调用方必须立即 return。
func parseUserIDParam(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return 0, false
	}
	return uint(id), true
}

// mustUserID 从 gin 上下文获取当前用户 ID。
// 缺失或类型不合法时直接写 401 响应并返回 ok=false，调用方必须立即 return。
// 避免 nil/0 强转 uint 透传到 service 后被当作"他人"触发"无权操作"。
func mustUserID(c *gin.Context) (uint, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return 0, false
	}
	id, ok := v.(uint)
	if !ok || id == 0 {
		response.Error(c, config.CodeUnauthorized, "用户身份异常")
		return 0, false
	}
	return id, true
}
