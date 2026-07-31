-- =============================================================
-- 软考学系平台 - 数据库初始化脚本
-- Database: softteststudyt
-- =============================================================

CREATE DATABASE IF NOT EXISTS `softteststudyt`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `softteststudyt`;

-- =============================================================
-- 1. 用户表
-- =============================================================
CREATE TABLE `users` (
  `id`              BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `username`        VARCHAR(50)      NOT NULL COMMENT '用户名',
  `email`           VARCHAR(100)     NOT NULL COMMENT '邮箱',
  `password_hash`   VARCHAR(255)     NOT NULL COMMENT '密码哈希(bcrypt)',
  `nickname`        VARCHAR(50)      DEFAULT NULL COMMENT '昵称',
  `avatar`          VARCHAR(255)     DEFAULT NULL COMMENT '头像URL',
  `role`            VARCHAR(20)      DEFAULT 'student' COMMENT '角色: student/admin',
  `status`          INT              DEFAULT 1 COMMENT '状态: 1=正常 0=禁用',
  `difficulty`      VARCHAR(20)      DEFAULT '' COMMENT '答题强度: easy/medium/hard',
  `failed_attempts` INT              DEFAULT 0 COMMENT '连续登录失败次数',
  `locked_until`    DATETIME         DEFAULT NULL COMMENT '账号锁定截止时间',
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =============================================================
-- 2. 考试等级表
-- =============================================================
CREATE TABLE `exam_levels` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(50)     NOT NULL COMMENT '等级名称: 初级/中级/高级',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试等级表';

-- =============================================================
-- 3. 考试科目表
-- =============================================================
CREATE TABLE `subjects` (
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

-- =============================================================
-- 4. 子科目表（如综合知识/案例分析）
-- =============================================================
CREATE TABLE `sub_subjects` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id` BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `name`       VARCHAR(200)    NOT NULL COMMENT '子科目名称',
  `sort_order` INT             DEFAULT 0 COMMENT '排序',
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  CONSTRAINT `fk_sub_subjects_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子科目表';

-- =============================================================
-- 5. 章节表
-- =============================================================
CREATE TABLE `chapters` (
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

-- =============================================================
-- 6. 题目表
-- =============================================================
CREATE TABLE `questions` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`     BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `sub_subject_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '所属子科目ID',
  `chapter_id`     BIGINT UNSIGNED DEFAULT 0 COMMENT '所属章节ID',
  `type`           VARCHAR(20)     NOT NULL COMMENT '题型: single/multi/judge/fill',
  `difficulty`     VARCHAR(20)     NOT NULL COMMENT '难度: easy/medium/hard',
  `content`        TEXT            NOT NULL COMMENT '题目内容(支持Markdown)',
  `options`        JSON            DEFAULT NULL COMMENT '选项(JSON数组)',
  `answer`         TEXT            NOT NULL COMMENT '正确答案',
  `analysis`       TEXT            DEFAULT NULL COMMENT '解析',
  `year`           INT             DEFAULT NULL COMMENT '考试年份',
  `status`         INT             DEFAULT 0 COMMENT '状态: 0=待审核 1=已发布',
  `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_sub_subject_id` (`sub_subject_id`),
  KEY `idx_chapter_id` (`chapter_id`),
  CONSTRAINT `fk_questions_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目表';

-- =============================================================
-- 7. 题目知识点关联表
-- =============================================================
CREATE TABLE `question_knowledge` (
  `question_id`       BIGINT UNSIGNED NOT NULL,
  `knowledge_point_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`question_id`, `knowledge_point_id`),
  CONSTRAINT `fk_qk_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目知识点关联表';

-- =============================================================
-- 8. 练习记录表
-- =============================================================
CREATE TABLE `practice_records` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `mode`        VARCHAR(20)     NOT NULL COMMENT '练习模式: chapter/random/wrong',
  `answer`      TEXT            DEFAULT NULL COMMENT '用户答案',
  `is_correct`  INT             DEFAULT NULL COMMENT '是否正确: 0=错误 1=正确',
  `duration`    INT             DEFAULT NULL COMMENT '答题耗时(秒)',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_pr_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pr_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='练习记录表';

-- =============================================================
-- 9. 收藏文件夹表
-- =============================================================
CREATE TABLE `favorite_folders` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `name`       VARCHAR(100)    NOT NULL COMMENT '文件夹名称',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_ff_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏文件夹表';

-- =============================================================
-- 10. 题目收藏表
-- =============================================================
CREATE TABLE `question_favorites` (
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

-- =============================================================
-- 11. 错题表
-- =============================================================
CREATE TABLE `wrong_questions` (
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

-- =============================================================
-- 12. 考试模板表
-- =============================================================
CREATE TABLE `exam_templates` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id`  BIGINT UNSIGNED NOT NULL COMMENT '所属科目ID',
  `name`        VARCHAR(200)    NOT NULL COMMENT '试卷名称',
  `duration`    INT             NOT NULL COMMENT '考试时长(分钟)',
  `total_score` INT             NOT NULL COMMENT '总分',
  `is_public`   INT             DEFAULT 1 COMMENT '是否公开: 1=是 0=否',
  `year`        INT             DEFAULT NULL COMMENT '考试年份',
  `status`      INT             DEFAULT 1 COMMENT '状态: 1=启用 0=禁用',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  CONSTRAINT `fk_et_subject` FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试模板表';

-- =============================================================
-- 13. 考试模板题目关联表
-- =============================================================
CREATE TABLE `exam_template_questions` (
  `template_id` BIGINT UNSIGNED NOT NULL COMMENT '模板ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `sort_order`  INT             DEFAULT 0 COMMENT '题目排序',
  `score`       INT             NOT NULL COMMENT '每题分值',
  PRIMARY KEY (`template_id`, `question_id`),
  KEY `idx_question_id` (`question_id`),
  CONSTRAINT `fk_etq_template` FOREIGN KEY (`template_id`) REFERENCES `exam_templates` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_etq_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='考试模板题目关联表';

-- =============================================================
-- 14. 考试记录表
-- =============================================================
CREATE TABLE `exam_records` (
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

-- =============================================================
-- 15. 考试答题记录表
-- =============================================================
CREATE TABLE `exam_record_answers` (
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

-- =============================================================
-- 种子数据
-- =============================================================

-- 2.1 考试等级
INSERT INTO `exam_levels` (`name`, `sort_order`) VALUES
  ('初级', 1),
  ('中级', 2),
  ('高级', 3);

-- 3.1 考试科目
INSERT INTO `subjects` (`level_id`, `name`, `short_name`, `description`, `sort_order`, `status`) VALUES
  (1, '程序员',          '程序员', '计算机基础与程序设计',     1, 1),
  (1, '网络管理员',       '网管',   '网络基础与管理',           2, 1),
  (1, '信息处理技术员',   '信处',   '信息处理技术',             3, 1),
  (2, '软件设计师',       '软设',   '软件工程与设计',           1, 1),
  (2, '网络工程师',       '网工',   '网络工程与技术',           2, 1),
  (2, '数据库系统工程师', '库工',   '数据库系统设计与管理',     3, 1),
  (2, '信息系统管理工程师','信管',   '信息系统管理',             4, 1),
  (3, '信息系统项目管理师','高项',   '信息系统项目管理',         1, 1),
  (3, '系统分析师',       '系分',   '系统分析与设计',           2, 1),
  (3, '系统架构设计师',   '架构',   '系统架构设计',             3, 1),
  (3, '网络规划设计师',   '网规',   '网络规划与设计',           4, 1);

-- 4.1 子科目（以高级科目和软设为例）
INSERT INTO `sub_subjects` (`subject_id`, `name`, `sort_order`)
SELECT s.`id`, '综合知识', 1 FROM `subjects` s WHERE s.`name` IN ('信息系统项目管理师', '系统分析师', '系统架构设计师', '网络规划设计师');

INSERT INTO `sub_subjects` (`subject_id`, `name`, `sort_order`)
SELECT s.`id`, '案例分析', 2 FROM `subjects` s WHERE s.`name` IN ('信息系统项目管理师', '系统分析师', '系统架构设计师', '网络规划设计师');

INSERT INTO `sub_subjects` (`subject_id`, `name`, `sort_order`)
SELECT s.`id`, '论文', 3 FROM `subjects` s WHERE s.`name` IN ('信息系统项目管理师', '系统分析师', '系统架构设计师', '网络规划设计师');

INSERT INTO `sub_subjects` (`subject_id`, `name`, `sort_order`)
SELECT s.`id`, '基础知识', 1 FROM `subjects` s WHERE s.`name` IN ('软件设计师', '网络工程师', '程序员');

INSERT INTO `sub_subjects` (`subject_id`, `name`, `sort_order`)
SELECT s.`id`, '应用技术', 2 FROM `subjects` s WHERE s.`name` IN ('软件设计师', '网络工程师', '程序员');
