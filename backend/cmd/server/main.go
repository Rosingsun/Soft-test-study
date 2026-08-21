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

	// OPT-02: 生产环境强制校验 JWT 密钥强度
	if cfg.AppEnv == "production" {
		if cfg.JWTSecret == "" ||
			cfg.JWTSecret == "dev-secret-change-me" ||
			len(cfg.JWTSecret) < 32 {
			log.Fatal("JWT_SECRET 未配置或强度不足，禁止生产环境启动")
		}
	}

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
