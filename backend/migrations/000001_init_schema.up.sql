-- =============================================================
-- 000001_init_schema.up.sql
-- 24 张业务表初始 DDL（从原 database.sql 拆出）
-- =============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -------------------------------------------------------------
-- 1. 用户表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 2. 考试等级表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `exam_levels` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(50)     NOT NULL COMMENT '等级名称: 初级/中级/高级',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试等级表';

-- -------------------------------------------------------------
-- 3. 考试科目表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 4. 子科目表（综合知识/案例分析/论文 等）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `sub_subjects` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id` BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `name`       VARCHAR(200)    NOT NULL COMMENT '子科目名称',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  CONSTRAINT `fk_sub_subjects_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子科目表';

-- -------------------------------------------------------------
-- 5. 章节表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 6. 题目表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `questions` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`     BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `sub_subject_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '所属子科目ID',
  `chapter_id`     BIGINT UNSIGNED DEFAULT 0 COMMENT '所属章节ID',
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
  CONSTRAINT `fk_questions_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目表';

-- -------------------------------------------------------------
-- 7. 题目知识点关联表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `question_knowledge` (
  `question_id`        BIGINT UNSIGNED NOT NULL,
  `knowledge_point_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`question_id`, `knowledge_point_id`),
  CONSTRAINT `fk_qk_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目知识点关联表';

-- -------------------------------------------------------------
-- 8. 练习记录表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 9. 收藏文件夹表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `favorite_folders` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `name`       VARCHAR(100)    NOT NULL COMMENT '文件夹名称',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_ff_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏文件夹表';

-- -------------------------------------------------------------
-- 10. 题目收藏表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 11. 题目标记表（重点题）
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 12. 错题表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 13. 考试模板表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 14. 考试模板题目关联表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 15. 考试记录表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 16. 考试答题记录表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 17. AI 生成题目表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 18. 论文/案例 AI 评分表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 19. 学习资料表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 20. 每日打卡汇总表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 21. 每日打卡题目快照表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 22. 学习计划表
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 23. 复习卡片表（SM-2 遗忘曲线）
-- -------------------------------------------------------------
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

-- -------------------------------------------------------------
-- 24. 系统通知表
-- -------------------------------------------------------------
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

SET FOREIGN_KEY_CHECKS = 1;
