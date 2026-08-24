-- =============================================================
-- 软考学系平台 - 生产数据库一键初始化脚本
-- 文件名: schema.sql
-- 用途: 整合 migrations + 基础种子数据，导入空的生产库
-- 数据源:
--   - backend/migrations/000001_init_schema.up.sql         ~ 000012
--   - backend/cmd/seed/main.go                              (考试等级/科目/子科目/章节/试卷骨架)
-- 兼容: MySQL 5.7+ / 8.0
-- 字符集: utf8mb4 / utf8mb4_unicode_ci
-- 幂等: 是 (重跑安全：所有 ADD INDEX / ADD COLUMN 用 prepared statement 包装；
--       所有数据插入用 NOT EXISTS 模式；UPDATE 用条件限定)
-- 注意:
--   1. 题目数据 (questions) 由 import_real / import_case_study 工具导入，不在本文件
--   2. 用户数据 (users) 由邀请码注册流程产生，不在本文件
--   3. 试卷 (exam_templates) 仅生成模板骨架，题目关联 (exam_template_questions)
--      需要在题目导入完成后由管理后台或脚本绑定
-- 使用: mysql -u root -p < schema.sql
-- =============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- =============================================================
-- 第 1 部分: Schema (整合自 migrations/000001 ~ 000012)
-- =============================================================

-- ---------- 000001: 24 张业务表初始 DDL ----------
-- (原始文件 backend/migrations/000001_init_schema.up.sql)

-- 1. 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id`              BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `username`        VARCHAR(50)      NOT NULL COMMENT '用户名',
  `email`           VARCHAR(100)     NOT NULL COMMENT '邮箱',
  `password_hash`   VARCHAR(255)     NOT NULL COMMENT '密码哈希(bcrypt)',
  `nickname`        VARCHAR(50)      DEFAULT NULL COMMENT '昵称',
  `avatar`          VARCHAR(255)     DEFAULT NULL COMMENT '头像URL',
  `role`            VARCHAR(20)      DEFAULT 'student' COMMENT '角色: student/admin',
  `status`          INT              DEFAULT 1 COMMENT '状态: 1=正常 0=禁用',
  `level_id`        BIGINT UNSIGNED  DEFAULT 0 COMMENT '考试等级ID',
  `subject_id`      BIGINT UNSIGNED  DEFAULT 0 COMMENT '考试科目ID',
  `difficulty`      VARCHAR(20)      DEFAULT '' COMMENT '偏好强度: easy/medium/hard',
  `failed_attempts` INT              DEFAULT 0 COMMENT '登录记录失败次数',
  `locked_until`    DATETIME         DEFAULT NULL COMMENT '账号锁定截止时间',
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 2. 考试等级表
CREATE TABLE IF NOT EXISTS `exam_levels` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(50)     NOT NULL COMMENT '等级名称: 初级/中级/高级',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试等级表';

-- 3. 考试科目表
CREATE TABLE IF NOT EXISTS `subjects` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `level_id`    BIGINT UNSIGNED NOT NULL COMMENT '所属等级ID',
  `name`        VARCHAR(100)    NOT NULL COMMENT '科目名称',
  `short_name`  VARCHAR(20)     DEFAULT NULL COMMENT '简称',
  `description` TEXT            DEFAULT NULL COMMENT '描述',
  `icon`        VARCHAR(255)    DEFAULT NULL COMMENT '图标',
  `sort_order`  INT             DEFAULT 0 COMMENT '排序',
  `status`      INT             DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  PRIMARY KEY (`id`),
  KEY `idx_level_id` (`level_id`),
  CONSTRAINT `fk_subjects_level` FOREIGN KEY (`level_id`) REFERENCES `exam_levels` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试科目表';

-- 4. 子科目表
CREATE TABLE IF NOT EXISTS `sub_subjects` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id` BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `name`       VARCHAR(200)    NOT NULL COMMENT '子科目名称',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  CONSTRAINT `fk_sub_subjects_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子科目表';

-- 5. 章节表
CREATE TABLE IF NOT EXISTS `chapters` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`     BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `sub_subject_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '所属子科目ID',
  `parent_id`      BIGINT UNSIGNED DEFAULT 0 COMMENT '父章节ID（支持树形结构）',
  `name`           VARCHAR(200)    NOT NULL COMMENT '章节名称',
  `sort_order`     INT             DEFAULT 0 COMMENT '排序',
  `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_sub_subject_id` (`sub_subject_id`),
  CONSTRAINT `fk_chapters_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节表';

