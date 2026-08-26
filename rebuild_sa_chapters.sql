-- ====================================================================
-- 系统分析师-综合知识 章节重建脚本
-- 适用：subject_id=9, sub_subject_id=4
-- 旧章节：18 条 (ID 200-217)，与官方 22 章不一致
-- 新章节：22 条 (ID 201-222)，与 sa_chapter_extra.sql 硬编码一致
-- ====================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
START TRANSACTION;

-- ============================================================
-- Step 1: 把 subject_id=9, sub_subject_id=4 下所有题目的 chapter_id 重置为 0
-- 涉及：questions, ai_generated_questions, ai_generated_tasks
-- 保留题目本身，不级联删除
-- ============================================================
UPDATE `questions`
SET `chapter_id` = 0
WHERE `subject_id` = 9 AND `sub_subject_id` = 4;

UPDATE `ai_generated_questions`
SET `chapter_id` = 0
WHERE `subject_id` = 9 AND `chapter_id` > 0;

UPDATE `ai_generated_tasks`
SET `chapter_id` = 0
WHERE `subject_id` = 9 AND `chapter_id` > 0;

-- ============================================================
-- Step 2: 删除旧的（错的）章节
-- 题目 chapter_id 已重置为 0，可安全删除
-- ============================================================
DELETE FROM `chapters`
WHERE `subject_id` = 9 AND `sub_subject_id` = 4;

-- ============================================================
-- Step 3: 用固定 ID 201-222 插入官方 22 章
-- 顺序按《系统分析师教程（第 2 版 · 2024）》
-- ID 与 sa_chapter_extra.sql 硬编码完全吻合
-- ============================================================
INSERT INTO `chapters` (`id`, `subject_id`, `sub_subject_id`, `parent_id`, `name`, `sort_order`, `weight`, `material_id`, `created_at`) VALUES
(201, 9, 4, 0, '绪论', 1, 1.00, 0, NOW()),
(202, 9, 4, 0, '数学与工程基础', 2, 1.00, 0, NOW()),
(203, 9, 4, 0, '计算机系统', 3, 1.00, 0, NOW()),
(204, 9, 4, 0, '计算机网络与分布式系统', 4, 1.00, 0, NOW()),
(205, 9, 4, 0, '数据库系统', 5, 1.00, 0, NOW()),
(206, 9, 4, 0, '企业信息化', 6, 1.00, 0, NOW()),
(207, 9, 4, 0, '软件工程', 7, 1.00, 0, NOW()),
(208, 9, 4, 0, '项目管理', 8, 1.00, 0, NOW()),
(209, 9, 4, 0, '信息安全', 9, 1.00, 0, NOW()),
(210, 9, 4, 0, '系统规划与分析', 10, 1.00, 0, NOW()),
(211, 9, 4, 0, '软件需求工程', 11, 1.00, 0, NOW()),
(212, 9, 4, 0, '软件架构设计', 12, 1.00, 0, NOW()),
(213, 9, 4, 0, '系统设计', 13, 1.00, 0, NOW()),
(214, 9, 4, 0, '软件实现与测试', 14, 1.00, 0, NOW()),
(215, 9, 4, 0, '系统运行与维护', 15, 1.00, 0, NOW()),
(216, 9, 4, 0, 'Web 应用系统', 16, 1.00, 0, NOW()),
(217, 9, 4, 0, '嵌入式系统', 17, 1.00, 0, NOW()),
(218, 9, 4, 0, '移动应用系统', 18, 1.00, 0, NOW()),
(219, 9, 4, 0, '大数据处理系统', 19, 1.00, 0, NOW()),
(220, 9, 4, 0, '微服务系统', 20, 1.00, 0, NOW()),
(221, 9, 4, 0, '信息物理系统', 21, 1.00, 0, NOW()),
(222, 9, 4, 0, '系统分析师论文写作要点', 22, 1.00, 0, NOW());

-- ============================================================
-- Step 4: 按关键词智能重归所有题目的 chapter_id
-- 优先级：与 import_real/main.go:226-252 chapterKeywords 顺序一致
-- 高优先级章节（嵌入式/移动/大数据/微服务/信息物理）先匹配
-- 低优先级章节（数学/计算机系统/网络/数据库等）后匹配
-- 兜底：未匹配上的题目归到「计算机系统」作为最通用章节
-- ============================================================

