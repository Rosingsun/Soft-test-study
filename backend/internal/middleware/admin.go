package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/repository"
)

// RequireAdmin 在 Auth 中间件之后挂载：仅 admin 角色可访问。
// 必须先经过 Auth 中间件（c.Get("user_id") 已被设置）。
func RequireAdmin(userRepo *repository.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    config.CodeUnauthorized,
				"message": "未登录",
			})
			return
		}
		userID, ok := v.(uint)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    config.CodeUnauthorized,
				"message": "用户身份异常",
			})
			return
		}
		user, err := userRepo.FindByID(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    config.CodeUnauthorized,
				"message": "用户不存在",
			})
			return
		}
		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    10003, // OPT-04: 文档约定的 10003 即无权限场景
				"message": "无权限",
			})
			return
		}
		c.Next()
	}
}
