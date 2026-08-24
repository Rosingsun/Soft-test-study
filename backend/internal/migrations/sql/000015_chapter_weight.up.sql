-- =============================================================
-- 000015_chapter_weight.up.sql
-- 章节表新增 weight 字段（用于预估分计算）
-- =============================================================

ALTER TABLE `chapters`
  ADD COLUMN `weight` DECIMAL(5,2) NOT NULL DEFAULT 1.00
    COMMENT '章节权重（用于预估分计算，默认均匀 1.00）';