-- 4.1 嵌入式系统 (ID=217)
UPDATE `questions` q
SET q.chapter_id = 217
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%嵌入式%' OR q.content LIKE '%实时操作系统%' OR q.content LIKE '%RTOS%'
    OR q.content LIKE '%片上系统%' OR q.content LIKE '%SoC%' OR q.content LIKE '%交叉编译%'
    OR q.content LIKE '%固件%' OR q.content LIKE '%微控制器%' OR q.content LIKE '%单片机%'
    OR q.content LIKE '%JTAG%' OR q.content LIKE '%看门狗%');

-- 4.2 信息物理系统 (ID=221)
UPDATE `questions` q
SET q.chapter_id = 221
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%信息物理系统%' OR q.content LIKE '%CPS%'
    OR q.content LIKE '%工业互联网%' OR q.content LIKE '%数字孪生%' OR q.content LIKE '%智能工厂%'
    OR q.content LIKE '%工业 4.0%' OR q.content LIKE '%工业4.0%' OR q.content LIKE '%OPC UA%'
    OR q.content LIKE '%MQTT%' OR q.content LIKE '%时间敏感网络%' OR q.content LIKE '%TSN%');

-- 4.3 移动应用系统 (ID=218)
UPDATE `questions` q
SET q.chapter_id = 218
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%移动应用%' OR q.content LIKE '%Android%' OR q.content LIKE '%iOS%'
    OR q.content LIKE '%Flutter%' OR q.content LIKE '%React Native%'
    OR q.content LIKE '%Activity%' OR q.content LIKE '%Service%'
    OR q.content LIKE '%BroadcastReceiver%' OR q.content LIKE '%Cocoa Touch%'
    OR q.content LIKE '%移动端%' OR q.content LIKE '%APP 架构%' OR q.content LIKE '%APP开发%');

-- 4.4 大数据处理系统 (ID=219)
UPDATE `questions` q
SET q.chapter_id = 219
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%大数据%' OR q.content LIKE '%Hadoop%' OR q.content LIKE '%HDFS%'
    OR q.content LIKE '%MapReduce%' OR q.content LIKE '%Spark%' OR q.content LIKE '%Flink%'
    OR q.content LIKE '%Storm%' OR q.content LIKE '%Kafka Streams%'
    OR q.content LIKE '%数据湖%' OR q.content LIKE '%数据中台%' OR q.content LIKE '%湖仓一体%'
    OR q.content LIKE '%Hive%' OR q.content LIKE '%HBase%' OR q.content LIKE '%NameNode%'
    OR q.content LIKE '%DataNode%');

-- 4.5 微服务系统 (ID=220)
UPDATE `questions` q
SET q.chapter_id = 220
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%微服务%' OR q.content LIKE '%服务拆分%' OR q.content LIKE '%服务治理%'
    OR q.content LIKE '%服务网关%' OR q.content LIKE '%API 网关%' OR q.content LIKE '%服务注册%'
    OR q.content LIKE '%服务发现%' OR q.content LIKE '%Eureka%' OR q.content LIKE '%Consul%'
    OR q.content LIKE '%Nacos%' OR q.content LIKE '%限流%' OR q.content LIKE '%熔断%'
    OR q.content LIKE '%Hystrix%' OR q.content LIKE '%Sentinel%'
    OR q.content LIKE '%Spring Cloud%' OR q.content LIKE '%Service Mesh%' OR q.content LIKE '%Istio%'
    OR q.content LIKE '%分布式事务%' OR q.content LIKE '%TCC%' OR q.content LIKE '%Saga%'
    OR q.content LIKE '%Seata%');

-- 4.6 Web 应用系统 (ID=216)
UPDATE `questions` q
SET q.chapter_id = 216
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%Web 应用%' OR q.content LIKE '%前后端分离%' OR q.content LIKE '%RESTful%'
    OR q.content LIKE '%GraphQL%' OR q.content LIKE '%PWA%' OR q.content LIKE '%Service Worker%');

