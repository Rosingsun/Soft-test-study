-- =============================================================
-- 000009_user_email_verified.down.sql
-- 撤销 users.email_verified 字段。
-- =============================================================

SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'email_verified'
);

SET @sql := IF(
  @col_exists > 0,
  'ALTER TABLE `users` DROP COLUMN `email_verified`',
  'SELECT 1'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
