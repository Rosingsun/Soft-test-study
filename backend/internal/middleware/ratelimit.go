package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
)

type visitorEntry struct {
	count   int
	resetAt time.Time
}

func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*visitorEntry)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists || now.After(v.resetAt) {
			visitors[ip] = &visitorEntry{count: 1, resetAt: now.Add(window)}
			mu.Unlock()
			c.Next()
			return
		}

		if v.count >= maxRequests {
			mu.Unlock()
			// OPT-23: 改用 CodeTooManyRequests (10006)，避免与 CodeForbidden (10004) 冲突
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    config.CodeTooManyRequests,
				"message": "操作过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		v.count++
		mu.Unlock()
		c.Next()
	}
}
