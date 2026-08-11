-- =============================================================
-- 000007_ai_generated_tasks.up.sql
-- 新增 ai_generated_tasks 表，持久化 AI 异步出题任务状态。
-- 修复 #20：原 service.aiTaskStore 内存 map 重启即丢，导致
-- 1) in-flight 任务结果丢失
-- 2) janitor 1h 后只能清不能超时回收
-- 3) 用户无法查询历史任务
-- 改为 DB 持久化后：
--   - service/ai_task.go 内存 store 降级为热缓存
--   - 所有 mutator 双写 DB
--   - 启动时 RecoverInflightTasks 把 status='running' 的行重新加进内存
--   - janitor 改为扫 DB + 按状态区分超时阈值（pending > 5min / running > 30min）
-- =============================================================

CREATE TABLE `ai_generated_tasks` (
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
