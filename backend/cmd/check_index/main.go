// 临时工具：验证 13 号迁移的 uk_exam_pending 唯一索引是否在 exam_records 上建好
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	// 1. 索引是否在建
	var idxCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = ? AND table_name = 'exam_records'
		  AND index_name = 'uk_exam_pending'`,
		cfg.DBName,
	).Scan(&idxCount)
	if err != nil {
		log.Fatalf("query index: %v", err)
	}
	fmt.Printf("[1] uk_exam_pending 索引存在数: %d (期望: 1)\n", idxCount)

	// 2. 生成列是否在表上
	var colCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = ? AND table_name = 'exam_records'
		  AND column_name = 'pending_dedup_key'`,
		cfg.DBName,
	).Scan(&colCount)
	if err != nil {
		log.Fatalf("query col: %v", err)
	}
	fmt.Printf("[2] pending_dedup_key 生成列存在数: %d (期望: 1)\n", colCount)

	// 3. schema_migrations 当前版本
	var version int
	var dirty int
	err = db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	if err != nil {
		log.Fatalf("query schema_migrations: %v", err)
	}
	fmt.Printf("[3] schema_migrations 当前版本: %d, dirty: %d (期望: 13, 0)\n", version, dirty)

	// 4. 模拟插入重复 pending 记录，验证唯一索引确实生效
	// 选取 user_id=1, template_id=1
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("begin: %v", err)
	}
	_, err = tx.Exec(`DELETE FROM exam_records WHERE user_id=1 AND template_id=1 AND status='pending'`)
	if err != nil {
		tx.Rollback()
		log.Fatalf("clean: %v", err)
	}
	_, err = tx.Exec(`INSERT INTO exam_records (user_id, template_id, status, started_at) VALUES (1, 1, 'pending', NOW())`)
	if err != nil {
		tx.Rollback()
		log.Fatalf("insert 1: %v", err)
	}
	_, err = tx.Exec(`INSERT INTO exam_records (user_id, template_id, status, started_at) VALUES (1, 1, 'pending', NOW())`)
	if err == nil {
		tx.Rollback()
		log.Fatalf("[4] 第二次插入未报错，唯一索引未生效!")
	}
	fmt.Printf("[4] 第二次插入被拒绝（符合预期）: %v\n", err)
	tx.Rollback()
	fmt.Println("✅ 13 号迁移验证通过")
}
