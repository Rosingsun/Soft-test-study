-- =============================================================
-- 000017_email_verification_codes_reset_password.up.sql
-- 适配未登录「通过邮箱验证码重置密码」流程：
--   - user_id 改为可空（reset_password 场景下没有 user_id）
--   - 新增 email + purpose 维度索引，加速未登录场景的「查有效码」查询
-- 已有数据：
--   - 原 user_id NOT NULL 字段下无 NULL 数据（业务保证），直接 MODIFY 即可
-- =============================================================

ALTER TABLE `email_verification_codes`
  MODIFY COLUMN `user_id` BIGINT UNSIGNED NULL
  COMMENT '发起请求的用户 ID；reset_password 场景为 NULL';

CREATE INDEX `idx_email_purpose`
  ON `email_verification_codes` (`email`, `purpose`, `used`, `expires_at`);
