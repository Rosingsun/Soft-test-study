-- =============================================================
-- 000001_init_schema.down.sql
-- 回滚 000001，按依赖反序删除 24 张表
-- =============================================================

SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `notifications`;
DROP TABLE IF EXISTS `review_cards`;
DROP TABLE IF EXISTS `study_plans`;
DROP TABLE IF EXISTS `check_in_questions`;
DROP TABLE IF EXISTS `check_ins`;
DROP TABLE IF EXISTS `study_materials`;
DROP TABLE IF EXISTS `essay_scores`;
DROP TABLE IF EXISTS `ai_generated_questions`;
DROP TABLE IF EXISTS `exam_record_answers`;
DROP TABLE IF EXISTS `exam_records`;
DROP TABLE IF EXISTS `exam_template_questions`;
DROP TABLE IF EXISTS `exam_templates`;
DROP TABLE IF EXISTS `wrong_questions`;
DROP TABLE IF EXISTS `question_marks`;
DROP TABLE IF EXISTS `question_favorites`;
DROP TABLE IF EXISTS `favorite_folders`;
DROP TABLE IF EXISTS `practice_records`;
DROP TABLE IF EXISTS `question_knowledge`;
DROP TABLE IF EXISTS `questions`;
DROP TABLE IF EXISTS `chapters`;
DROP TABLE IF EXISTS `sub_subjects`;
DROP TABLE IF EXISTS `subjects`;
DROP TABLE IF EXISTS `exam_levels`;
DROP TABLE IF EXISTS `users`;

SET FOREIGN_KEY_CHECKS = 1;