-- 4.7 信息安全 (ID=209)
UPDATE `questions` q
SET q.chapter_id = 209
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%信息安全%' OR q.content LIKE '%加密%' OR q.content LIKE '%对称密钥%'
    OR q.content LIKE '%非对称密钥%' OR q.content LIKE '%RSA%' OR q.content LIKE '%DES%'
    OR q.content LIKE '%AES%' OR q.content LIKE '%ECC%' OR q.content LIKE '%MD5%'
    OR q.content LIKE '%SHA%' OR q.content LIKE '%数字签名%' OR q.content LIKE '%数字证书%'
    OR q.content LIKE '%PKI%' OR q.content LIKE '%CA%' OR q.content LIKE '%防火墙%'
    OR q.content LIKE '%入侵检测%' OR q.content LIKE '%IDS%' OR q.content LIKE '%IPS%'
    OR q.content LIKE '%SQL 注入%' OR q.content LIKE '%SQL注入%' OR q.content LIKE '%XSS%'
    OR q.content LIKE '%访问控制%' OR q.content LIKE '%权限管理%' OR q.content LIKE '%等级保护%'
    OR q.content LIKE '%灾备%' OR q.content LIKE '%SSL%' OR q.content LIKE '%TLS%'
    OR q.content LIKE '%IPSec%' OR q.content LIKE '%VPN%' OR q.content LIKE '%Kerberos%'
    OR q.content LIKE '%公钥%' OR q.content LIKE '%私钥%' OR q.content LIKE '%报文摘要%'
    OR q.content LIKE '%认证技术%');

-- 4.8 数学与工程基础 (ID=202)
UPDATE `questions` q
SET q.chapter_id = 202
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%线性规划%' OR q.content LIKE '%单纯形%' OR q.content LIKE '%整数规划%'
    OR q.content LIKE '%0-1 规划%' OR q.content LIKE '%运输问题%' OR q.content LIKE '%指派问题%'
    OR q.content LIKE '%最大流%' OR q.content LIKE '%最小费用流%' OR q.content LIKE '%排队论%'
    OR q.content LIKE '%M/M/1%' OR q.content LIKE '%博弈论%' OR q.content LIKE '%纳什均衡%'
    OR q.content LIKE '%决策树%' OR q.content LIKE '%盈亏平衡%' OR q.content LIKE '%期望收益%'
    OR q.content LIKE '%最小生成树%' OR q.content LIKE '%Prim%' OR q.content LIKE '%Kruskal%'
    OR q.content LIKE '%Dijkstra%' OR q.content LIKE '%Floyd%' OR q.content LIKE '%离散数学%'
    OR q.content LIKE '%命题逻辑%' OR q.content LIKE '%谓词逻辑%' OR q.content LIKE '%等价关系%'
    OR q.content LIKE '%偏序%' OR q.content LIKE '%概率论%' OR q.content LIKE '%数学期望%'
    OR q.content LIKE '%方差%' OR q.content LIKE '%正态分布%' OR q.content LIKE '%二项分布%'
    OR q.content LIKE '%排列组合%' OR q.content LIKE '%微积分%' OR q.content LIKE '%线性代数%'
    OR q.content LIKE '%矩阵运算%' OR q.content LIKE '%特征值%' OR q.content LIKE '%参数估计%'
    OR q.content LIKE '%假设检验%');

-- 4.9 数据库系统 (ID=205)
UPDATE `questions` q
SET q.chapter_id = 205
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%数据库%' OR q.content LIKE '%关系代数%' OR q.content LIKE '%元组%'
    OR q.content LIKE '%候选键%' OR q.content LIKE '%主键%' OR q.content LIKE '%外键%'
    OR q.content LIKE '%函数依赖%' OR q.content LIKE '%范式%' OR q.content LIKE '%1NF%'
    OR q.content LIKE '%2NF%' OR q.content LIKE '%3NF%' OR q.content LIKE '%BCNF%'
    OR q.content LIKE '%模式分解%' OR q.content LIKE '%事务的 ACID%' OR q.content LIKE '%ACID%'
    OR q.content LIKE '%隔离级别%' OR q.content LIKE '%封锁协议%' OR q.content LIKE '%两段锁%'
    OR q.content LIKE '%共享锁%' OR q.content LIKE '%排他锁%' OR q.content LIKE '%视图%'
    OR q.content LIKE '%触发器%' OR q.content LIKE '%存储过程%' OR q.content LIKE '%游标%'
    OR q.content LIKE '%数据仓库%' OR q.content LIKE '%星型模型%' OR q.content LIKE '%雪花模型%'
    OR q.content LIKE '%OLAP%' OR q.content LIKE '%OLTP%' OR q.content LIKE '%数据挖掘%'
    OR q.content LIKE '%NoSQL%' OR q.content LIKE '%反规范化%' OR q.content LIKE '%分布式数据库%'
    OR q.content LIKE '%两阶段提交%' OR q.content LIKE '%并发控制%' OR q.content LIKE '%E-R 图%'
    OR q.content LIKE '%ER 图%' OR q.content LIKE '%实体联系%');

