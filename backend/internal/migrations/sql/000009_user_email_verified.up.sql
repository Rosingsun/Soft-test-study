-- =============================================================
-- 000009_user_email_verified.up.sql
-- 在 users 表新增 email_verified 字段，标识当前邮箱是否已验证。
-- 业务说明：
--   - 字段默认 0（未验证）
--   - 注册时默认 0（注册流程暂不要求验证，UX 上不强制打断新用户）
--   - 用户在「个人中心 - 账号安全」通过 6 位验证码验证后置 1
--   - 换绑新邮箱时，新邮箱验证通过后 email_verified 重新置 1
-- 兼容性：
--   - 旧数据统一回填 1（已存在的用户邮箱历史可视为已验证，避免老用户看到「未验证」）
--   - 仅当表为空或字段不存在时才执行 ALTER（幂等性）
-- =============================================================

SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'email_verified'
);

SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `email_verified` TINYINT(1) NOT NULL DEFAULT 1
     COMMENT ''邮箱是否已验证 0=否 1=是'' AFTER `email`',
  'SELECT 1'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
