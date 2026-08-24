-- =============================================================
-- 000016_chapter_material_id.down.sql
-- 回滚：移除 chapters.material_id
-- =============================================================

ALTER TABLE `chapters`
  DROP INDEX `idx_material_id`,
  DROP COLUMN `material_id`;
