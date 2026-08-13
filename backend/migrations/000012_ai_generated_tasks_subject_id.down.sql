-- =============================================================
-- 000012_ai_generated_tasks_subject_id.down.sql
-- 撤销 ai_generated_tasks.subject_id 列。
-- =============================================================

ALTER TABLE `ai_generated_tasks`
  DROP COLUMN `subject_id`;