-- 4.10 计算机系统 (ID=203)
UPDATE `questions` q
SET q.chapter_id = 203
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%哈夫曼%' OR q.content LIKE '%邻接矩阵%' OR q.content LIKE '%邻接表%'
    OR q.content LIKE '%拓扑排序%' OR q.content LIKE '%关键路径%' OR q.content LIKE '%AOE%'
    OR q.content LIKE '%时间复杂度%' OR q.content LIKE '%空间复杂度%' OR q.content LIKE '%二叉树%'
    OR q.content LIKE '%平衡树%' OR q.content LIKE '%红黑树%' OR q.content LIKE '%线索二叉树%'
    OR q.content LIKE '%顺序栈%' OR q.content LIKE '%链栈%' OR q.content LIKE '%循环队列%'
    OR q.content LIKE '%链表%' OR q.content LIKE '%顺序表%' OR q.content LIKE '%散列表%'
    OR q.content LIKE '%KMP%' OR q.content LIKE '%贪心%' OR q.content LIKE '%动态规划%'
    OR q.content LIKE '%分治法%' OR q.content LIKE '%回溯法%' OR q.content LIKE '%分支限界%'
    OR q.content LIKE '%快速排序%' OR q.content LIKE '%归并排序%' OR q.content LIKE '%堆排序%'
    OR q.content LIKE '%基数排序%' OR q.content LIKE '%希尔排序%' OR q.content LIKE '%折半查找%'
    OR q.content LIKE '%B 树%' OR q.content LIKE '%B+ 树%' OR q.content LIKE '%B+树%'
    OR q.content LIKE '%CPU%' OR q.content LIKE '%运算器%' OR q.content LIKE '%控制器%'
    OR q.content LIKE '%寄存器%' OR q.content LIKE '%指令周期%' OR q.content LIKE '%指令流水线%'
    OR q.content LIKE '%流水线%' OR q.content LIKE '%Cache%' OR q.content LIKE '%高速缓存%'
    OR q.content LIKE '%主存%' OR q.content LIKE '%辅存%' OR q.content LIKE '%寻址%'
    OR q.content LIKE '%总线仲裁%' OR q.content LIKE '%冯·诺依曼%' OR q.content LIKE '%哈佛结构%'
    OR q.content LIKE '%RISC%' OR q.content LIKE '%CISC%' OR q.content LIKE '%Flynn%'
    OR q.content LIKE '%SISD%' OR q.content LIKE '%SIMD%' OR q.content LIKE '%MIMD%'
    OR q.content LIKE '%多核%' OR q.content LIKE '%海明码%' OR q.content LIKE '%CRC%'
    OR q.content LIKE '%原码%' OR q.content LIKE '%反码%' OR q.content LIKE '%补码%'
    OR q.content LIKE '%移码%' OR q.content LIKE '%浮点数%' OR q.content LIKE '%规格化%'
    OR q.content LIKE '%DMA%' OR q.content LIKE '%中断%' OR q.content LIKE '%通道%'
    OR q.content LIKE '%MTBF%' OR q.content LIKE '%RAID%' OR q.content LIKE '%磁盘阵列%'
    OR q.content LIKE '%字长%' OR q.content LIKE '%操作系统%' OR q.content LIKE '%进程%'
    OR q.content LIKE '%线程%' OR q.content LIKE '%协程%' OR q.content LIKE '%死锁%'
    OR q.content LIKE '%银行家算法%' OR q.content LIKE '%信号量%' OR q.content LIKE '%PV 操作%'
    OR q.content LIKE '%PV操作%' OR q.content LIKE '%管程%' OR q.content LIKE '%临界资源%'
    OR q.content LIKE '%临界区%' OR q.content LIKE '%页面置换%' OR q.content LIKE '%LRU%'
    OR q.content LIKE '%FIFO 置换%' OR q.content LIKE '%缺页中断%' OR q.content LIKE '%虚拟存储%'
    OR q.content LIKE '%分页%' OR q.content LIKE '%分段%' OR q.content LIKE '%段页式%'
    OR q.content LIKE '%SCAN%' OR q.content LIKE '%SSTF%' OR q.content LIKE '%索引节点%'
    OR q.content LIKE '%位示图%' OR q.content LIKE '%Spooling%' OR q.content LIKE '%作业调度%'
    OR q.content LIKE '%时间片轮转%' OR q.content LIKE '%多级反馈队列%' OR q.content LIKE '%前趋图%'
    OR q.content LIKE '%前驱图%' OR q.content LIKE '%文法%' OR q.content LIKE '%正规式%'
    OR q.content LIKE '%正则表达式%' OR q.content LIKE '%词法分析%' OR q.content LIKE '%语法分析%'
    OR q.content LIKE '%语义分析%' OR q.content LIKE '%中间代码%' OR q.content LIKE '%目标代码%'
    OR q.content LIKE '%编译器%' OR q.content LIKE '%解释器%' OR q.content LIKE '%传值调用%'
    OR q.content LIKE '%引用调用%' OR q.content LIKE '%递归%' OR q.content LIKE '%闭包%'
    OR q.content LIKE '%作用域%' OR q.content LIKE '%动态绑定%' OR q.content LIKE '%静态绑定%'
    OR q.content LIKE '%多态%' OR q.content LIKE '%逆波兰%' OR q.content LIKE '%后缀表达式%'
    OR q.content LIKE '%多媒体%' OR q.content LIKE '%图像存储%' OR q.content LIKE '%采样频率%'
    OR q.content LIKE '%量化位数%' OR q.content LIKE '%色深%' OR q.content LIKE '%像素%'
    OR q.content LIKE '%MPEG%' OR q.content LIKE '%JPEG%' OR q.content LIKE '%H.26%'
    OR q.content LIKE '%帧率%' OR q.content LIKE '%有损压缩%' OR q.content LIKE '%无损压缩%'
    OR q.content LIKE '%排序%' OR q.content LIKE '%查找%' OR q.content LIKE '%算法%'
    OR q.content LIKE '%存储%' OR q.content LIKE '%冯%');

