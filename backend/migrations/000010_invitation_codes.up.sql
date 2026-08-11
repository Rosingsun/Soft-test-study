-- =============================================================
-- 000010_invitation_codes.up.sql
-- 邀请码表：用于控制新用户注册（一次性、必填、记录邀请人/使用人）
--
-- 业务说明：
--   - 注册流程要求用户必须提供有效邀请码（管理员用户名 ross 除外）
--   - 邀请码为一次性：used_by 不为空即视为已用
--   - code 字段唯一：CLI 批量生成时按唯一约束自动去重
--   - 明文存码便于管理员直接查询分发；DB 泄露风险由 status 控制（停用即失效）
--
-- 兼容性：
--   - 全新表，无需 ALTER
--   - 与现有 users 表无强外键关联（邀请人/使用人均用 user_id 软引用，保留历史）
-- =============================================================

CREATE TABLE IF NOT EXISTS `invitation_codes` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32)     NOT NULL COMMENT '邀请码明文(管理员可见)',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生成者 user_id(admin)',
  `used_by`    BIGINT UNSIGNED DEFAULT NULL COMMENT '使用者 user_id',
  `used_at`    DATETIME        DEFAULT NULL,
  `note`       VARCHAR(200)    DEFAULT '' COMMENT '备注(批次/活动)',
  `status`     TINYINT         NOT NULL DEFAULT 1 COMMENT '1=启用 0=停用',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_status` (`status`),
  KEY `idx_used_by` (`used_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请码表';
