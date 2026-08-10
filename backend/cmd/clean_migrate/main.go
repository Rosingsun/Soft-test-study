// 临时工具：清理 schema_migrations 整张表（脏状态完全重置）。
//
// 适用于：版本表脏、无法 force 回到干净状态。
//
// 用法：
//   go run ./cmd/clean_migrate
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/soft-test-study/backend/internal/config"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.DBPassword == "" {
		log.Println("警告: 未配置 DB_PASSWORD")
	}
	// 不用 multiStatements，单条 DROP 也够了
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("打开失败: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("DROP TABLE IF EXISTS schema_migrations"); err != nil {
		log.Fatalf("DROP schema_migrations 失败: %v", err)
	}
	fmt.Println("schema_migrations 已删除。下次启动 server 将从版本 0 开始重新应用迁移。")
}