-- 4.11 计算机网络与分布式系统 (ID=204)
UPDATE `questions` q
SET q.chapter_id = 204
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%分布式系统%' OR q.content LIKE '%CAP 定理%' OR q.content LIKE '%BASE%'
    OR q.content LIKE '%最终一致性%' OR q.content LIKE '%Paxos%' OR q.content LIKE '%Raft%'
    OR q.content LIKE '%ZooKeeper%' OR q.content LIKE '%一致性哈希%' OR q.content LIKE '%数据分片%'
    OR q.content LIKE '%消息队列%' OR q.content LIKE '%发布订阅%' OR q.content LIKE '%Kafka%'
    OR q.content LIKE '%RabbitMQ%' OR q.content LIKE '%负载均衡%' OR q.content LIKE '%CDN%'
    OR q.content LIKE '%缓存穿透%' OR q.content LIKE '%Redis%' OR q.content LIKE '%RPC%'
    OR q.content LIKE '%SOA%' OR q.content LIKE '%ESB%' OR q.content LIKE '%中间件%'
    OR q.content LIKE '%Kubernetes%' OR q.content LIKE '%计算机网络%' OR q.content LIKE '%OSI%'
    OR q.content LIKE '%TCP/IP%' OR q.content LIKE '%三次握手%' OR q.content LIKE '%四次挥手%'
    OR q.content LIKE '%滑动窗口%' OR q.content LIKE '%拥塞控制%' OR q.content LIKE '%IP 地址%'
    OR q.content LIKE '%IP地址%' OR q.content LIKE '%子网划分%' OR q.content LIKE '%子网掩码%'
    OR q.content LIKE '%CIDR%' OR q.content LIKE '%IPv6%' OR q.content LIKE '%ARP%'
    OR q.content LIKE '%ICMP%' OR q.content LIKE '%DHCP%' OR q.content LIKE '%DNS%'
    OR q.content LIKE '%HTTP%' OR q.content LIKE '%HTTPS%' OR q.content LIKE '%FTP%'
    OR q.content LIKE '%SMTP%' OR q.content LIKE '%POP3%' OR q.content LIKE '%SNMP%'
    OR q.content LIKE '%路由协议%' OR q.content LIKE '%RIP%' OR q.content LIKE '%OSPF%'
    OR q.content LIKE '%BGP%' OR q.content LIKE '%交换机%' OR q.content LIKE '%VLAN%'
    OR q.content LIKE '%以太网%' OR q.content LIKE '%CSMA/CD%' OR q.content LIKE '%令牌环%'
    OR q.content LIKE '%网桥%' OR q.content LIKE '%集线器%' OR q.content LIKE '%拓扑%'
    OR q.content LIKE '%星型%' OR q.content LIKE '%环型%' OR q.content LIKE '%总线型%'
    OR q.content LIKE '%曼彻斯特%' OR q.content LIKE '%奈奎斯特%' OR q.content LIKE '%香农%'
    OR q.content LIKE '%波特率%' OR q.content LIKE '%比特率%' OR q.content LIKE '%多路复用%'
    OR q.content LIKE '%时分复用%' OR q.content LIKE '%频分复用%' OR q.content LIKE '%ATM%'
    OR q.content LIKE '%路由器%' OR q.content LIKE '%IP 报文%' OR q.content LIKE '%QoS%'
    OR q.content LIKE '%IntServ%' OR q.content LIKE '%DiffServ%' OR q.content LIKE '%RSVP%'
    OR q.content LIKE '%Wi-Fi%' OR q.content LIKE '%5G%');

