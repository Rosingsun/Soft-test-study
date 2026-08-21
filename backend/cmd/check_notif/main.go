// 临时工具：直接通过 service 验证 notification SQL 转义修复
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	repo := repository.NewNotificationRepo(db)

	// 清理上次测试残留
	db.Where("user_id = ?", 99999).Delete(&model.Notification{})

	// 插 3 条未读
	for i := 0; i < 3; i++ {
		if err := repo.Create(&model.Notification{
			UserID: 99999, Type: "system", Title: fmt.Sprintf("t%d", i), Content: "x", Read: false,
		}); err != nil {
			log.Fatalf("create: %v", err)
		}
	}
	// 插 1 条已读
	if err := repo.Create(&model.Notification{
		UserID: 99999, Type: "system", Title: "read1", Content: "x", Read: true,
	}); err != nil {
		log.Fatalf("create read: %v", err)
	}

	n, err := repo.CountUnread(99999)
	if err != nil {
		log.Fatalf("CountUnread FAILED: %v", err)
	}
	fmt.Printf("[1] CountUnread = %d (期望 3)\n", n)

	if err := repo.MarkAllRead(99999); err != nil {
		log.Fatalf("MarkAllRead FAILED: %v", err)
	}
	fmt.Println("[2] MarkAllRead OK")

	n2, err := repo.CountUnread(99999)
	if err != nil {
		log.Fatalf("CountUnread after FAILED: %v", err)
	}
	fmt.Printf("[3] CountUnread after = %d (期望 0)\n", n2)

	// MarkRead 单条
	if err := repo.Create(&model.Notification{
		UserID: 99999, Type: "system", Title: "t4", Content: "x", Read: false,
	}); err != nil {
		log.Fatalf("create t4: %v", err)
	}
	var last model.Notification
	db.Where("user_id = ? AND title = ?", 99999, "t4").First(&last)
	if err := repo.MarkRead(99999, last.ID); err != nil {
		log.Fatalf("MarkRead FAILED: %v", err)
	}
	fmt.Println("[4] MarkRead OK")

	// 清理
	db.Where("user_id = ?", 99999).Delete(&model.Notification{})
	fmt.Println("✅ notifications.read 反引号修复验证通过")
}
