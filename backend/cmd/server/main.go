package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/middleware"
	"github.com/soft-test-study/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	// OPT-01: CORS 白名单（替换原通配符实现）
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins, cfg.AppEnv))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	router.Setup(db, r, cfg)

	// OPT-03: 静态目录不再以 r.Static 直接暴露，由 handler 鉴权后流式返回
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
