-- =============================================================
-- 000006_ai_generated_batch_id.up.sql
-- 为 ai_generated_questions 增加 batch_id 字段，标识「一次生成的一批题目」，
-- 用于 AI 生成题历史记录按批次分组。
--
-- 历史数据回填：batch_id 为空的历史行，按 用户 + 科目 + 创建时间(到分钟)
-- 分为同一批次，回填 legacy-{user_id}-{subject_id}-{ts}。
-- =============================================================

ALTER TABLE `ai_generated_questions`
  ADD COLUMN `batch_id` varchar(64) NOT NULL DEFAULT '' AFTER `question_id`,
  ADD INDEX `idx_ai_generated_batch_id` (`batch_id`);

UPDATE `ai_generated_questions` AS a
  JOIN (
    SELECT `user_id`, `subject_id`, DATE_FORMAT(`created_at`, '%Y%m%d%H%i') AS `ts`
      FROM `ai_generated_questions`
     WHERE `batch_id` = ''
     GROUP BY `user_id`, `subject_id`, `ts`
  ) AS b
    ON a.`user_id` = b.`user_id`
   AND a.`subject_id` = b.`subject_id`
   AND DATE_FORMAT(a.`created_at`, '%Y%m%d%H%i') = b.`ts`
   SET a.`batch_id` = CONCAT('legacy-', b.`user_id`, '-', b.`subject_id`, '-', b.`ts`)
 WHERE a.`batch_id` = '';
