-- OPT-07: 回滚 uk_exam_pending 唯一索引
-- 注意：必须先 DROP INDEX 再 DROP COLUMN，否则会因为索引依赖列而失败
DROP INDEX uk_exam_pending ON exam_records;
ALTER TABLE exam_records DROP COLUMN pending_dedup_key;
