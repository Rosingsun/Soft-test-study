-- 案例分析大题+小题：questions 表增加 parent_id 字段
-- parent_id=0 表示独立题目或案例分析大题目
-- parent_id>0 表示小题目，值为所属大题目的 ID

ALTER TABLE `questions`
  ADD COLUMN `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '父题目ID: 0=独立题目/大题目, >0=小题目'
    AFTER `chapter_id`,
  ADD INDEX `idx_parent_id` (`parent_id`);

-- 数据迁移：将共享相同 case_material 的多条 case_study 记录合并为 1 个父题 + N 个子题
-- 1. 每组保留 id 最小的一条作为父题（type='case_study'，case_material 保留）
-- 2. 其余记录转为子题（type 改为 'short'，parent_id 指向父题，清空 case_material）
CREATE TEMPORARY TABLE tmp_case_parent_mapping (
  old_id BIGINT UNSIGNED PRIMARY KEY,
  new_parent_id BIGINT UNSIGNED NOT NULL
);

-- 映射表写入组内【全部成员】id -> 组内最小 id（父题），
-- 仅写 MIN(id) 会使 UPDATE 被 q.id != m.new_parent_id 排除而变成空操作
INSERT IGNORE INTO tmp_case_parent_mapping (old_id, new_parent_id)
SELECT q1.id, grp.min_id
FROM questions q1
INNER JOIN (
  SELECT case_material, MIN(id) AS min_id
  FROM questions
  WHERE type = 'case_study'
    AND case_material IS NOT NULL
    AND case_material != ''
  GROUP BY case_material
  HAVING COUNT(*) > 1
) grp ON q1.case_material = grp.case_material
WHERE q1.type = 'case_study'
  AND q1.case_material IS NOT NULL
  AND q1.case_material != '';

UPDATE questions q
INNER JOIN tmp_case_parent_mapping m ON q.id = m.old_id
SET q.parent_id = m.new_parent_id,
    q.type = 'short',
    q.case_material = NULL
WHERE q.id != m.new_parent_id;

DROP TEMPORARY TABLE tmp_case_parent_mapping;
