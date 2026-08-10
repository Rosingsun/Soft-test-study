-- =============================================================
-- 000002_practice_records_indexes.down.sql
-- =============================================================

ALTER TABLE `practice_records` DROP INDEX `idx_user_mode_created`;
ALTER TABLE `practice_records` DROP INDEX `idx_user_created_at`;
