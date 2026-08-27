-- 000017_email_verification_codes_reset_password.down.sql
-- 回滚：删除 email 维度索引、把 user_id 还原为 NOT NULL
-- 注意：若有 reset_password 产生的 user_id NULL 行，需先清理

DROP INDEX `idx_email_purpose` ON `email_verification_codes`;

ALTER TABLE `email_verification_codes`
  MODIFY COLUMN `user_id` BIGINT UNSIGNED NOT NULL
  COMMENT '发起请求的用户 ID';
