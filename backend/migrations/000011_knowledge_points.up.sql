-- -------------------------------------------------------------
-- 11. 知识点与用户知识点表
-- -------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `knowledge_points` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`        VARCHAR(200)    NOT NULL COMMENT '知识点名称，全局唯一',
  `subject_id`  BIGINT UNSIGNED DEFAULT NULL COMMENT '归属科目ID，可为空',
  `description` TEXT            COMMENT '知识点描述（Markdown）',
  `source`      VARCHAR(20)     NOT NULL DEFAULT 'ai' COMMENT 'ai / manual',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_kp_name` (`name`),
  KEY `idx_kp_subject` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识点全局表';

CREATE TABLE IF NOT EXISTS `user_knowledge_points` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`            BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `knowledge_point_id` BIGINT UNSIGNED NOT NULL COMMENT '知识点ID',
  `source_question_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '来源题目ID',
  `is_important`       TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否重点',
  `is_mastered`        TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否已掌握',
  `mastered_at`        DATETIME        DEFAULT NULL COMMENT '标记已掌握时间',
  `note`               TEXT            COMMENT '个人笔记（V2 预留）',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_kp` (`user_id`, `knowledge_point_id`),
  KEY `idx_ukp_user_important` (`user_id`, `is_important`),
  KEY `idx_ukp_user_mastered` (`user_id`, `is_mastered`),
  KEY `idx_ukp_user_created` (`user_id`, `created_at`),
  KEY `idx_ukp_source_question` (`source_question_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户知识点关联表';
