package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	router.Setup(db, r, cfg)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
