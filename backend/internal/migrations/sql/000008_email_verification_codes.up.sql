-- =============================================================
-- 000008_email_verification_codes.up.sql
-- 新增 email_verification_codes 表，存储邮箱验证码。
-- 用于个人中心的「邮箱验证 / 换绑」功能：
--   - purpose=verify：验证当前邮箱
--   - purpose=change：换绑新邮箱（先发到新邮箱，校验通过后写入 users.email）
-- 字段：
--   - user_id：发起请求的用户（换绑场景下也归属原用户，用于审计/限流）
--   - email：验证码发往的目标邮箱
--   - code：6 位数字验证码（crypto/rand 生成）
--   - purpose：用途（verify / change）
--   - used：是否已使用（防止重复消费）
--   - expires_at：过期时间（5 分钟）
-- 设计：
--   - 同一用户同 purpose 同时刻最多 1 条有效记录（发送时先失效旧记录）
--   - 通过 idx_user_email_purpose 加速「查有效码」查询
--   - 通过 idx_expires 加速 janitor 清理过期记录
-- =============================================================

CREATE TABLE `email_verification_codes` (
  `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED  NOT NULL                COMMENT '发起请求的用户 ID',
  `email`      VARCHAR(100)     NOT NULL                COMMENT '接收验证码的邮箱',
  `code`       VARCHAR(6)       NOT NULL                COMMENT '6 位数字验证码',
  `purpose`    VARCHAR(20)      NOT NULL                COMMENT 'verify=验证当前邮箱 / change=换绑新邮箱',
  `used`       TINYINT(1)       NOT NULL DEFAULT 0      COMMENT '是否已使用',
  `expires_at` DATETIME         NOT NULL                COMMENT '过期时间（5 分钟）',
  `created_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_email_purpose` (`user_id`, `email`, `purpose`),
  KEY `idx_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='邮箱验证码表';
