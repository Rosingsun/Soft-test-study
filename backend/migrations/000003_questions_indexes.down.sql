-- =============================================================
-- 000003_questions_indexes.down.sql
-- =============================================================

ALTER TABLE `questions` DROP INDEX `idx_year_status_type`;
ALTER TABLE `questions` DROP INDEX `idx_chapter_status`;
ALTER TABLE `questions` DROP INDEX `idx_subject_status_type`;
