-- =============================================================
-- 000012_ai_generated_tasks_subject_id.up.sql
-- ai_generated_tasks 增加 subject_id 列，让进行中任务在历史列表里能补全 subject_name。
-- 修复：ListGenerateHistory 把 pending/running 任务 union 进历史后，
-- 需要 subject_id 查 subjectNameMap。ChapterID=0（不限定章节）的任务
-- 也有合法 subject_id，必须显式存。
-- =============================================================

ALTER TABLE `ai_generated_tasks`
  ADD COLUMN `subject_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '提交时所属科目；用于历史列表补全 subject_name'
    AFTER `user_id`;
