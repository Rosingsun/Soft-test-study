package middleware

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// allowedMethods 与 allowedHeaders 为 OPT-01 限定的白名单
var allowedMethods = "GET, POST, PUT, DELETE, PATCH, OPTIONS"
var allowedHeaders = "Origin, Content-Type, Authorization, X-Requested-With"

// CORSMiddleware 构造 CORS 中间件，按 origin 白名单动态返回头。
// 命中白名单：设置 Allow-Origin / Allow-Methods / Allow-Headers / Allow-Credentials
// 未命中：所有 CORS 头**不**被设置（浏览器将阻断跨域响应）
func CORSMiddleware(allowedOrigins []string, appEnv string) gin.HandlerFunc {
	whiteSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			whiteSet[o] = struct{}{}
		}
	}
	if len(whiteSet) == 0 && appEnv == "production" {
		log.Fatal("[cors] 生产环境必须配置 CORS_ALLOWED_ORIGINS，禁止启动")
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// 非浏览器请求（如 curl/服务端调用）继续放行
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
			return
		}
		if _, ok := whiteSet[origin]; !ok {
			// 不在白名单：不写任何 CORS 头，浏览器会拦截
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
			return
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "600")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
