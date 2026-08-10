-- =============================================================
-- 000004_user_time_indexes.down.sql
-- =============================================================

ALTER TABLE `exam_record_answers` DROP INDEX `uk_record_question`;
ALTER TABLE `exam_records` DROP INDEX `idx_user_status_created`;
ALTER TABLE `check_ins` DROP INDEX `idx_user_status_date`;
ALTER TABLE `wrong_questions` DROP INDEX `idx_user_lastwrong`;
