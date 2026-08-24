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
  `parent_id`      BIGINT UNSIGNED DEFAULT 0 COMMENT '父题目ID: 0=独立题目/大题目, >0=小题目',
  `type`           VARCHAR(20)     NOT NULL COMMENT '题型: single/multi/judge/fill',
  `difficulty`     VARCHAR(20)     NOT NULL COMMENT '难度: easy/medium/hard',
  `content`        TEXT            NOT NULL COMMENT '题目内容(支持Markdown)',
  `case_material`  TEXT            DEFAULT NULL COMMENT '案例材料(案例分析题专用)',
  `options`        JSON            DEFAULT NULL COMMENT '选项(JSON数组)',
  `answer`         TEXT            NOT NULL COMMENT '正确答案',
  `analysis`       TEXT            DEFAULT NULL COMMENT '解析',
  `year`           INT             DEFAULT NULL COMMENT '考试年份',
  `source`         VARCHAR(100)    DEFAULT '' COMMENT '题源: ai=AI生成, seed=内置种子, sa-2023=系统分析师2023真题 等',
  `status`         INT             DEFAULT 0 COMMENT '状态: 0=待审核 1=已发布',
  `created_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_subject_id` (`subject_id`),
  KEY `idx_sub_subject_id` (`sub_subject_id`),
  KEY `idx_chapter_id` (`chapter_id`),
  KEY `idx_parent_id` (`parent_id`),
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
-- 11. 题目标记表（标记疑问）
-- =============================================================
CREATE TABLE `question_marks` (
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

-- =============================================================
-- 12. 错题表
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
-- 13. 考试模板表
-- =============================================================
CREATE TABLE `exam_templates` (
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

-- =============================================================
-- 14. 考试模板题目关联表
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
-- 15. 考试记录表
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
-- 16. 考试答题记录表
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

-- 4.2 系统分析师-案例分析 示例真题（1 个大题目 + 3 个小题目）
SET @sa_case_ss = (SELECT id FROM `sub_subjects` WHERE `name` = '案例分析' AND `subject_id` = (SELECT id FROM `subjects` WHERE `name` = '系统分析师') LIMIT 1);
SET @sa_id = (SELECT id FROM `subjects` WHERE `name` = '系统分析师' LIMIT 1);

-- 大题目（父题）：包含案例材料
INSERT INTO `questions` (`subject_id`, `sub_subject_id`, `chapter_id`, `parent_id`, `type`, `difficulty`, `content`, `case_material`, `options`, `answer`, `analysis`, `year`, `status`)
VALUES
(@sa_id, @sa_case_ss, 0, 0, 'case_study', 'medium',
 '某大型企业拟构建一套跨部门协同办公系统，要求支持高并发访问、便于后期功能扩展，并能够与既有的 ERP、CRM 系统对接。架构师在方案评审时提出采用经典的分层架构风格，将系统划分为表示层、业务逻辑层和数据访问层，各层之间通过明确定义的接口进行通信。请阅读上述案例材料，回答以下问题。',
 '某大型企业拟构建一套跨部门协同办公系统，要求支持高并发访问、便于后期功能扩展，并能够与既有的 ERP、CRM 系统对接。架构师在方案评审时提出采用经典的分层架构风格，将系统划分为表示层、业务逻辑层和数据访问层，各层之间通过明确定义的接口进行通信。',
 '{}', '', '',
 2023, 1);

SET @case_parent_id = LAST_INSERT_ID();

-- 小题目 1：简答题
INSERT INTO `questions` (`subject_id`, `sub_subject_id`, `chapter_id`, `parent_id`, `type`, `difficulty`, `content`, `case_material`, `options`, `answer`, `analysis`, `year`, `status`)
VALUES
(@sa_id, @sa_case_ss, 0, @case_parent_id, 'short', 'medium',
 '【问题 1】结合上述说明，请用 200 字以内简要说明该系统采用分层架构的优点。',
 '', '{}', '',
 '分层架构将关注点分离，表示层负责交互、业务逻辑层封装核心规则、数据访问层屏蔽存储细节；层间通过接口解耦，便于独立开发与测试，也更有利于功能扩展与既有系统对接。',
 2023, 1);

-- 小题目 2：简答题
INSERT INTO `questions` (`subject_id`, `sub_subject_id`, `chapter_id`, `parent_id`, `type`, `difficulty`, `content`, `case_material`, `options`, `answer`, `analysis`, `year`, `status`)
VALUES
(@sa_id, @sa_case_ss, 0, @case_parent_id, 'short', 'medium',
 '【问题 2】请指出该分层架构在性能与扩展性方面可能存在的两个不足，并给出改进建议。',
 '', '{}', '',
 '不足：①层间调用链长，跨层请求需逐层转发，高并发下存在性能损耗；②严格分层可能形成数据访问瓶颈。改进：引入缓存层（如 Redis）减少数据库访问；对热点服务采用异步消息解耦；必要时以微服务拆分高内聚模块。',
 2023, 1);

-- 小题目 3：简答题
INSERT INTO `questions` (`subject_id`, `sub_subject_id`, `chapter_id`, `parent_id`, `type`, `difficulty`, `content`, `case_material`, `options`, `answer`, `analysis`, `year`, `status`)
VALUES
(@sa_id, @sa_case_ss, 0, @case_parent_id, 'short', 'hard',
 '【问题 3】若需与既有 ERP、CRM 系统对接，请说明应采用的集成方式及需要注意的数据一致性问题。',
 '', '{}', '',
 '集成方式：优先采用基于 API 的集成（REST/消息中间件），通过企业服务总线或 API 网关统一编排；数据一致性：跨系统写操作采用最终一致性，通过事务补偿、幂等设计与消息确认机制保证；需统一主数据口径并做对账。',
 2023, 1);

-- =============================================================
-- 学习资料表（xmind 笔记分享）
-- 文件约定: backend/data/uploads/materials/{id}/cover.png + note.xmind
-- =============================================================
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

-- =============================================================
-- Check-in / Study Plan / Review Card / Ranking
-- =============================================================
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

SET @has_idx := (SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'practice_records' AND index_name = 'idx_user_created_at');
SET @ddl := IF(@has_idx = 0, 'ALTER TABLE practice_records ADD INDEX idx_user_created_at (user_id, created_at)', 'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;