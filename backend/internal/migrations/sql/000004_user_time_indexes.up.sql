-- =============================================================
-- 000004_user_time_indexes.up.sql
-- 为业务表加 user_id+时间/状态 复合索引，消除全表扫描：
--   - wrong_questions(user_id, last_wrong_at)  错题本 ORDER BY 排序
--   - check_ins(user_id, status, check_date)    打卡统计/历史
--   - exam_records(user_id, status, created_at) 考试记录列表
--   - exam_record_answers(record_id, question_id) 唯一约束 + 答卷详情拼装
-- =============================================================

ALTER TABLE `wrong_questions`
  ADD INDEX `idx_user_lastwrong` (`user_id`, `last_wrong_at`);

ALTER TABLE `check_ins`
  ADD INDEX `idx_user_status_date` (`user_id`, `status`, `check_date`);

ALTER TABLE `exam_records`
  ADD INDEX `idx_user_status_created` (`user_id`, `status`, `created_at`);

ALTER TABLE `exam_record_answers`
  ADD UNIQUE INDEX `uk_record_question` (`record_id`, `question_id`);
