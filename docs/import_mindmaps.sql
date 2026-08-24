-- =====================================================================
-- 学习资料导入：系统分析师教材思维导图（第 1~13 章，缺第 2 章）
-- 前置条件：12 份 xmind 已放置于 backend/data/uploads/materials/{id}/note.xmind
-- 关联科目：系统分析师（通过子查询按名称匹配，避免 id 硬编码）
-- 说明：第 6 章存在两个版本的 xmind，此处采用新版（含 EAI 等 8 小节）
-- =====================================================================

INSERT INTO `study_materials`
  (`id`, `title`, `description`, `subject_id`, `cover_url`, `file_url`, `file_name`, `file_size`, `sort_order`, `status`)
VALUES
  (1,  '第1章 绪论',                   '信息与信息系统核心考点思维导图',                               (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/1/note.xmind',  '第1章 绪论.xmind',                   542936,   101, 1),
  (2,  '第3章 计算机系统',             '计算机系统组成、存储器、输入输出、指令系统与操作系统思维导图', (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/2/note.xmind',  '第3章 计算机系统.xmind',             3413266,  100, 1),
  (3,  '第4章 计算机网络与分布式系统', '网络基础、体系结构协议、局域网广域网与云计算思维导图',         (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/3/note.xmind',  '第4章 计算机网络与分布式系统.xmind', 798443,   99,  1),
  (4,  '第5章 数据库系统',             'DBMS、关系数据库、数据仓库与非关系数据库思维导图',             (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/4/note.xmind',  '第5章 数据库系统.xmind',             1477516,  98,  1),
  (5,  '第6章 企业信息化',             '信息化规划、电子商务电子政务、业务流程重组与 EAI 思维导图',     (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/5/note.xmind',  '第6章 企业信息化.xmind',             3497404,  97,  1),
  (6,  '第7章 软件工程',               '生命周期、开发模型、过程管理与软件重用思维导图',               (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/6/note.xmind',  '第7章 软件工程.xmind',               1315086,  96,  1),
  (7,  '第8章 项目管理',               '范围、进度、成本、配置、质量与风险管理思维导图',               (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/7/note.xmind',  '第8章 项目管理.xmind',               564035,   95,  1),
  (8,  '第9章 信息安全',               '安全体系、数据保密、访问控制、容灾与可靠性思维导图',           (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/8/note.xmind',  '第9章 信息安全.xmind',               1076392,  94,  1),
  (9,  '第10章 系统规划与分析',        '项目提出与选择、问题分析、业务流程与可行性分析思维导图',       (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/9/note.xmind',  '第10章 系统规划与分析.xmind',        325358,   93,  1),
  (10, '第11章 软件需求工程',          '需求获取、结构化与面向对象分析、需求验证与管理思维导图',       (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/10/note.xmind', '第11章 软件需求工程.xmind',          112490,   92,  1),
  (11, '第12章 软件架构设计',          '架构建模、架构风格、质量属性与实现思维导图',                   (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/11/note.xmind', '第12章 软件架构设计.xmind',          2229809,  91,  1),
  (12, '第13章 系统设计',              '处理流程设计与结构化设计思维导图',                             (SELECT s.id FROM (SELECT id FROM subjects WHERE name = '系统分析师') s), NULL, '/uploads/materials/12/note.xmind', '第13章 系统设计.xmind',              693531,   90,  1)
ON DUPLICATE KEY UPDATE
  `title` = VALUES(`title`),
  `description` = VALUES(`description`),
  `file_url` = VALUES(`file_url`),
  `file_name` = VALUES(`file_name`),
  `updated_at` = CURRENT_TIMESTAMP;

-- =====================================================================
-- 章节关联思维导图：综合知识考纲章节 → 教材章节导图（material_id）
-- 前置条件：已执行迁移 000016_chapter_material_id.up.sql
-- 仅映射名称对应的章节；其余考纲章（法律法规/数学/运筹学/数据结构等）暂无导图
-- =====================================================================

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 1   -- 第1章 绪论
WHERE ss.`name` = '综合知识' AND c.`name` = '绪论'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 2   -- 第3章 计算机系统
WHERE ss.`name` = '综合知识' AND c.`name` = '计算机组成与体系结构'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 3   -- 第4章 计算机网络与分布式系统（网络 + 分布式两章共用）
WHERE ss.`name` = '综合知识' AND c.`name` IN ('计算机网络', '分布式系统与中间件')
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 4   -- 第5章 数据库系统
WHERE ss.`name` = '综合知识' AND c.`name` = '数据库系统'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 5   -- 第6章 企业信息化
WHERE ss.`name` = '综合知识' AND c.`name` = '企业信息化'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 6   -- 第7章 软件工程
WHERE ss.`name` = '综合知识' AND c.`name` = '软件工程'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 7   -- 第8章 项目管理
WHERE ss.`name` = '综合知识' AND c.`name` = '项目管理'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');

UPDATE `chapters` c
JOIN `sub_subjects` ss ON ss.`id` = c.`sub_subject_id`
SET c.`material_id` = 8   -- 第9章 信息安全
WHERE ss.`name` = '综合知识' AND c.`name` = '信息安全'
  AND ss.`subject_id` = (SELECT id FROM `subjects` WHERE name = '系统分析师');