-- 6. 题目表
CREATE TABLE IF NOT EXISTS `questions` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`     BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `sub_subject_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '所属子科目ID',
  `chapter_id`     BIGINT UNSIGNED DEFAULT 0 COMMENT '所属章节ID',
  `parent_id`      BIGINT UNSIGNED DEFAULT 0 COMMENT '父题目ID: 0=独立题目/大题目, >0=小题目',
  `type`           VARCHAR(20)     NOT NULL COMMENT '题型: single/multi/judge/fill',
  `difficulty`     VARCHAR(20)     NOT NULL COMMENT '难度: easy/medium/hard',
  `content`        TEXT            NOT NULL COMMENT '题目内容(支持Markdown)',
  `case_material`  TEXT            DEFAULT NULL COMMENT '案例素材(案例分析题专用)',
  `options`        JSON            DEFAULT NULL COMMENT '选项(JSON数组)',
  `blank_options`  JSON            DEFAULT NULL COMMENT '多空题专用: 每空独立选项',
  `answer`         TEXT            NOT NULL COMMENT '正确答案',
  `analysis`       TEXT            DEFAULT NULL COMMENT '解析',
  `year`           INT             DEFAULT NULL COMMENT '考试年份',
  `source`         VARCHAR(100)    DEFAULT '' COMMENT '来源: ai=AI生成, seed=种子里程, sa-2023=系统分析师2023真题 等',
  `status`         INT             DEFAULT 0 COMMENT '状态: 0=未上架 1=已发布',
  `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_sub_subject_id` (`sub_subject_id`),
  KEY `idx_chapter_id` (`chapter_id`),
  KEY `idx_parent_id` (`parent_id`),
  CONSTRAINT `fk_questions_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目表';

-- 7. 题目知识点关联表
CREATE TABLE IF NOT EXISTS `question_knowledge` (
  `question_id`        BIGINT UNSIGNED NOT NULL,
  `knowledge_point_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`question_id`, `knowledge_point_id`),
  CONSTRAINT `fk_qk_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目知识点关联表';

-- 8. 练习记录表
CREATE TABLE IF NOT EXISTS `practice_records` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `mode`        VARCHAR(20)     NOT NULL COMMENT '练习模式: chapter/random/wrong',
  `answer`      TEXT            DEFAULT NULL COMMENT '用户答案',
  `is_correct`  INT             DEFAULT NULL COMMENT '是否正确: 0=错误 1=正确',
  `duration`    INT             DEFAULT NULL COMMENT '答题用时(秒)',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_pr_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pr_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='练习记录表';

-- 9. 收藏文件夹表
CREATE TABLE IF NOT EXISTS `favorite_folders` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `name`       VARCHAR(100)    NOT NULL COMMENT '文件夹名称',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_ff_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏文件夹表';

-- 10. 题目收藏表
CREATE TABLE IF NOT EXISTS `question_favorites` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `folder_id`   BIGINT UNSIGNED DEFAULT 0 COMMENT '所属文件夹ID',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question` (`user_id`, `question_id`),
  KEY `idx_folder_id` (`folder_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_qf_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_qf_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目收藏表';

-- 11. 题目标记表
CREATE TABLE IF NOT EXISTS `question_marks` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question` (`user_id`, `question_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_qm_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_qm_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目标记表';

-- 12. 错题表
CREATE TABLE IF NOT EXISTS `wrong_questions` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `question_id`   BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `wrong_count`   INT             DEFAULT 1 COMMENT '错误次数',
  `correct_count` INT             DEFAULT 0 COMMENT '连续正确次数',
  `last_wrong_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近错误时间',
  `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question` (`user_id`, `question_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_wq_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_wq_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='错题表';

-- 13. 考试模板表
CREATE TABLE IF NOT EXISTS `exam_templates` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`    BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `name`          VARCHAR(200)    NOT NULL COMMENT '试卷名称',
  `duration`      INT             NOT NULL COMMENT '考试时长(分钟)',
  `total_score`   INT             NOT NULL COMMENT '总分',
  `question_type` VARCHAR(20)     DEFAULT '' COMMENT '题目类型过滤: 空=不限 essay=仅论文',
  `is_public`     INT             DEFAULT 1 COMMENT '是否公开: 1=是 0=否',
  `year`          INT             DEFAULT NULL COMMENT '考试年份',
  `status`        INT             DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  CONSTRAINT `fk_et_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试模板表';

-- 14. 考试模板题目关联表
CREATE TABLE IF NOT EXISTS `exam_template_questions` (
  `template_id` BIGINT UNSIGNED NOT NULL COMMENT '模板ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `sort_order`  INT             DEFAULT 0 COMMENT '题目排序',
  `score`       INT             NOT NULL COMMENT '每题分值',
  PRIMARY KEY (`template_id`, `question_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_etq_template` FOREIGN KEY (`template_id`) REFERENCES `exam_templates` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_etq_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试模板题目关联表';

-- 15. 考试记录表
CREATE TABLE IF NOT EXISTS `exam_records` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `template_id` BIGINT UNSIGNED NOT NULL COMMENT '模板ID',
  `score`       INT             DEFAULT NULL COMMENT '得分',
  `total_score` INT             DEFAULT NULL COMMENT '总分',
  `duration`    INT             DEFAULT NULL COMMENT '实际用时(秒)',
  `status`      VARCHAR(20)     DEFAULT 'pending' COMMENT '状态: pending=进行中 finished=已完成',
  `started_at`  DATETIME        DEFAULT NULL COMMENT '开始时间',
  `finished_at` DATETIME        DEFAULT NULL COMMENT '完成时间',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_template_id` (`template_id`),
  CONSTRAINT `fk_er_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_er_template` FOREIGN KEY (`template_id`) REFERENCES `exam_templates` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试记录表';

-- 16. 考试答题记录表
CREATE TABLE IF NOT EXISTS `exam_record_answers` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `record_id`   BIGINT UNSIGNED NOT NULL COMMENT '考试记录ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `answer`      TEXT            DEFAULT NULL COMMENT '用户答案',
  `is_correct`  INT             DEFAULT NULL COMMENT '是否正确: 0=错误 1=正确',
  `score`       INT             DEFAULT NULL COMMENT '本题得分',
  PRIMARY KEY (`id`),
  KEY `idx_record_id` (`record_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_era_record` FOREIGN KEY (`record_id`) REFERENCES `exam_records` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_era_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试答题记录表';

-- 17. AI 生成题目表
CREATE TABLE IF NOT EXISTS `ai_generated_questions` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `question_id`     BIGINT UNSIGNED DEFAULT 0,
  `subject_id`      BIGINT UNSIGNED NOT NULL,
  `chapter_id`      BIGINT UNSIGNED DEFAULT 0,
  `type`            VARCHAR(20)     NOT NULL,
  `difficulty`      VARCHAR(20)     NOT NULL,
  `content`         TEXT            NOT NULL,
  `case_material`   TEXT            DEFAULT NULL,
  `options`         JSON            DEFAULT NULL,
  `answer`          VARCHAR(500)    NOT NULL,
  `analysis`        TEXT            DEFAULT NULL,
  `knowledge_point` VARCHAR(200)    DEFAULT '',
  `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_question_id` (`question_id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_chapter_id` (`chapter_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 生成题目表';

-- 18. 论文/案例 AI 评分表
CREATE TABLE IF NOT EXISTS `essay_scores` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`        BIGINT UNSIGNED NOT NULL,
  `question_id`    BIGINT UNSIGNED NOT NULL,
  `question_type`  VARCHAR(20)     DEFAULT NULL COMMENT 'essay/case_study',
  `record_type`    VARCHAR(20)     NOT NULL COMMENT 'practice/exam',
  `record_id`      BIGINT UNSIGNED NOT NULL,
  `exam_answer_id` BIGINT UNSIGNED DEFAULT 0,
  `user_answer`    TEXT            DEFAULT NULL,
  `total_score`    INT             DEFAULT NULL,
  `argument_score`    INT          DEFAULT NULL,
  `structure_score`   INT          DEFAULT NULL,
  `language_score`    INT          DEFAULT NULL,
  `depth_score`       INT          DEFAULT NULL,
  `comment`           TEXT         DEFAULT NULL,
  `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_record` (`user_id`, `record_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文/案例 AI 评分表';

-- 19. 学习资料表
CREATE TABLE IF NOT EXISTS `study_materials` (
  `id`             BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `title`          VARCHAR(200)     NOT NULL COMMENT '资料标题',
  `description`    TEXT             DEFAULT NULL COMMENT '资料简介',
  `subject_id`     BIGINT UNSIGNED  DEFAULT NULL COMMENT '关联科目 subjects.id',
  `cover_url`      VARCHAR(255)     DEFAULT NULL COMMENT '预览图相对路径 /uploads/materials/{id}/cover.png',
  `file_url`       VARCHAR(255)     NOT NULL COMMENT 'xmind 文件相对路径 /uploads/materials/{id}/note.xmind',
  `file_name`      VARCHAR(255)     DEFAULT NULL COMMENT '下载时使用的原始文件名',
  `file_size`      BIGINT           DEFAULT 0 COMMENT '文件大小（字节）',
  `download_count` INT              DEFAULT 0 COMMENT '下载次数',
  `view_count`     INT              DEFAULT 0 COMMENT '预览次数',
  `sort_order`     INT              DEFAULT 0 COMMENT '排序（越大越靠前）',
  `status`         INT              DEFAULT 1 COMMENT '状态: 1=上线 0=下线',
  `created_at`     DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学习资料表';

-- 20. 每日打卡汇总表
CREATE TABLE IF NOT EXISTS `check_ins` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `check_date`        DATE            NOT NULL,
  `total_count`       INT             NOT NULL DEFAULT 10,
  `correct_count`     INT             NOT NULL DEFAULT 0 COMMENT '当日最新一次答对数（重新打卡会更新）',
  `accuracy`          DECIMAL(5,2)    NOT NULL DEFAULT 0 COMMENT '当日最新一次正确率（重新打卡会更新）',
  `duration`          INT             NOT NULL DEFAULT 0 COMMENT '当日最新一次耗时（重新打卡会更新）',
  `status`            TINYINT         NOT NULL DEFAULT 1,
  `rank_correct_count` INT            NOT NULL DEFAULT 0 COMMENT '首次打卡快照：答对数（排名/个人平均口径，重打不变）',
  `rank_accuracy`     DECIMAL(5,2)    NOT NULL DEFAULT 0 COMMENT '首次打卡快照：正确率（排名/个人平均口径，重打不变）',
  `rank_duration`     INT             NOT NULL DEFAULT 0 COMMENT '首次打卡快照：耗时（重打不变）',
  `created_at`        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_date` (`user_id`, `check_date`),
  KEY `idx_date` (`check_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='check-in daily summary';

-- 21. 每日打卡题目快照表
CREATE TABLE IF NOT EXISTS `check_in_questions` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `check_date`  DATE            NOT NULL,
  `question_id` BIGINT UNSIGNED NOT NULL,
  `answer`      TEXT            DEFAULT NULL,
  `is_correct`  TINYINT         NOT NULL DEFAULT 0,
  `duration`    INT             NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_user_date` (`user_id`, `check_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='check-in question snapshot';

-- 22. 学习计划表
CREATE TABLE IF NOT EXISTS `study_plans` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `subject_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `title`      VARCHAR(100)    NOT NULL,
  `daily_goal` INT             NOT NULL DEFAULT 20,
  `start_date` DATE            NOT NULL,
  `end_date`   DATE            NOT NULL,
  `status`     TINYINT         NOT NULL DEFAULT 1,
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='study plan';

-- 23. 复习卡片表（SM-2 遗忘曲线）
CREATE TABLE IF NOT EXISTS `review_cards` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `question_id`      BIGINT UNSIGNED NOT NULL,
  `repetition`       INT             NOT NULL DEFAULT 0,
  `interval_days`    INT             NOT NULL DEFAULT 1,
  `ease_factor`      DECIMAL(4,2)    NOT NULL DEFAULT 2.50,
  `due_date`         DATE            NOT NULL,
  `last_reviewed_at` DATETIME        DEFAULT NULL,
  `status`           TINYINT         NOT NULL DEFAULT 1,
  `created_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question` (`user_id`, `question_id`),
  KEY `idx_due` (`user_id`, `status`, `due_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SM-2 review card';

-- 24. 系统通知表
CREATE TABLE IF NOT EXISTS `notifications` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `type`       VARCHAR(40)     DEFAULT '',
  `title`      VARCHAR(200)    NOT NULL,
  `content`    TEXT            DEFAULT NULL,
  `link`       VARCHAR(500)    DEFAULT '',
  `read`       TINYINT(1)      DEFAULT 0,
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_read` (`read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表';


-- ---------- 000002: practice_records 复合索引 (已带幂等) ----------
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'practice_records'
    AND index_name = 'idx_user_created_at'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE practice_records ADD INDEX idx_user_created_at (user_id, created_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'practice_records'
    AND index_name = 'idx_user_mode_created'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE practice_records ADD INDEX idx_user_mode_created (user_id, mode, created_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- ---------- 000003: questions 复合索引 (改幂等) ----------
-- 原始: 直接 ADD INDEX, 重复执行会报错. 此处全部包装为 prepared statement
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'questions'
    AND index_name = 'idx_subject_status_type'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE questions ADD INDEX idx_subject_status_type (subject_id, status, type)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'questions'
    AND index_name = 'idx_chapter_status'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE questions ADD INDEX idx_chapter_status (chapter_id, status)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'questions'
    AND index_name = 'idx_year_status_type'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE questions ADD INDEX idx_year_status_type (year, status, type)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- ---------- 000004: 业务表 user_id+时间 索引 (改幂等) ----------
-- wrong_questions(idx_user_lastwrong)
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'wrong_questions'
    AND index_name = 'idx_user_lastwrong'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE wrong_questions ADD INDEX idx_user_lastwrong (user_id, last_wrong_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- check_ins(idx_user_status_date)
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'check_ins'
    AND index_name = 'idx_user_status_date'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE check_ins ADD INDEX idx_user_status_date (user_id, status, check_date)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- exam_records(idx_user_status_created)
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'exam_records'
    AND index_name = 'idx_user_status_created'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE exam_records ADD INDEX idx_user_status_created (user_id, status, created_at)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- exam_record_answers(uk_record_question) - 唯一约束
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'exam_record_answers'
    AND index_name = 'uk_record_question'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE exam_record_answers ADD UNIQUE INDEX uk_record_question (record_id, question_id)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- ---------- 000005: check_ins 排名快照回填 (UPDATE 已带 WHERE 条件, 天然幂等) ----------
UPDATE `check_ins`
   SET `rank_correct_count` = `correct_count`,
       `rank_accuracy`      = `accuracy`,
       `rank_duration`      = `duration`
 WHERE `rank_correct_count` = 0
   AND `rank_accuracy` = 0;


-- ---------- 000006: ai_generated_questions 加 batch_id (改幂等) ----------
-- 1. ADD COLUMN batch_id (幂等)
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE table_schema = DATABASE()
    AND table_name = 'ai_generated_questions'
    AND column_name = 'batch_id'
);
SET @ddl := IF(@col_exists = 0,
  'ALTER TABLE `ai_generated_questions` ADD COLUMN `batch_id` VARCHAR(64) NOT NULL DEFAULT '''' AFTER `question_id`',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. ADD INDEX idx_ai_generated_batch_id (幂等)
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'ai_generated_questions'
    AND index_name = 'idx_ai_generated_batch_id'
);
SET @ddl := IF(@has_idx = 0,
  'ALTER TABLE ai_generated_questions ADD INDEX idx_ai_generated_batch_id (batch_id)',
  'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. 历史数据回填 batch_id (UPDATE 已带 WHERE 条件, 天然幂等)
UPDATE `ai_generated_questions` AS a
  JOIN (
    SELECT `user_id`, `subject_id`, DATE_FORMAT(`created_at`, '%Y%m%d%H%i') AS `ts`
      FROM `ai_generated_questions`
     WHERE `batch_id` = ''
     GROUP BY `user_id`, `subject_id`, `ts`
  ) AS b
    ON a.`user_id` = b.`user_id`
   AND a.`subject_id` = b.`subject_id`
   AND DATE_FORMAT(a.`created_at`, '%Y%m%d%H%i') = b.`ts`
   SET a.`batch_id` = CONCAT('legacy-', b.`user_id`, '-', b.`subject_id`, '-', b.`ts`)
 WHERE a.`batch_id` = '';


-- ---------- 000007: ai_generated_tasks 表 ----------
CREATE TABLE IF NOT EXISTS `ai_generated_tasks` (
  `id`               BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `task_id`          VARCHAR(64)      NOT NULL                COMMENT '前端可见的 task_id（16 字节 hex）',
  `user_id`          BIGINT UNSIGNED  NOT NULL                COMMENT '归属用户',
  `status`           VARCHAR(20)      NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed',
  `question_type`    VARCHAR(20)      NOT NULL                COMMENT 'single/multi/judge/case_study/essay',
  `chapter_id`       BIGINT UNSIGNED  NOT NULL DEFAULT 0,
  `chapter_name`     VARCHAR(200)     NOT NULL DEFAULT '',
  `difficulty`       VARCHAR(20)      NOT NULL DEFAULT '',
  `count`            INT              NOT NULL DEFAULT 0,
  `error`            TEXT             DEFAULT NULL             COMMENT '失败原因',
  `result_questions` JSON             DEFAULT NULL             COMMENT '成功后保存题目 JSON 数组（与 ai_generated_questions 解耦）',
  `created_at`       DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `finished_at`      DATETIME         DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_id` (`task_id`),
  KEY `idx_user_status_updated` (`user_id`, `status`, `updated_at`),
  KEY `idx_status_updated` (`status`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='AI 异步出题任务状态';


-- ---------- 000008: email_verification_codes 表 ----------
CREATE TABLE IF NOT EXISTS `email_verification_codes` (
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


-- ---------- 000009: users 加 email_verified (原始文件已幂等) ----------
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
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- ---------- 000010: invitation_codes 表 ----------
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


-- ---------- 000011: 知识点相关表 ----------
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


-- ---------- 000012: ai_generated_tasks 加 subject_id (改幂等) ----------
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ai_generated_tasks'
    AND COLUMN_NAME = 'subject_id'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `ai_generated_tasks` ADD COLUMN `subject_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
     COMMENT ''提交时所属科目；用于历史列表补全 subject_name'' AFTER `user_id`',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;


-- =============================================================
-- 第 2 部分: 基础种子数据 (整合自 backend/cmd/seed/main.go)
-- 数据插入使用 INSERT ... SELECT ... WHERE NOT EXISTS 模式
-- (因 exam_levels/subjects 等表无 unique (name) 约束, 改用 NOT EXISTS 保证幂等)
-- 显式指定 id 以保证外部引用关系稳定
-- =============================================================

-- 2.1 考试等级
INSERT INTO `exam_levels` (`id`, `name`, `sort_order`)
SELECT * FROM (
  SELECT 1 AS id, '初级' AS name, 1 AS sort_order UNION ALL
  SELECT 2, '中级', 2 UNION ALL
  SELECT 3, '高级', 3
) v
WHERE NOT EXISTS (SELECT 1 FROM `exam_levels` WHERE `name` = v.`name`);


-- 2.2 考试科目 (11 条)
INSERT INTO `subjects` (`id`, `level_id`, `name`, `short_name`, `description`, `sort_order`, `status`)
SELECT * FROM (
  -- 初级 (level_id=1)
  SELECT  1 AS id, 1 AS level_id, '程序员'              AS name, '程序员' AS short_name, '计算机基础与程序设计' AS description, 1 AS sort_order, 1 AS status UNION ALL
  SELECT  2,        1,              '网络管理员',         '网管',     '网络基础与管理',          2,                  1 UNION ALL
  SELECT  3,        1,              '信息处理技术员',     '信处',     '信息处理技术',            3,                  1 UNION ALL
  -- 中级 (level_id=2)
  SELECT  4,        2,              '软件设计师',         '软设',     '软件工程与设计',          1,                  1 UNION ALL
  SELECT  5,        2,              '网络工程师',         '网工',     '网络工程与技术',          2,                  1 UNION ALL
  SELECT  6,        2,              '数据库系统工程师',   '库工',     '数据库系统设计与管理',    3,                  1 UNION ALL
  SELECT  7,        2,              '信息系统管理工程师', '信管',     '信息系统管理',            4,                  1 UNION ALL
  -- 高级 (level_id=3)
  SELECT  8,        3,              '信息系统项目管理师', '高项',     '信息系统项目管理',        1,                  1 UNION ALL
  SELECT  9,        3,              '系统分析师',         '系分',     '系统分析与设计',          2,                  1 UNION ALL
  SELECT 10,        3,              '系统架构设计师',     '架构',     '系统架构设计',            3,                  1 UNION ALL
  SELECT 11,        3,              '网络规划设计师',     '网规',     '网络规划与设计',          4,                  1
) v
WHERE NOT EXISTS (SELECT 1 FROM `subjects` WHERE `name` = v.`name` AND `level_id` = v.`level_id`);


-- 2.3 子科目
-- 高级 4 科目各 3 个: 综合知识 / 案例分析 / 论文
-- 初/中级 7 科目各 2 个: 基础知识 / 应用技术
-- 共 4*3 + 7*2 = 26 条
INSERT INTO `sub_subjects` (`id`, `subject_id`, `name`, `sort_order`)
SELECT * FROM (
  -- 高级: 信息系统项目管理师 (subject_id=8) → 综合知识/案例分析/论文
  SELECT  1 AS id,  8 AS subject_id, '综合知识'   AS name, 1 AS sort_order UNION ALL
  SELECT  2,         8,                '案例分析',                  2 UNION ALL
  SELECT  3,         8,                '论文',                      3 UNION ALL
  -- 高级: 系统分析师 (subject_id=9) → 综合知识/案例分析/论文
  SELECT  4,         9,                '综合知识',                  1 UNION ALL
  SELECT  5,         9,                '案例分析',                  2 UNION ALL
  SELECT  6,         9,                '论文',                      3 UNION ALL
  -- 高级: 系统架构设计师 (subject_id=10) → 综合知识/案例分析/论文
  SELECT  7,        10,                '综合知识',                  1 UNION ALL
  SELECT  8,        10,                '案例分析',                  2 UNION ALL
  SELECT  9,        10,                '论文',                      3 UNION ALL
  -- 高级: 网络规划设计师 (subject_id=11) → 综合知识/案例分析/论文
  SELECT 10,        11,                '综合知识',                  1 UNION ALL
  SELECT 11,        11,                '案例分析',                  2 UNION ALL
  SELECT 12,        11,                '论文',                      3 UNION ALL
  -- 初级: 程序员 (subject_id=1) → 基础知识/应用技术
  SELECT 13,         1,                '基础知识',                  1 UNION ALL
  SELECT 14,         1,                '应用技术',                  2 UNION ALL
  -- 初级: 网络管理员 (subject_id=2) → 基础知识/应用技术
  SELECT 15,         2,                '基础知识',                  1 UNION ALL
  SELECT 16,         2,                '应用技术',                  2 UNION ALL
  -- 初级: 信息处理技术员 (subject_id=3) → 基础知识/应用技术
  SELECT 17,         3,                '基础知识',                  1 UNION ALL
  SELECT 18,         3,                '应用技术',                  2 UNION ALL
  -- 中级: 软件设计师 (subject_id=4) → 基础知识/应用技术
  SELECT 19,         4,                '基础知识',                  1 UNION ALL
  SELECT 20,         4,                '应用技术',                  2 UNION ALL
  -- 中级: 网络工程师 (subject_id=5) → 基础知识/应用技术
  SELECT 21,         5,                '基础知识',                  1 UNION ALL
  SELECT 22,         5,                '应用技术',                  2 UNION ALL
  -- 中级: 数据库系统工程师 (subject_id=6) → 基础知识/应用技术
  SELECT 23,         6,                '基础知识',                  1 UNION ALL
  SELECT 24,         6,                '应用技术',                  2 UNION ALL
  -- 中级: 信息系统管理工程师 (subject_id=7) → 基础知识/应用技术
  SELECT 25,         7,                '基础知识',                  1 UNION ALL
  SELECT 26,         7,                '应用技术',                  2
) v
WHERE NOT EXISTS (SELECT 1 FROM `sub_subjects` WHERE `subject_id` = v.`subject_id` AND `name` = v.`name`);


-- 2.4 章节
-- 按 (subject_name, sub_subject_name) 决定的章节列表
-- 规则: 综合知识/基础知识 → 用 subject 对应章节; 案例分析/应用技术 → 综合案例一二三; 论文 → 论文写作专题
-- 共 25 个 subject×sub_subject 组合, 章节数随科目变化
INSERT INTO `chapters` (`id`, `subject_id`, `sub_subject_id`, `parent_id`, `name`, `sort_order`)
SELECT * FROM (
  -- ============ 信息系统项目管理师 (subject_id=8) ============
  -- 综合知识 (sub_subject_id=1) - 22 章
  SELECT   1 AS id,  8 AS subject_id,  1 AS sub_subject_id, 0 AS parent_id, '信息化和信息系统'             AS name,  1 AS sort_order UNION ALL
  SELECT   2,         8,                1,                   0,             '信息技术发展',                                2 UNION ALL
  SELECT   3,         8,                1,                   0,             '信息系统服务',                                3 UNION ALL
  SELECT   4,         8,                1,                   0,             '信息系统治理',                                4 UNION ALL
  SELECT   5,         8,                1,                   0,             '信息系统工程',                                5 UNION ALL
  SELECT   6,         8,                1,                   0,             '项目管理概论',                                6 UNION ALL
  SELECT   7,         8,                1,                   0,             '项目立项与招投标管理',                        7 UNION ALL
  SELECT   8,         8,                1,                   0,             '项目整体管理',                                8 UNION ALL
  SELECT   9,         8,                1,                   0,             '项目范围管理',                                9 UNION ALL
  SELECT  10,         8,                1,                   0,             '项目进度管理',                               10 UNION ALL
  SELECT  11,         8,                1,                   0,             '项目成本管理',                               11 UNION ALL
  SELECT  12,         8,                1,                   0,             '项目质量管理',                               12 UNION ALL
  SELECT  13,         8,                1,                   0,             '项目资源管理',                               13 UNION ALL
  SELECT  14,         8,                1,                   0,             '项目沟通管理',                               14 UNION ALL
  SELECT  15,         8,                1,                   0,             '项目风险管理',                               15 UNION ALL
  SELECT  16,         8,                1,                   0,             '项目采购管理',                               16 UNION ALL
  SELECT  17,         8,                1,                   0,             '项目合同管理',                               17 UNION ALL
  SELECT  18,         8,                1,                   0,             '项目变更管理',                               18 UNION ALL
  SELECT  19,         8,                1,                   0,             '组织通用治理',                               19 UNION ALL
  SELECT  20,         8,                1,                   0,             '组织通用管理',                               20 UNION ALL
  SELECT  21,         8,                1,                   0,             '流程管理',                                   21 UNION ALL
  SELECT  22,         8,                1,                   0,             '项目集、项目组合与组织战略',                 22 UNION ALL
  -- 案例分析 (sub_subject_id=2)
  SELECT  23,         8,                2,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  24,         8,                2,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  25,         8,                2,                   0,             '综合案例（三）',                              3 UNION ALL
  -- 论文 (sub_subject_id=3)
  SELECT  26,         8,                3,                   0,             '论文写作专题',                                1 UNION ALL

  -- ============ 系统分析师 (subject_id=9) ============
  -- 综合知识 (sub_subject_id=4) - 18 章
  -- 注: 使用独立 ID 段 200~217, 避免与其他科目章节的固定 ID 冲突
  SELECT 200,         9,                4,                   0,             '绪论',                                        1 UNION ALL
  SELECT 201,         9,                4,                   0,             '法律法规与标准化',                            2 UNION ALL
  SELECT 202,         9,                4,                   0,             '数学基础',                                    3 UNION ALL
  SELECT 203,         9,                4,                   0,             '运筹学基础',                                  4 UNION ALL
  SELECT 204,         9,                4,                   0,             '数据结构与算法',                              5 UNION ALL
  SELECT 205,         9,                4,                   0,             '计算机组成与体系结构',                        6 UNION ALL
  SELECT 206,         9,                4,                   0,             '操作系统',                                    7 UNION ALL
  SELECT 207,         9,                4,                   0,             '程序设计语言与语言处理',                      8 UNION ALL
  SELECT 208,         9,                4,                   0,             '嵌入式系统',                                  9 UNION ALL
  SELECT 209,         9,                4,                   0,             '计算机网络',                                 10 UNION ALL
  SELECT 210,         9,                4,                   0,             '分布式系统与中间件',                         11 UNION ALL
  SELECT 211,         9,                4,                   0,             '多媒体基础',                                 12 UNION ALL
  SELECT 212,         9,                4,                   0,             '数据库系统',                                 13 UNION ALL
  SELECT 213,         9,                4,                   0,             '企业信息化',                                 14 UNION ALL
  SELECT 214,         9,                4,                   0,             '软件工程',                                   15 UNION ALL
  SELECT 215,         9,                4,                   0,             '面向对象方法与设计模式',                     16 UNION ALL
  SELECT 216,         9,                4,                   0,             '项目管理',                                   17 UNION ALL
  SELECT 217,         9,                4,                   0,             '信息安全',                                   18 UNION ALL
  -- 案例分析 (sub_subject_id=5)
  SELECT  36,         9,                5,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  37,         9,                5,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  38,         9,                5,                   0,             '综合案例（三）',                              3 UNION ALL
  -- 论文 (sub_subject_id=6)
  SELECT  39,         9,                6,                   0,             '论文写作专题',                                1 UNION ALL

  -- ============ 系统架构设计师 (subject_id=10) ============
  -- 综合知识 (sub_subject_id=7) - 4 章
  SELECT  40,        10,                7,                   0,             '架构设计基础',                                1 UNION ALL
  SELECT  41,        10,                7,                   0,             '软件架构风格',                                2 UNION ALL
  SELECT  42,        10,                7,                   0,             '系统质量与性能',                              3 UNION ALL
  SELECT  43,        10,                7,                   0,             '云与大数据架构',                              4 UNION ALL
  -- 案例分析 (sub_subject_id=8)
  SELECT  44,        10,                8,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  45,        10,                8,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  46,        10,                8,                   0,             '综合案例（三）',                              3 UNION ALL
  -- 论文 (sub_subject_id=9)
  SELECT  47,        10,                9,                   0,             '论文写作专题',                                1 UNION ALL

  -- ============ 网络规划设计师 (subject_id=11) ============
  -- 综合知识 (sub_subject_id=10) - 3 章
  SELECT  48,        11,               10,                   0,             '网络规划基础',                                1 UNION ALL
  SELECT  49,        11,               10,                   0,             '网络拓扑设计',                                2 UNION ALL
  SELECT  50,        11,               10,                   0,             '网络性能与安全',                              3 UNION ALL
  -- 案例分析 (sub_subject_id=11)
  SELECT  51,        11,               11,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  52,        11,               11,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  53,        11,               11,                   0,             '综合案例（三）',                              3 UNION ALL
  -- 论文 (sub_subject_id=12)
  SELECT  54,        11,               12,                   0,             '论文写作专题',                                1 UNION ALL

  -- ============ 程序员 (subject_id=1) ============
  -- 基础知识 (sub_subject_id=13) - 6 章
  SELECT  55,         1,               13,                   0,             '计算机基础',                                  1 UNION ALL
  SELECT  56,         1,               13,                   0,             '操作系统',                                    2 UNION ALL
  SELECT  57,         1,               13,                   0,             '程序设计基础',                                3 UNION ALL
  SELECT  58,         1,               13,                   0,             '数据结构与算法',                              4 UNION ALL
  SELECT  59,         1,               13,                   0,             '数据库基础',                                  5 UNION ALL
  SELECT  60,         1,               13,                   0,             '网络与信息安全',                              6 UNION ALL
  -- 应用技术 (sub_subject_id=14)
  SELECT  61,         1,               14,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  62,         1,               14,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  63,         1,               14,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 网络管理员 (subject_id=2) ============
  -- 基础知识 (sub_subject_id=15) - 3 章
  SELECT  64,         2,               15,                   0,             '网络基础',                                    1 UNION ALL
  SELECT  65,         2,               15,                   0,             '网络设备与配置',                              2 UNION ALL
  SELECT  66,         2,               15,                   0,             '网络安全与管理',                              3 UNION ALL
  -- 应用技术 (sub_subject_id=16)
  SELECT  67,         2,               16,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  68,         2,               16,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  69,         2,               16,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 信息处理技术员 (subject_id=3) ============
  -- 基础知识 (sub_subject_id=17) - 3 章
  SELECT  70,         3,               17,                   0,             '信息处理基础',                                1 UNION ALL
  SELECT  71,         3,               17,                   0,             '办公软件应用',                                2 UNION ALL
  SELECT  72,         3,               17,                   0,             '信息安全与法规',                              3 UNION ALL
  -- 应用技术 (sub_subject_id=18)
  SELECT  73,         3,               18,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  74,         3,               18,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  75,         3,               18,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 软件设计师 (subject_id=4) ============
  -- 基础知识 (sub_subject_id=19) - 8 章
  SELECT  76,         4,               19,                   0,             '计算机组成与体系结构',                        1 UNION ALL
  SELECT  77,         4,               19,                   0,             '操作系统',                                    2 UNION ALL
  SELECT  78,         4,               19,                   0,             '数据结构与算法',                              3 UNION ALL
  SELECT  79,         4,               19,                   0,             '程序设计语言',                                4 UNION ALL
  SELECT  80,         4,               19,                   0,             '软件工程基础',                                5 UNION ALL
  SELECT  81,         4,               19,                   0,             '数据库系统',                                  6 UNION ALL
  SELECT  82,         4,               19,                   0,             '计算机网络',                                  7 UNION ALL
  SELECT  83,         4,               19,                   0,             '信息安全',                                    8 UNION ALL
  -- 应用技术 (sub_subject_id=20)
  SELECT  84,         4,               20,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  85,         4,               20,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  86,         4,               20,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 网络工程师 (subject_id=5) ============
  -- 基础知识 (sub_subject_id=21) - 6 章
  SELECT  87,         5,               21,                   0,             '计算机网络概述',                              1 UNION ALL
  SELECT  88,         5,               21,                   0,             '数据通信基础',                                2 UNION ALL
  SELECT  89,         5,               21,                   0,             '局域网与以太网',                              3 UNION ALL
  SELECT  90,         5,               21,                   0,             '网络互连与路由',                              4 UNION ALL
  SELECT  91,         5,               21,                   0,             '网络安全',                                    5 UNION ALL
  SELECT  92,         5,               21,                   0,             '网络管理与维护',                              6 UNION ALL
  -- 应用技术 (sub_subject_id=22)
  SELECT  93,         5,               22,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT  94,         5,               22,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT  95,         5,               22,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 数据库系统工程师 (subject_id=6) ============
  -- 基础知识 (sub_subject_id=23) - 6 章
  SELECT  96,         6,               23,                   0,             '数据库系统概论',                              1 UNION ALL
  SELECT  97,         6,               23,                   0,             '关系数据模型',                                2 UNION ALL
  SELECT  98,         6,               23,                   0,             'SQL语言',                                     3 UNION ALL
  SELECT  99,         6,               23,                   0,             '数据库设计',                                  4 UNION ALL
  SELECT 100,         6,               23,                   0,             '事务管理与并发控制',                          5 UNION ALL
  SELECT 101,         6,               23,                   0,             '数据库管理与安全',                            6 UNION ALL
  -- 应用技术 (sub_subject_id=24)
  SELECT 102,         6,               24,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT 103,         6,               24,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT 104,         6,               24,                   0,             '综合案例（三）',                              3 UNION ALL

  -- ============ 信息系统管理工程师 (subject_id=7) ============
  -- 基础知识 (sub_subject_id=25) - 3 章
  SELECT 105,         7,               25,                   0,             '信息系统基础',                                1 UNION ALL
  SELECT 106,         7,               25,                   0,             '系统运维管理',                                2 UNION ALL
  SELECT 107,         7,               25,                   0,             'IT服务管理',                                  3 UNION ALL
  -- 应用技术 (sub_subject_id=26)
  SELECT 108,         7,               26,                   0,             '综合案例（一）',                              1 UNION ALL
  SELECT 109,         7,               26,                   0,             '综合案例（二）',                              2 UNION ALL
  SELECT 110,         7,               26,                   0,             '综合案例（三）',                              3
) v
WHERE NOT EXISTS (SELECT 1 FROM `chapters` WHERE `subject_id` = v.`subject_id` AND `sub_subject_id` = v.`sub_subject_id` AND `name` = v.`name`);


-- 2.5 试卷模板骨架
-- 11 个科目各生成 1 个 "X模拟试卷（一）" 和含论文的 4 个高级科目额外生成 "X论文写作卷"
-- 此处仅插入模板骨架, exam_template_questions 题目关联待题目导入后由后台绑定
-- total_score 留 0, 由绑定题目后回填
INSERT INTO `exam_templates` (`id`, `subject_id`, `name`, `duration`, `total_score`, `question_type`, `is_public`, `year`, `status`)
SELECT * FROM (
  -- 模拟试卷（11 个科目各 1 份）
  SELECT   1 AS id,  1 AS subject_id, '程序员模拟试卷（一）'              AS name,  90 AS duration, 0 AS total_score, '' AS question_type, 1 AS is_public, 2024 AS year, 1 AS status UNION ALL
  SELECT   2,         2,                '网络管理员模拟试卷（一）',                          90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   3,         3,                '信息处理技术员模拟试卷（一）',                      90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   4,         4,                '软件设计师模拟试卷（一）',                          90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   5,         5,                '网络工程师模拟试卷（一）',                          90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   6,         6,                '数据库系统工程师模拟试卷（一）',                    90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   7,         7,                '信息系统管理工程师模拟试卷（一）',                  90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   8,         8,                '信息系统项目管理师模拟试卷（一）',                  90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT   9,         9,                '系统分析师模拟试卷（一）',                          90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT  10,        10,                '系统架构设计师模拟试卷（一）',                      90,    0,             '',         1,               2024,         1 UNION ALL
  SELECT  11,        11,                '网络规划设计师模拟试卷（一）',                      90,    0,             '',         1,               2024,         1 UNION ALL
  -- 论文写作卷（4 个高级科目各 1 份）
  SELECT  12,         8,                '信息系统项目管理师论文写作卷',                     150,    1,             'essay',    1,               2024,         1 UNION ALL
  SELECT  13,         9,                '系统分析师论文写作卷',                              150,    1,             'essay',    1,               2024,         1 UNION ALL
  SELECT  14,        10,                '系统架构设计师论文写作卷',                          150,    1,             'essay',    1,               2024,         1 UNION ALL
  SELECT  15,        11,                '网络规划设计师论文写作卷',                          150,    1,             'essay',    1,               2024,         1
) v
WHERE NOT EXISTS (SELECT 1 FROM `exam_templates` WHERE `subject_id` = v.`subject_id` AND `name` = v.`name`);


SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================
-- 初始化完成
-- 表数: 24 (000001) + 3 (000007/000008/000010/000011) = 28 张业务表
-- 种子数据:
--   - exam_levels: 3
--   - subjects: 11
--   - sub_subjects: 26
--   - chapters: 110
--   - exam_templates: 15 (11 模拟卷 + 4 论文卷)
-- 后续操作:
--   1. 运行 import_real / import_case_study 导入题目
--   2. 在后台管理界面绑定 exam_template_questions
--   3. 通过邀请码注册 admin 账户
-- =============================================================
