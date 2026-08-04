DROP TEMPORARY TABLE IF EXISTS q_rownum;
DROP TEMPORARY TABLE IF EXISTS c_rownum;

-- 1) 题目按 (subject_id, sub_subject_id) 分组编号
CREATE TEMPORARY TABLE q_rownum (
  id BIGINT PRIMARY KEY,
  subject_id INT NOT NULL,
  sub_subject_id INT NOT NULL,
  rn INT NOT NULL,
  k VARCHAR(64)
);
SET @k = '', @rn = 0;
INSERT INTO q_rownum (id, subject_id, sub_subject_id, rn, k)
SELECT id, subject_id, sub_subject_id,
       IF(@k = CONCAT(subject_id, ':', sub_subject_id), @rn := @rn + 1, @rn := 1),
       @k := CONCAT(subject_id, ':', sub_subject_id)
FROM questions
WHERE chapter_id = 0 AND status = 1
ORDER BY subject_id, sub_subject_id, id;

-- 2) 章节按 (subject_id, sub_subject_id) 分组编号，并统计每组数量
CREATE TEMPORARY TABLE c_rownum (
  id BIGINT PRIMARY KEY,
  subject_id INT NOT NULL,
  sub_subject_id INT NOT NULL,
  cidx INT NOT NULL,
  cnt INT NOT NULL,
  k VARCHAR(64)
);
SET @c = '', @ci = 0;
INSERT INTO c_rownum (id, subject_id, sub_subject_id, cidx, cnt, k)
SELECT id, subject_id, sub_subject_id,
       IF(@c = CONCAT(subject_id, ':', sub_subject_id), @ci := @ci + 1, @ci := 0),
       0,
       @c := CONCAT(subject_id, ':', sub_subject_id)
FROM chapters
ORDER BY subject_id, sub_subject_id, id;

-- 3) 按组键补充每组章节数量
UPDATE c_rownum cr
JOIN (SELECT subject_id, sub_subject_id, COUNT(*) AS cnt
      FROM chapters GROUP BY subject_id, sub_subject_id) s
  ON s.subject_id = cr.subject_id AND s.sub_subject_id = cr.sub_subject_id
SET cr.cnt = s.cnt;

-- 4) 分配章节：题目按组内序号轮询到章节
UPDATE questions q
JOIN q_rownum qr ON qr.id = q.id
JOIN chapters c ON c.subject_id = q.subject_id AND c.sub_subject_id = q.sub_subject_id
JOIN c_rownum cr ON cr.id = c.id
   AND cr.cidx = (qr.rn - 1) % cr.cnt
SET q.chapter_id = c.id;
