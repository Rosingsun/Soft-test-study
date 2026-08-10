-- =============================================================
-- 000002_practice_records_indexes.up.sql
-- 为 practice_records 加复合索引，支撑：
--   - stats.Overview/Daily/Calendar/SubjectProgress/ChapterProgress（按 user_id + created_at 聚合）
--   - repository.practice_record.CountByMode / DeleteCheckInRecordsByUserAndDate
--
-- 兼容性：旧版本 database.go 启动时已手动添加过 idx_user_created_at，
-- 本迁移用 prepared statement 检测后幂等创建，避免重复索引报错。
-- =============================================================

SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'practice_records'
    AND index_name = 'idx_user_created_at'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE practice_records ADD INDEX idx_user_created_at (user_id, created_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'practice_records'
    AND index_name = 'idx_user_mode_created'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE practice_records ADD INDEX idx_user_mode_created (user_id, mode, created_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