-- 4.12 企业信息化 (ID=206)
UPDATE `questions` q
SET q.chapter_id = 206
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%企业信息化%' OR q.content LIKE '%信息化战略%' OR q.content LIKE '%数字化转型%'
    OR q.content LIKE '%ERP%' OR q.content LIKE '%企业资源计划%' OR q.content LIKE '%CRM%'
    OR q.content LIKE '%客户关系%' OR q.content LIKE '%SCM%' OR q.content LIKE '%供应链%'
    OR q.content LIKE '%PLM%' OR q.content LIKE '%电子商务%' OR q.content LIKE '%B2B%'
    OR q.content LIKE '%B2C%' OR q.content LIKE '%C2C%' OR q.content LIKE '%O2O%'
    OR q.content LIKE '%电子政务%' OR q.content LIKE '%G2G%' OR q.content LIKE '%G2B%'
    OR q.content LIKE '%G2C%' OR q.content LIKE '%智能制造%' OR q.content LIKE '%两化融合%'
    OR q.content LIKE '%业务流程重组%' OR q.content LIKE '%BPR%' OR q.content LIKE '%BPMN%'
    OR q.content LIKE '%决策支持%' OR q.content LIKE '%DSS%' OR q.content LIKE '%商业智能%'
    OR q.content LIKE '%BI%' OR q.content LIKE '%知识管理%' OR q.content LIKE '%显性知识%'
    OR q.content LIKE '%隐性知识%' OR q.content LIKE '%数据治理%' OR q.content LIKE '%主数据%'
    OR q.content LIKE '%元数据%' OR q.content LIKE '%ITIL%' OR q.content LIKE '%ITSS%'
    OR q.content LIKE '%SLA%' OR q.content LIKE '%EAI%');

-- 4.13 项目管理 (ID=208)
UPDATE `questions` q
SET q.chapter_id = 208
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%项目管理%' OR q.content LIKE '%挣值%' OR q.content LIKE '%PV EV AC%'
    OR q.content LIKE '%成本偏差%' OR q.content LIKE '%进度偏差%' OR q.content LIKE '%CPI%'
    OR q.content LIKE '%SPI%' OR q.content LIKE '%WBS%' OR q.content LIKE '%工作包%'
    OR q.content LIKE '%范围基准%' OR q.content LIKE '%里程碑%' OR q.content LIKE '%甘特图%'
    OR q.content LIKE '%PERT%' OR q.content LIKE '%计划评审%' OR q.content LIKE '%CPM%'
    OR q.content LIKE '%关键路径%' OR q.content LIKE '%赶工%' OR q.content LIKE '%资源平衡%'
    OR q.content LIKE '%风险识别%' OR q.content LIKE '%风险应对%' OR q.content LIKE '%干系人%'
    OR q.content LIKE '%沟通渠道%' OR q.content LIKE '%团队建设%' OR q.content LIKE '%冲突解决%'
    OR q.content LIKE '%合同类型%' OR q.content LIKE '%总价合同%' OR q.content LIKE '%成本补偿%'
    OR q.content LIKE '%工料合同%' OR q.content LIKE '%索赔%' OR q.content LIKE '%紧前作业%'
    OR q.content LIKE '%总工期%' OR q.content LIKE '%总时差%' OR q.content LIKE '%自由时差%'
    OR q.content LIKE '%最早开始%' OR q.content LIKE '%最迟开始%');

