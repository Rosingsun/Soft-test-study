// 临时工具：将 schema_migrations 强制设为指定版本，用于修复"脏"状态。
//
// 用法：
//   go run ./cmd/force -v 0   // 把版本回退到 0（重新应用 000001）
//   go run ./cmd/force -v 5   // 把版本设为 5（认为 000001~000005 已应用）
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/migrations"
)

func main() {
	version := flag.Int("v", 0, "目标版本号")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.DBPassword == "" {
		log.Println("警告: 未配置 DB_PASSWORD")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	if err := migrations.Force(dsn, *version); err != nil {
		log.Fatalf("force 失败: %v", err)
	}
	fmt.Printf("schema_migrations 已设为版本 %d\n", *version)
	os.Exit(0)
}
