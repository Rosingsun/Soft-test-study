-- =============================================================
-- 000016_chapter_material_id.up.sql
-- 章节表新增 material_id：关联学习资料（章节思维导图知识点）
-- 0 / NULL 表示该章节暂无关联思维导图
-- =============================================================

ALTER TABLE `chapters`
  ADD COLUMN `material_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '关联学习资料 study_materials.id，0=未关联'
    AFTER `weight`,
  ADD INDEX `idx_material_id` (`material_id`);
