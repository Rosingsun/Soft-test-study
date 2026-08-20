// Package migrations 封装 golang-migrate 启动逻辑
//
// 业务说明：
//   - 启动时只调一次 MigrateUp，按顺序执行未应用版本的迁移
//   - 不再依赖 GORM AutoMigrate，避免启动期 100+ 条 information_schema 元数据查询
//   - 迁移文件通过 embed.FS 编译进二进制，不再依赖外部 ./migrations 目录
//   - 文件命名规则 NNN_name.up.sql / .down.sql，位于 internal/migrations/sql/
package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

// embeddedDir 是 iofs 在 embed FS 内的根路径。
const embeddedDir = "sql"

// newSource 基于编译期嵌入的迁移文件构造一个 source.Driver。
func newSource() (source.Driver, error) {
	src, err := iofs.New(migrationsFS, embeddedDir)
	if err != nil {
		return nil, fmt.Errorf("migrations: 加载嵌入迁移失败: %w", err)
	}
	return src, nil
}

// MigrateUp 应用所有未执行的迁移
//
// 行为：
//   - 若 schema_migrations 表无脏数据，则按版本顺序应用所有未应用迁移
//   - 若 ErrNoChange（已是最新版本），直接返回 nil，不视为错误
//   - 注意：MySQL 驱动要求 DSN 携带 multiStatements=true（迁移文件含多语句）
//   - 业务连接仍用单语句 DSN（避免 SQL 注入面扩大）
func MigrateUp(appDSN string) error {
	db, driver, err := openMigrationDB(appDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	defer driver.Close()

	src, err := newSource()
	if err != nil {
		return err
	}
	defer src.Close()

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrations: 创建 migrate 实例失败: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: 应用迁移失败: %w", err)
	}
	return nil
}

// MigrateDown 回滚 N 个迁移版本（默认 1）
func MigrateDown(appDSN string, steps int) error {
	if steps <= 0 {
		steps = 1
	}
	db, driver, err := openMigrationDB(appDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	defer driver.Close()

	src, err := newSource()
	if err != nil {
		return err
	}
	defer src.Close()

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrations: 创建 migrate 实例失败: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: 回滚迁移失败: %w", err)
	}
	return nil
}

// Force 强制将 schema_migrations 版本号设为指定值（仅用于修复脏数据）
func Force(appDSN string, version int) error {
	db, driver, err := openMigrationDB(appDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	defer driver.Close()

	src, err := newSource()
	if err != nil {
		return err
	}
	defer src.Close()

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrations: 创建 migrate 实例失败: %w", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("migrations: force 版本失败: %w", err)
	}
	return nil
}

// openMigrationDB 打开一个「仅给迁移用」的连接：multiStatements=true。
//
// 输入 DSN 是业务 DSN（不允许多语句），这里在 DSN 上追加 multiStatements=true
// 后开一个独立的 *sql.DB 给 golang-migrate 用，业务连接保持单语句。
func openMigrationDB(appDSN string) (*sql.DB, *migratemysql.Mysql, error) {
	migDSN, err := withMultiStatements(appDSN)
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := sql.Open("mysql", migDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("migrations: 打开 mysql 连接失败: %w", err)
	}
	// 迁移连接只跑几条 SQL，不需要大连接池
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(1)
	drv, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("migrations: 初始化 mysql 驱动失败: %w", err)
	}
	driver, ok := drv.(*migratemysql.Mysql)
	if !ok {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("migrations: 驱动类型断言失败")
	}
	return sqlDB, driver, nil
}

// withMultiStatements 在原 DSN 上追加 multiStatements=true。
// 已存在则保持不变。
func withMultiStatements(dsn string) (string, error) {
	// go-sql-driver/mysql DSN 是 key=value 形式，用 & 连接
	parts := strings.Split(dsn, "?")
	if len(parts) == 1 {
		return dsn + "?multiStatements=true", nil
	}
	base, query := parts[0], parts[1]
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("migrations: 解析 DSN query 失败: %w", err)
	}
	values.Set("multiStatements", "true")
	return base + "?" + values.Encode(), nil
}