-- 4.14 系统规划与分析 (ID=210)
UPDATE `questions` q
SET q.chapter_id = 210
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%系统规划%' OR q.content LIKE '%可行性研究%' OR q.content LIKE '%可行性分析%'
    OR q.content LIKE '%技术可行性%' OR q.content LIKE '%经济可行性%' OR q.content LIKE '%运行可行性%'
    OR q.content LIKE '%业务流程分析%' OR q.content LIKE '%业务流程图%' OR q.content LIKE '%系统分析报告%'
    OR q.content LIKE '%总体规划%');

-- 4.15 软件需求工程 (ID=211)
UPDATE `questions` q
SET q.chapter_id = 211
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%需求工程%' OR q.content LIKE '%需求获取%' OR q.content LIKE '%需求分类%'
    OR q.content LIKE '%需求建模%' OR q.content LIKE '%SRS%' OR q.content LIKE '%需求规格说明%'
    OR q.content LIKE '%需求评审%' OR q.content LIKE '%需求验证%' OR q.content LIKE '%需求跟踪%'
    OR q.content LIKE '%需求变更%' OR q.content LIKE '%JRP%' OR q.content LIKE '%联合需求计划%'
    OR q.content LIKE '%用例图%' OR q.content LIKE '%用例关系%' OR q.content LIKE '%用例驱动%');

-- 4.16 软件架构设计 (ID=212)
UPDATE `questions` q
SET q.chapter_id = 212
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%软件架构%' OR q.content LIKE '%架构风格%' OR q.content LIKE '%分层架构%'
    OR q.content LIKE '%管道-过滤器%' OR q.content LIKE '%MVC%' OR q.content LIKE '%微内核%'
    OR q.content LIKE '%事件驱动%' OR q.content LIKE '%SOA%' OR q.content LIKE '%云原生%'
    OR q.content LIKE '%4+1 视图%' OR q.content LIKE '%ADL%' OR q.content LIKE '%ATAM%'
    OR q.content LIKE '%SAAM%' OR q.content LIKE '%CBAM%' OR q.content LIKE '%架构评估%'
    OR q.content LIKE '%EJB%' OR q.content LIKE '%.NET%' OR q.content LIKE '%Spring%'
    OR q.content LIKE '%4+1视图%' OR q.content LIKE '%架构描述%' OR q.content LIKE '%质量属性%');

-- 4.17 系统设计 (ID=213)
UPDATE `questions` q
SET q.chapter_id = 213
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%概要设计%' OR q.content LIKE '%详细设计%' OR q.content LIKE '%模块设计%'
    OR q.content LIKE '%接口设计%' OR q.content LIKE '%数据结构设计%' OR q.content LIKE '%程序流程图%'
    OR q.content LIKE '%N-S 图%' OR q.content LIKE '%PAD 图%' OR q.content LIKE '%PDL%'
    OR q.content LIKE '%设计评审%');

-- 4.18 软件实现与测试 (ID=214)
UPDATE `questions` q
SET q.chapter_id = 214
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%编码规范%' OR q.content LIKE '%代码审查%' OR q.content LIKE '%持续集成%'
    OR q.content LIKE '%持续交付%' OR q.content LIKE '%灰度发布%' OR q.content LIKE '%CI/CD%'
    OR q.content LIKE '%调试技术%' OR q.content LIKE '%蛮力法%' OR q.content LIKE '%回溯法%'
    OR q.content LIKE '%原因排除法%' OR q.content LIKE '%演绎法%' OR q.content LIKE '%归纳法%'
    OR q.content LIKE '%测试管理%' OR q.content LIKE '%测试计划%' OR q.content LIKE '%测试报告%'
    OR q.content LIKE '%缺陷管理%' OR q.content LIKE '%白盒测试%' OR q.content LIKE '%黑盒测试%'
    OR q.content LIKE '%等价类%' OR q.content LIKE '%边界值%' OR q.content LIKE '%判定表%'
    OR q.content LIKE '%因果图%' OR q.content LIKE '%单元测试%' OR q.content LIKE '%集成测试%'
    OR q.content LIKE '%系统测试%' OR q.content LIKE '%回归测试%' OR q.content LIKE '%覆盖率%'
    OR q.content LIKE '%McCabe%' OR q.content LIKE '%环形复杂度%');

