package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/migrations"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// slowQueryThreshold 慢 SQL 阈值（毫秒）
const slowQueryThreshold = 200

// slowQueryLogger 仅打印超过阈值的慢 SQL，避免日志爆炸
type slowQueryLogger struct {
	threshold time.Duration
	infoLog   logger.Interface
}

func newSlowQueryLogger(thresholdMs int) logger.Interface {
	threshold := time.Duration(thresholdMs) * time.Millisecond
	base := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             threshold,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	return &slowQueryLogger{threshold: threshold, infoLog: base}
}

func (s *slowQueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	return s.infoLog.LogMode(level)
}

func (s *slowQueryLogger) Info(_ context.Context, _ string, _ ...interface{}) {}

func (s *slowQueryLogger) Warn(_ context.Context, _ string, _ ...interface{}) {}

func (s *slowQueryLogger) Error(_ context.Context, _ string, _ ...interface{}) {}

func (s *slowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	s.infoLog.Trace(ctx, begin, fc, err)
}

// Init 初始化数据库连接：
//   1. 打开 gorm.DB（业务代码继续用 gorm API）
//   2. 设置合理的连接池参数（远程 MySQL 单次 RTT ~300ms，必须放大池子）
//   3. 仅打印慢 SQL 日志（>=200ms），不再打印所有 SQL
//   4. 调用 migrations.MigrateUp 应用所有未执行的迁移
//
// 不再使用 GORM AutoMigrate，避免每次启动 100+ 条 information_schema 元数据慢 SQL。
func Init(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newSlowQueryLogger(slowQueryThreshold),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 *sql.DB 失败: %w", err)
	}
	// 远程 MySQL 网络延迟高（RTT ~300ms），连接池不能太小，否则请求排队
	// 同时也不能无限放大，避免打爆 DB
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// 业务 DSN 不带 multiStatements（安全考虑），迁移库内部自加 multiStatements=true 后单独连接
	if err := migrations.MigrateUp(dsn); err != nil {
		return nil, err
	}

	return db, nil
}
