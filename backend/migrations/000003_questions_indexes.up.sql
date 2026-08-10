-- =============================================================
-- 000003_questions_indexes.up.sql
-- 为 questions 加复合索引，支撑：
--   - repository.question.FindRandom/FindRandomFiltered/FindSpecial/FindRandomReal/FindRandomObjective
--   - repository.question.CountBySubjectAndType
--   - repository.question.FindByChapterID/CountByChapterID
--   - repository.question.FindEssayQuestions（year + type + status）
-- =============================================================

ALTER TABLE `questions`
  ADD INDEX `idx_subject_status_type` (`subject_id`, `status`, `type`);

ALTER TABLE `questions`
  ADD INDEX `idx_chapter_status` (`chapter_id`, `status`);

ALTER TABLE `questions`
  ADD INDEX `idx_year_status_type` (`year`, `status`, `type`);