-- 4.19 系统运行与维护 (ID=215)
UPDATE `questions` q
SET q.chapter_id = 215
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%系统运行%' OR q.content LIKE '%运维管理%' OR q.content LIKE '%ITIL%'
    OR q.content LIKE '%ITSS%' OR q.content LIKE '%事件管理%' OR q.content LIKE '%系统维护%'
    OR q.content LIKE '%改正性维护%' OR q.content LIKE '%适应性维护%' OR q.content LIKE '%完善性维护%'
    OR q.content LIKE '%预防性维护%' OR q.content LIKE '%系统评价%' OR q.content LIKE '%技术债务%'
    OR q.content LIKE '%再工程%');

-- 4.20 软件工程 (ID=207) — 兜底在最广义的「软件工程」
UPDATE `questions` q
SET q.chapter_id = 207
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%软件工程%' OR q.content LIKE '%软件过程%' OR q.content LIKE '%开发模型%'
    OR q.content LIKE '%瀑布模型%' OR q.content LIKE '%原型模型%' OR q.content LIKE '%增量模型%'
    OR q.content LIKE '%螺旋模型%' OR q.content LIKE '%喷泉模型%' OR q.content LIKE '%敏捷%'
    OR q.content LIKE '%Scrum%' OR q.content LIKE '%极限编程%' OR q.content LIKE '%RUP%'
    OR q.content LIKE '%CMMI%' OR q.content LIKE '%成熟度%' OR q.content LIKE '%结构化%'
    OR q.content LIKE '%面向对象%' OR q.content LIKE '%UML%' OR q.content LIKE '%类图%'
    OR q.content LIKE '%时序图%' OR q.content LIKE '%顺序图%' OR q.content LIKE '%状态图%'
    OR q.content LIKE '%活动图%' OR q.content LIKE '%部署图%' OR q.content LIKE '%设计模式%'
    OR q.content LIKE '%单例%' OR q.content LIKE '%工厂方法%' OR q.content LIKE '%观察者%'
    OR q.content LIKE '%策略模式%' OR q.content LIKE '%适配器%' OR q.content LIKE '%装饰器%'
    OR q.content LIKE '%代理模式%' OR q.content LIKE '%外观模式%' OR q.content LIKE '%SOLID%'
    OR q.content LIKE '%开闭原则%' OR q.content LIKE '%里氏替换%' OR q.content LIKE '%依赖倒置%'
    OR q.content LIKE '%接口隔离%' OR q.content LIKE '%高内聚%' OR q.content LIKE '%低耦合%'
    OR q.content LIKE '%模块独立%' OR q.content LIKE '%逆向工程%' OR q.content LIKE '%软件再工程%'
    OR q.content LIKE '%软件度量%' OR q.content LIKE '%功能点%' OR q.content LIKE '%COCOMO%'
    OR q.content LIKE '%DFD%' OR q.content LIKE '%数据流图%' OR q.content LIKE '%数据字典%'
    OR q.content LIKE '%封装%' OR q.content LIKE '%继承%' OR q.content LIKE '%多态%'
    OR q.content LIKE '%软件工具%' OR q.content LIKE '%配置管理%' OR q.content LIKE '%基线%'
    OR q.content LIKE '%版本控制%' OR q.content LIKE '%质量保证%' OR q.content LIKE '%QA%'
    OR q.content LIKE '%QC%' OR q.content LIKE '%评审%' OR q.content LIKE '%审计%');

-- 4.21 绪论 (ID=201) — 角色定位、职业道德、生命周期
UPDATE `questions` q
SET q.chapter_id = 201
WHERE q.subject_id = 9 AND q.sub_subject_id = 4 AND q.chapter_id = 0
  AND (q.content LIKE '%系统分析师%' OR q.content LIKE '%生命周期%' OR q.content LIKE '%软件危机%'
    OR q.content LIKE '%职业道德%' OR q.content LIKE '%角色%' OR q.content LIKE '%职责%'
    OR q.content LIKE '%信息化与%' OR q.content LIKE '%知识产权%' OR q.content LIKE '%标准化%'
    OR q.content LIKE '%招投标法%' OR q.content LIKE '%政府采购法%' OR q.content LIKE '%网络安全法%'
    OR q.content LIKE '%数据安全法%' OR q.content LIKE '%个人信息保护%');

-- 4.22 兜底：未匹配任何关键词的题目 → 计算机系统 (ID=203) 作为最通用章节
UPDATE `questions`
SET `chapter_id` = 203
WHERE `subject_id` = 9 AND `sub_subject_id` = 4 AND `chapter_id` = 0;

-- 重新开启外键检查
SET FOREIGN_KEY_CHECKS = 1;
COMMIT;
