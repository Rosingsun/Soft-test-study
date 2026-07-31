package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	jwtutil "github.com/soft-test-study/backend/pkg/jwt"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": config.CodeUnauthorized, "message": "未登录"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"code": config.CodeUnauthorized, "message": "Token 格式错误"})
			c.Abort()
			return
		}

		userID, err := jwtutil.Parse(jwtSecret, tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": config.CodeUnauthorized, "message": "Token 无效或已过期"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
