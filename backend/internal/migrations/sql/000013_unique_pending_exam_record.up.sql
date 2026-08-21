-- OPT-07: 防止同用户对同模板并发 StartExam 产生多条 pending 记录
-- MySQL 不支持 partial unique index (CREATE INDEX ... WHERE)，改用"生成列 + 唯一索引"实现：
--   pending_dedup_key 在 status='pending' 时存 user_id:template_id，其他状态为 NULL；
--   MySQL 唯一索引允许多个 NULL、但不允许非 NULL 重复 → 等价于 partial unique index。

-- 1. 先清理历史脏数据：保留每组 (user_id, template_id) 中 status='pending' 的最新一条
DELETE r1 FROM exam_records r1
INNER JOIN exam_records r2
  ON r1.user_id = r2.user_id
 AND r1.template_id = r2.template_id
 AND r1.status = 'pending' AND r2.status = 'pending'
 AND r1.id < r2.id;

-- 2. 添加虚拟生成列：pending 时存 user_id:template_id，否则 NULL
ALTER TABLE exam_records
  ADD COLUMN pending_dedup_key VARCHAR(64)
    GENERATED ALWAYS AS (
      CASE WHEN status = 'pending'
           THEN CONCAT(user_id, ':', template_id)
           ELSE NULL
      END
    ) VIRTUAL;

-- 3. 在生成列上建唯一索引
CREATE UNIQUE INDEX uk_exam_pending
  ON exam_records (pending_dedup_key);
