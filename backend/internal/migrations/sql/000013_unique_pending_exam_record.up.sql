-- OPT-07: 防止同用户对同模板并发 StartExam 产生多条 pending 记录
-- 1. 先清理历史脏数据：保留每组 (user_id, template_id) 中 status='pending' 的最新一条
DELETE r1 FROM exam_records r1
INNER JOIN exam_records r2
  ON r1.user_id = r2.user_id
 AND r1.template_id = r2.template_id
 AND r1.status = 'pending' AND r2.status = 'pending'
 AND r1.id < r2.id;

-- 2. 新增部分唯一索引：同用户同模板下 status='pending' 最多一条
--    仅约束 pending，已完成（finished）/ 已交卷的记录不受影响
CREATE UNIQUE INDEX uk_exam_pending
  ON exam_records (user_id, template_id, status)
  WHERE status = 'pending';
