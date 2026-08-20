-- =============================================================
-- 000006_ai_generated_batch_id.down.sql
-- 撤销 batch_id 字段及其索引。
-- 注意：回滚会丢失批次分组信息（历史回填的 legacy-* 值）。
-- =============================================================

ALTER TABLE `ai_generated_questions`
  DROP INDEX `idx_ai_generated_batch_id`,
  DROP COLUMN `batch_id`;
