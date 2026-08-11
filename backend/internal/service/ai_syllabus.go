package service

import "strings"

// systemAnalystSyllabus 系统分析师《系统分析师教程（第 2 版）》综合知识 9 章
// 章→节→考点 三级结构。该数据用于在 AI 出题 prompt 中精准注入考点清单，
// 让 LLM 知道"当前是哪一章、该章有哪些必考细分知识点"，
// 避免出现与教材无关的通用泛题。
//
// 章节名严格与 seed 数据中 chapter.name 一致（见 backend/cmd/seed/main.go）。

// SAPoint 单条考点（用于注入 prompt 的最小单元）。
type SAPoint struct {
	Section string   // 节名
	Name    string   // 考点名（最终写入 knowledge_point 字段）
	Tags    []string // 关键概念词，便于 LLM 准确选题
}

// SASection 章节下的一个节
type SASection struct {
	Name   string    // 节名
	Points []SAPoint // 节内的考点
}

// SAChapter 章
type SAChapter struct {
	Name     string      // 章名（与数据库 chapter.name 一致）
	Sections []SASection // 章下的节
}

// systemAnalystChapters 9 章完整大纲（章→节→考点）。
// 节与考点按《系统分析师教程（第 2 版）》综合知识篇整理；
// 供 AI 出题 prompt 在 buildGeneratePrompt 中按章节名匹配注入。
var systemAnalystChapters = []SAChapter{
	{
		Name: "绪论",
		Sections: []SASection{
			{
				Name: "系统分析师职责与信息系统生命周期",
				Points: []SAPoint{
					{Section: "角色与职责", Name: "系统分析师角色定位与职责", Tags: []string{"系统分析师", "角色", "职责", "桥梁"}},
					{Section: "生命周期", Name: "信息系统生命周期各阶段任务", Tags: []string{"生命周期", "立项", "开发", "运维", "消亡"}},
					{Section: "需求获取", Name: "需求获取方法", Tags: []string{"访谈", "问卷", "原型法", "观察", "焦点小组"}},
				},
			},
			{
				Name: "法律法规、标准化与知识产权",
				Points: []SAPoint{
					{Section: "法律法规", Name: "信息化相关法律法规", Tags: []string{"招投标法", "合同法", "著作权法", "专利法", "网络安全法"}},
					{Section: "标准化", Name: "标准化基础与常用标准", Tags: []string{"ISO", "GB", "国标", "行标", "CMMI"}},
					{Section: "知识产权", Name: "知识产权归属与保护", Tags: []string{"著作权", "专利权", "软件登记", "侵权"}},
				},
			},
			{
				Name: "职业道德与职业规范",
				Points: []SAPoint{
					{Section: "职业道德", Name: "系统分析师职业道德", Tags: []string{"诚信", "公正", "保密", "职业操守"}},
				},
			},
		},
	},
	{
		Name: "数学与工程基础",
		Sections: []SASection{
			{
				Name: "集合论与数理逻辑",
				Points: []SAPoint{
					{Section: "集合", Name: "集合运算", Tags: []string{"并", "交", "差", "对称差", "笛卡尔积"}},
					{Section: "命题逻辑", Name: "命题逻辑与真值表", Tags: []string{"合取", "析取", "蕴含", "等价"}},
					{Section: "谓词逻辑", Name: "谓词逻辑推理", Tags: []string{"量词", "推理规则", "归结原理"}},
				},
			},
			{
				Name: "图论",
				Points: []SAPoint{
					{Section: "最小生成树", Name: "最小生成树算法", Tags: []string{"Prim", "Kruskal", "生成树"}},
					{Section: "最短路径", Name: "最短路径算法", Tags: []string{"Dijkstra", "Floyd", "Bellman-Ford"}},
					{Section: "拓扑排序", Name: "拓扑排序与关键路径", Tags: []string{"AOV", "AOE", "拓扑序列"}},
				},
			},
			{
				Name: "线性规划与运筹学",
				Points: []SAPoint{
					{Section: "线性规划", Name: "线性规划建模与单纯形法", Tags: []string{"约束", "目标函数", "单纯形", "对偶"}},
					{Section: "整数规划", Name: "整数规划与分支定界", Tags: []string{"0-1 规划", "分支定界", "割平面"}},
					{Section: "运输问题", Name: "运输问题与表上作业法", Tags: []string{"产销平衡", "初始解", "最优解判别"}},
				},
			},
			{
				Name: "排队论与博弈论",
				Points: []SAPoint{
					{Section: "排队论", Name: "排队论与 M/M/1 模型", Tags: []string{"M/M/1", "M/M/c", "Little 定律", "服务强度"}},
					{Section: "博弈论", Name: "博弈论基本概念", Tags: []string{"纳什均衡", "囚徒困境", "合作博弈"}},
				},
			},
			{
				Name: "概率论与数理统计",
				Points: []SAPoint{
					{Section: "概率分布", Name: "常见概率分布", Tags: []string{"二项分布", "泊松分布", "正态分布", "指数分布"}},
					{Section: "参数估计", Name: "参数估计与抽样", Tags: []string{"点估计", "区间估计", "置信区间"}},
					{Section: "假设检验", Name: "假设检验", Tags: []string{"原假设", "显著性", "两类错误", "p 值"}},
				},
			},
			{
				Name: "数值分析与组合数学",
				Points: []SAPoint{
					{Section: "数值分析", Name: "数值分析基础", Tags: []string{"误差", "插值", "拟合", "数值积分"}},
					{Section: "组合数学", Name: "组合计数原理", Tags: []string{"加法原理", "乘法原理", "排列", "组合", "鸽巢原理"}},
				},
			},
		},
	},
	{
		Name: "计算机系统",
		Sections: []SASection{
			{
				Name: "计算机组成与体系结构",
				Points: []SAPoint{
					{Section: "基本组成", Name: "冯·诺依曼体系结构", Tags: []string{"运算器", "控制器", "存储器", "输入", "输出"}},
					{Section: "CPU", Name: "CPU 结构与指令系统", Tags: []string{"ALU", "寄存器", "PC", "IR", "指令周期"}},
					{Section: "流水线", Name: "指令流水线与吞吐率", Tags: []string{"IF", "ID", "EX", "WB", "冒险", "吞吐率"}},
					{Section: "Cache", Name: "Cache 映射与命中率", Tags: []string{"直接映射", "组相联", "全相联", "命中率", "缺失率"}},
				},
			},
			{
				Name: "体系结构分类与并行处理",
				Points: []SAPoint{
					{Section: "CISC/RISC", Name: "CISC 与 RISC 对比", Tags: []string{"复杂指令集", "精简指令集", "流水线"}},
					{Section: "Flynn 分类", Name: "Flynn 分类法", Tags: []string{"SISD", "SIMD", "MISD", "MIMD"}},
					{Section: "多核与并行", Name: "多核与并行处理", Tags: []string{"多核", "对称多处理", "集群"}},
				},
			},
			{
				Name: "操作系统",
				Points: []SAPoint{
					{Section: "进程线程", Name: "进程与线程", Tags: []string{"PCB", "线程", "上下文切换"}},
					{Section: "同步互斥", Name: "进程同步与互斥", Tags: []string{"信号量", "管程", "PV", "临界区"}},
					{Section: "死锁", Name: "死锁与处理", Tags: []string{"死锁条件", "银行家算法", "预防", "检测"}},
					{Section: "调度", Name: "处理机调度算法", Tags: []string{"FCFS", "SJF", "时间片", "优先级", "多级反馈队列"}},
					{Section: "内存管理", Name: "虚拟内存与页面置换", Tags: []string{"分页", "分段", "FIFO", "LRU", "缺页率"}},
					{Section: "文件系统", Name: "文件管理与磁盘调度", Tags: []string{"FCFS", "SSTF", "SCAN", "FAT", "索引节点"}},
				},
			},
			{
				Name: "编译原理与嵌入式",
				Points: []SAPoint{
					{Section: "编译过程", Name: "编译过程与中间代码", Tags: []string{"词法分析", "语法分析", "语义分析", "三地址码"}},
					{Section: "嵌入式", Name: "嵌入式系统基础", Tags: []string{"实时", "固件", "ARM", "DSP"}},
				},
			},
		},
	},
	{
		Name: "计算机网络与分布式系统",
		Sections: []SASection{
			{
				Name: "网络基础与体系结构",
				Points: []SAPoint{
					{Section: "OSI/TCP-IP", Name: "OSI 七层与 TCP/IP 四层模型", Tags: []string{"物理层", "数据链路", "网络层", "传输层", "应用层"}},
					{Section: "IP 地址", Name: "IP 地址与子网划分", Tags: []string{"A 类", "B 类", "C 类", "子网掩码", "CIDR"}},
					{Section: "路由", Name: "路由协议", Tags: []string{"RIP", "OSPF", "BGP", "静态路由"}},
				},
			},
			{
				Name: "应用层与传输层协议",
				Points: []SAPoint{
					{Section: "应用层", Name: "常用应用层协议", Tags: []string{"HTTP", "HTTPS", "DNS", "FTP", "SMTP", "SNMP"}},
					{Section: "传输层", Name: "TCP 与 UDP 对比", Tags: []string{"三次握手", "四次挥手", "拥塞控制", "滑动窗口"}},
				},
			},
			{
				Name: "网络安全协议",
				Points: []SAPoint{
					{Section: "SSL/TLS", Name: "SSL/TLS 握手过程", Tags: []string{"对称加密", "非对称加密", "证书", "握手"}},
					{Section: "IPSec/VPN", Name: "IPSec 与 VPN", Tags: []string{"隧道", "ESP", "AH", "VPN"}},
				},
			},
			{
				Name: "分布式系统理论",
				Points: []SAPoint{
					{Section: "CAP", Name: "CAP 定理与 BASE 理论", Tags: []string{"一致性", "可用性", "分区容错", "最终一致"}},
					{Section: "共识算法", Name: "分布式共识算法", Tags: []string{"Paxos", "Raft", "两阶段提交", "三阶段提交"}},
					{Section: "一致性哈希", Name: "一致性哈希", Tags: []string{"哈希环", "虚拟节点", "数据分片"}},
				},
			},
			{
				Name: "微服务与中间件",
				Points: []SAPoint{
					{Section: "微服务", Name: "微服务架构特点", Tags: []string{"服务拆分", "独立部署", "服务治理", "去中心化"}},
					{Section: "消息中间件", Name: "消息队列与发布订阅", Tags: []string{"Kafka", "RabbitMQ", "RocketMQ", "发布订阅"}},
					{Section: "网关", Name: "API 网关与限流熔断", Tags: []string{"网关", "限流", "熔断", "服务降级"}},
				},
			},
		},
	},
	{
		Name: "数据库系统",
		Sections: []SASection{
			{
				Name: "关系模型与数据库设计",
				Points: []SAPoint{
					{Section: "E-R 模型", Name: "E-R 图与关系模式转换", Tags: []string{"实体", "属性", "联系", "1:1", "1:n", "m:n"}},
					{Section: "关系模型", Name: "关系模型与关系代数", Tags: []string{"选择", "投影", "连接", "并", "差", "笛卡尔积"}},
					{Section: "范式", Name: "范式理论 1NF/2NF/3NF/BCNF", Tags: []string{"函数依赖", "码", "范式", "分解"}},
				},
			},
			{
				Name: "SQL 与数据库对象",
				Points: []SAPoint{
					{Section: "SQL", Name: "SQL DDL/DML/DQL", Tags: []string{"SELECT", "INSERT", "UPDATE", "CREATE", "授权"}},
					{Section: "视图", Name: "视图与物化视图", Tags: []string{"视图", "物化视图", "可更新视图"}},
					{Section: "索引", Name: "索引结构 B+ 树与哈希", Tags: []string{"聚簇索引", "非聚簇索引", "B+ 树", "哈希"}},
				},
			},
			{
				Name: "事务与并发控制",
				Points: []SAPoint{
					{Section: "ACID", Name: "事务的 ACID 特性", Tags: []string{"原子性", "一致性", "隔离性", "持久性"}},
					{Section: "并发控制", Name: "并发控制与封锁协议", Tags: []string{"共享锁", "排他锁", "两段锁", "活锁", "死锁"}},
					{Section: "隔离级别", Name: "事务隔离级别", Tags: []string{"读未提交", "读已提交", "可重复读", "串行化"}},
				},
			},
			{
				Name: "数据仓库与大数据",
				Points: []SAPoint{
					{Section: "数据仓库", Name: "数据仓库与维度建模", Tags: []string{"事实表", "维度表", "星型模型", "雪花模型"}},
					{Section: "OLAP", Name: "OLAP 多维分析", Tags: []string{"钻取", "切片", "旋转", "上卷"}},
					{Section: "NoSQL", Name: "NoSQL 数据库分类", Tags: []string{"键值", "文档", "列式", "图数据库"}},
					{Section: "大数据", Name: "大数据与分布式数据库", Tags: []string{"Hadoop", "HDFS", "MapReduce", "分库分表"}},
				},
			},
		},
	},
	{
		Name: "企业信息化",
		Sections: []SASection{
			{
				Name: "信息化战略与规划",
				Points: []SAPoint{
					{Section: "信息化战略", Name: "企业信息化战略与规划", Tags: []string{"战略", "规划", "数字化转型"}},
				},
			},
			{
				Name: "应用系统",
				Points: []SAPoint{
					{Section: "ERP", Name: "ERP 与企业资源计划", Tags: []string{"财务", "采购", "库存", "生产", "销售"}},
					{Section: "CRM", Name: "CRM 客户关系管理", Tags: []string{"客户", "营销", "服务", "销售自动化"}},
					{Section: "SCM/HRM", Name: "SCM 与 HRM 系统", Tags: []string{"供应链", "人力资源", "物流"}},
				},
			},
			{
				Name: "电子商务与电子政务",
				Points: []SAPoint{
					{Section: "电子商务", Name: "电子商务模式", Tags: []string{"B2B", "B2C", "C2C", "O2O", "C2M"}},
					{Section: "电子政务", Name: "电子政务体系结构", Tags: []string{"G2G", "G2B", "G2C", "G2E"}},
				},
			},
			{
				Name: "决策支持与商业智能",
				Points: []SAPoint{
					{Section: "DSS", Name: "决策支持系统 DSS", Tags: []string{"模型库", "数据库", "方法库", "人机对话"}},
					{Section: "BI", Name: "商业智能 BI", Tags: []string{"数据可视化", "报表", "指标体系"}},
				},
			},
			{
				Name: "业务流程与知识管理",
				Points: []SAPoint{
					{Section: "BPR", Name: "业务流程重组 BPR", Tags: []string{"BPR", "流程再造", "以流程为中心"}},
					{Section: "BPMN", Name: "BPMN 流程建模", Tags: []string{"BPMN", "事件", "网关", "活动"}},
					{Section: "知识管理", Name: "知识管理与 KMS", Tags: []string{"显性知识", "隐性知识", "知识库"}},
					{Section: "数据治理", Name: "数据治理与元数据", Tags: []string{"数据治理", "元数据", "主数据", "数据质量"}},
				},
			},
		},
	},
	{
		Name: "软件工程",
		Sections: []SASection{
			{
				Name: "需求工程",
				Points: []SAPoint{
					{Section: "需求获取", Name: "需求获取与需求分类", Tags: []string{"访谈", "问卷", "原型", "功能需求", "非功能需求"}},
					{Section: "用例建模", Name: "用例图与用例规约", Tags: []string{"参与者", "用例", "关联", "包含", "扩展"}},
					{Section: "需求验证", Name: "需求验证与需求管理", Tags: []string{"需求评审", "需求跟踪", "需求变更"}},
				},
			},
			{
				Name: "面向对象分析与设计",
				Points: []SAPoint{
					{Section: "UML", Name: "UML 各种图", Tags: []string{"用例图", "类图", "时序图", "活动图", "状态图", "组件图", "部署图"}},
					{Section: "OOAD", Name: "面向对象分析与设计原则", Tags: []string{"封装", "继承", "多态", "SOLID", "GRASP"}},
				},
			},
			{
				Name: "设计模式",
				Points: []SAPoint{
					{Section: "GoF 模式", Name: "GoF 23 种设计模式", Tags: []string{"单例", "工厂", "观察者", "策略", "适配器", "装饰器", "代理"}},
					{Section: "模式应用", Name: "设计模式典型应用场景", Tags: []string{"MVC", "DAO", "外观模式", "组合模式"}},
				},
			},
			{
				Name: "软件测试",
				Points: []SAPoint{
					{Section: "测试分类", Name: "测试阶段与测试方法", Tags: []string{"单元测试", "集成测试", "系统测试", "验收测试", "回归测试"}},
					{Section: "测试用例", Name: "测试用例设计方法", Tags: []string{"等价类", "边界值", "因果图", "判定表", "场景法"}},
					{Section: "测试覆盖", Name: "测试覆盖率与质量度量", Tags: []string{"语句覆盖", "分支覆盖", "路径覆盖", "McCabe"}},
				},
			},
			{
				Name: "软件维护与重构",
				Points: []SAPoint{
					{Section: "维护类型", Name: "软件维护类型", Tags: []string{"改正性", "适应性", "完善性", "预防性"}},
					{Section: "重构", Name: "软件重构", Tags: []string{"代码重构", "坏味道", "重构时机"}},
				},
			},
			{
				Name: "软件过程与项目管理",
				Points: []SAPoint{
					{Section: "开发模型", Name: "软件开发模型", Tags: []string{"瀑布", "增量", "迭代", "原型", "螺旋", "敏捷"}},
					{Section: "敏捷", Name: "敏捷开发 Scrum", Tags: []string{"Sprint", "Backlog", "Stand-up", "Retrospective"}},
					{Section: "CMMI", Name: "CMMI 能力成熟度模型", Tags: []string{"初始级", "可重复级", "已定义级", "定量管理", "优化级"}},
				},
			},
			{
				Name: "配置管理",
				Points: []SAPoint{
					{Section: "配置项", Name: "软件配置管理与基线", Tags: []string{"配置项", "基线", "版本控制", "变更控制"}},
				},
			},
		},
	},
	{
		Name: "项目管理",
		Sections: []SASection{
			{
				Name: "项目整体管理与范围",
				Points: []SAPoint{
					{Section: "整体管理", Name: "项目整体管理过程", Tags: []string{"章程", "计划", "指导", "监控", "收尾"}},
					{Section: "WBS", Name: "范围管理与 WBS 分解", Tags: []string{"WBS", "工作包", "范围基准"}},
				},
			},
			{
				Name: "进度管理",
				Points: []SAPoint{
					{Section: "网络图", Name: "网络图与关键路径", Tags: []string{"PDM", "ADM", "关键路径", "CPM"}},
					{Section: "PERT", Name: "PERT 计划评审技术", Tags: []string{"三点估算", "乐观", "最可能", "悲观"}},
					{Section: "甘特图", Name: "甘特图与里程碑", Tags: []string{"甘特图", "里程碑", "资源平衡"}},
				},
			},
			{
				Name: "成本管理",
				Points: []SAPoint{
					{Section: "估算", Name: "项目成本估算方法", Tags: []string{"类比估算", "参数估算", "三点估算", "自下而上"}},
					{Section: "挣值", Name: "挣值分析 PV/EV/AC", Tags: []string{"PV", "EV", "AC", "CV", "SV", "CPI", "SPI"}},
				},
			},
			{
				Name: "质量、风险与沟通",
				Points: []SAPoint{
					{Section: "质量管理", Name: "质量管理工具与技术", Tags: []string{"七种工具", "因果图", "帕累托", "控制图", "质量保证"}},
					{Section: "风险管理", Name: "项目风险管理", Tags: []string{"风险识别", "定性分析", "定量分析", "风险应对"}},
					{Section: "沟通管理", Name: "项目沟通与团队管理", Tags: []string{"沟通模型", "冲突管理", "团队建设"}},
				},
			},
			{
				Name: "采购与合同",
				Points: []SAPoint{
					{Section: "招投标", Name: "招投标流程", Tags: []string{"招标", "投标", "评标", "中标"}},
					{Section: "合同类型", Name: "合同类型与计价方式", Tags: []string{"固定价", "成本补偿", "工料合同"}},
				},
			},
		},
	},
	{
		Name: "信息安全",
		Sections: []SASection{
			{
				Name: "密码学基础",
				Points: []SAPoint{
					{Section: "对称加密", Name: "对称加密算法", Tags: []string{"DES", "3DES", "AES", "分组密码", "流密码"}},
					{Section: "非对称加密", Name: "非对称加密算法", Tags: []string{"RSA", "ECC", "ElGamal", "公钥私钥"}},
					{Section: "哈希", Name: "哈希函数与摘要", Tags: []string{"MD5", "SHA-1", "SHA-256", "碰撞"}},
				},
			},
			{
				Name: "数字签名与 PKI",
				Points: []SAPoint{
					{Section: "数字签名", Name: "数字签名原理", Tags: []string{"签名", "验证", "不可否认性"}},
					{Section: "PKI", Name: "PKI 与数字证书", Tags: []string{"CA", "RA", "X.509", "证书链"}},
				},
			},
			{
				Name: "访问控制",
				Points: []SAPoint{
					{Section: "访问控制模型", Name: "访问控制模型", Tags: []string{"ACL", "RBAC", "MAC", "DAC", "ABAC"}},
				},
			},
			{
				Name: "网络安全",
				Points: []SAPoint{
					{Section: "防火墙", Name: "防火墙类型与工作原理", Tags: []string{"包过滤", "状态检测", "代理防火墙", "WAF"}},
					{Section: "IDS/IPS", Name: "入侵检测与入侵防御", Tags: []string{"误用检测", "异常检测", "Snort", "签名"}},
					{Section: "VPN", Name: "VPN 与远程安全接入", Tags: []string{"IPSec VPN", "SSL VPN", "MPLS VPN"}},
				},
			},
			{
				Name: "应用与系统安全",
				Points: []SAPoint{
					{Section: "Web 安全", Name: "常见 Web 攻击", Tags: []string{"SQL 注入", "XSS", "CSRF", "SSRF", "文件上传"}},
					{Section: "系统安全", Name: "系统安全与提权", Tags: []string{"缓冲区溢出", "提权", "后门", "弱口令"}},
				},
			},
			{
				Name: "安全管理与法规",
				Points: []SAPoint{
					{Section: "安全策略", Name: "信息安全策略与管理", Tags: []string{"安全策略", "安全审计", "风险评估"}},
					{Section: "等级保护", Name: "等级保护与灾备", Tags: []string{"等保 2.0", "五级", "灾备", "RPO", "RTO"}},
				},
			},
		},
	},
}

// 软件设计师精简大纲（章→节 2 层，每节放 1~2 个核心考点关键词）。
// 命中后 prompt 会注入「章→节→核心考点」，比纯文字约束更精准。
var softwareDesignerChapters = []SAChapter{
	{Name: "计算机组成与体系结构", Sections: []SASection{
		{Name: "数据的表示与运算", Points: []SAPoint{{Section: "数据表示", Name: "进制转换与原码/反码/补码"}}},
		{Name: "存储系统", Points: []SAPoint{{Section: "存储", Name: "Cache 映射与主存编址"}}},
		{Name: "指令系统与流水线", Points: []SAPoint{{Section: "流水线", Name: "指令流水线与吞吐率"}}},
	}},
	{Name: "操作系统", Sections: []SASection{
		{Name: "进程管理", Points: []SAPoint{{Section: "进程", Name: "进程同步与死锁"}}},
		{Name: "存储管理", Points: []SAPoint{{Section: "存储", Name: "分页存储与页面置换"}}},
		{Name: "文件系统", Points: []SAPoint{{Section: "文件", Name: "文件组织与目录结构"}}},
	}},
	{Name: "程序设计语言基础", Sections: []SASection{
		{Name: "文法与语言", Points: []SAPoint{{Section: "文法", Name: "正规式与上下文无关文法"}}},
		{Name: "编译过程", Points: []SAPoint{{Section: "编译", Name: "词法分析与语法分析"}}},
	}},
	{Name: "数据结构与算法", Sections: []SASection{
		{Name: "线性结构", Points: []SAPoint{{Section: "线性表", Name: "链表与栈队列"}}},
		{Name: "树与二叉树", Points: []SAPoint{{Section: "树", Name: "二叉树遍历与 Huffman"}}},
		{Name: "图", Points: []SAPoint{{Section: "图", Name: "最小生成树与最短路径"}}},
		{Name: "排序与查找", Points: []SAPoint{{Section: "排序", Name: "排序算法复杂度对比"}}},
	}},
	{Name: "软件工程", Sections: []SASection{
		{Name: "软件过程与开发模型", Points: []SAPoint{{Section: "过程", Name: "瀑布/增量/敏捷"}}},
		{Name: "需求与设计", Points: []SAPoint{{Section: "设计", Name: "结构化与面向对象设计"}}},
		{Name: "测试", Points: []SAPoint{{Section: "测试", Name: "白盒/黑盒测试用例设计"}}},
		{Name: "维护与重构", Points: []SAPoint{{Section: "维护", Name: "维护类型与 McCabe 复杂度"}}},
	}},
	{Name: "面向对象技术", Sections: []SASection{
		{Name: "面向对象基础", Points: []SAPoint{{Section: "OO", Name: "封装/继承/多态"}}},
		{Name: "UML 与设计模式", Points: []SAPoint{{Section: "UML", Name: "UML 图与 GoF 模式"}}},
	}},
	{Name: "数据库系统", Sections: []SASection{
		{Name: "关系模型与 SQL", Points: []SAPoint{{Section: "SQL", Name: "SQL DDL/DML/DQL"}}},
		{Name: "范式与设计", Points: []SAPoint{{Section: "范式", Name: "1NF/2NF/3NF/BCNF"}}},
		{Name: "事务与并发", Points: []SAPoint{{Section: "事务", Name: "ACID 与封锁协议"}}},
	}},
	{Name: "计算机网络", Sections: []SASection{
		{Name: "网络体系结构", Points: []SAPoint{{Section: "网络", Name: "OSI/TCP-IP 模型"}}},
		{Name: "协议与 IP", Points: []SAPoint{{Section: "IP", Name: "IP 地址与子网划分"}}},
		{Name: "应用层", Points: []SAPoint{{Section: "应用", Name: "HTTP/DNS/FTP/SMTP"}}},
	}},
	{Name: "信息安全", Sections: []SASection{
		{Name: "加密与签名", Points: []SAPoint{{Section: "加密", Name: "对称/非对称加密与摘要"}}},
		{Name: "网络安全", Points: []SAPoint{{Section: "安全", Name: "防火墙/IDS/IPS 与 Web 攻击"}}},
	}},
}

// 信息系统项目管理师精简大纲（10 大知识领域 + 5 大过程组）
var projectManagerChapters = []SAChapter{
	{Name: "项目立项与整合", Sections: []SASection{
		{Name: "立项管理", Points: []SAPoint{{Section: "立项", Name: "可行性研究与招投标"}}},
		{Name: "整体管理", Points: []SAPoint{{Section: "整合", Name: "章程、初步范围说明书、变更控制"}}},
	}},
	{Name: "项目范围与进度", Sections: []SASection{
		{Name: "范围管理", Points: []SAPoint{{Section: "范围", Name: "WBS 分解与范围基准"}}},
		{Name: "进度管理", Points: []SAPoint{{Section: "进度", Name: "关键路径与 PERT 三点估算"}}},
	}},
	{Name: "项目成本与质量", Sections: []SASection{
		{Name: "成本管理", Points: []SAPoint{{Section: "成本", Name: "挣值分析 PV/EV/AC/CPI/SPI"}}},
		{Name: "质量管理", Points: []SAPoint{{Section: "质量", Name: "七种工具与质量保证"}}},
	}},
	{Name: "项目资源与沟通", Sections: []SASection{
		{Name: "资源管理", Points: []SAPoint{{Section: "资源", Name: "团队建设与冲突管理"}}},
		{Name: "沟通管理", Points: []SAPoint{{Section: "沟通", Name: "沟通模型与干系人管理"}}},
	}},
	{Name: "项目风险与采购", Sections: []SASection{
		{Name: "风险管理", Points: []SAPoint{{Section: "风险", Name: "风险识别、定性/定量分析、应对"}}},
		{Name: "采购管理", Points: []SAPoint{{Section: "采购", Name: "合同类型与招投标流程"}}},
	}},
	{Name: "项目绩效与收尾", Sections: []SASection{
		{Name: "绩效域", Points: []SAPoint{{Section: "绩效", Name: "PMBOK 8 大绩效域"}}},
		{Name: "收尾管理", Points: []SAPoint{{Section: "收尾", Name: "合同收尾与经验教训"}}},
	}},
	{Name: "信息系统基础", Sections: []SASection{
		{Name: "信息化与系统", Points: []SAPoint{{Section: "信息化", Name: "信息系统生命周期与 ERP/CRM/SCM"}}},
		{Name: "安全与运维", Points: []SAPoint{{Section: "运维", Name: "ITIL、配置管理与变更控制"}}},
	}},
	{Name: "管理科学基础", Sections: []SASection{
		{Name: "运筹学", Points: []SAPoint{{Section: "运筹", Name: "线性规划与最短路径"}}},
		{Name: "决策分析", Points: []SAPoint{{Section: "决策", Name: "决策树与盈亏平衡分析"}}},
	}},
}

// 系统架构设计师精简大纲
var architectDesignerChapters = []SAChapter{
	{Name: "操作系统与硬件基础", Sections: []SASection{
		{Name: "操作系统", Points: []SAPoint{{Section: "OS", Name: "进程/存储/文件系统"}}},
		{Name: "计算机体系结构", Points: []SAPoint{{Section: "体系", Name: "CISC/RISC、流水线、Cache"}}},
	}},
	{Name: "软件架构基础", Sections: []SASection{
		{Name: "架构概念", Points: []SAPoint{{Section: "架构", Name: "架构 4+1 视图与生命周期"}}},
		{Name: "架构风格", Points: []SAPoint{{Section: "风格", Name: "分层/管道-过滤器/MVC/微服务"}}},
		{Name: "设计模式", Points: []SAPoint{{Section: "模式", Name: "GoF 模式与架构模式"}}},
	}},
	{Name: "架构设计方法", Sections: []SASection{
		{Name: "需求与领域", Points: []SAPoint{{Section: "领域", Name: "领域驱动设计 DDD"}}},
		{Name: "架构评估", Points: []SAPoint{{Section: "评估", Name: "ATAM/SAAM/CBAM"}}},
	}},
	{Name: "分布式系统架构", Sections: []SASection{
		{Name: "分布式理论", Points: []SAPoint{{Section: "分布", Name: "CAP/BASE/最终一致"}}},
		{Name: "微服务", Points: []SAPoint{{Section: "微服务", Name: "服务治理、限流熔断、Saga"}}},
		{Name: "中间件", Points: []SAPoint{{Section: "中间件", Name: "消息队列/缓存/分布式事务"}}},
	}},
	{Name: "系统可靠性与安全", Sections: []SASection{
		{Name: "可靠性", Points: []SAPoint{{Section: "可靠性", Name: "冗余/容错/故障转移"}}},
		{Name: "信息安全架构", Points: []SAPoint{{Section: "安全", Name: "PKI/访问控制/审计"}}},
	}},
	{Name: "数据库与持久化", Sections: []SASection{
		{Name: "数据库架构", Points: []SAPoint{{Section: "DB", Name: "分库分表/读写分离/连接池"}}},
		{Name: "大数据架构", Points: []SAPoint{{Section: "大数据", Name: "Hadoop/Kafka/数据湖"}}},
	}},
	{Name: "系统性能与运维", Sections: []SASection{
		{Name: "性能优化", Points: []SAPoint{{Section: "性能", Name: "负载均衡/CDN/缓存策略"}}},
		{Name: "运维架构", Points: []SAPoint{{Section: "运维", Name: "DevOps/监控/日志"}}},
	}},
	{Name: "未来架构趋势", Sections: []SASection{
		{Name: "云原生", Points: []SAPoint{{Section: "云", Name: "容器/Service Mesh/Serverless"}}},
		{Name: "AI 架构", Points: []SAPoint{{Section: "AI", Name: "MLOps 与智能应用架构"}}},
	}},
}

// 网络工程师精简大纲
var networkEngineerChapters = []SAChapter{
	{Name: "计算机网络基础", Sections: []SASection{
		{Name: "体系结构", Points: []SAPoint{{Section: "体系", Name: "OSI/TCP-IP 模型"}}},
		{Name: "编码与传输", Points: []SAPoint{{Section: "传输", Name: "曼彻斯特编码与多路复用"}}},
	}},
	{Name: "数据链路层", Sections: []SASection{
		{Name: "差错控制", Points: []SAPoint{{Section: "差错", Name: "CRC 校验与海明码"}}},
		{Name: "介质访问", Points: []SAPoint{{Section: "MAC", Name: "CSMA/CD 与以太网帧结构"}}},
		{Name: "交换技术", Points: []SAPoint{{Section: "交换", Name: "VLAN/STP/Trunk"}}},
	}},
	{Name: "网络层", Sections: []SASection{
		{Name: "IP 地址", Points: []SAPoint{{Section: "IP", Name: "子网划分与 CIDR"}}},
		{Name: "路由协议", Points: []SAPoint{{Section: "路由", Name: "RIP/OSPF/BGP"}}},
		{Name: "ICMP 与 ARP", Points: []SAPoint{{Section: "协议", Name: "ICMP/ARP/DHCP"}}},
	}},
	{Name: "传输层与应用层", Sections: []SASection{
		{Name: "TCP/UDP", Points: []SAPoint{{Section: "传输", Name: "三次握手/四次挥手/拥塞控制"}}},
		{Name: "应用协议", Points: []SAPoint{{Section: "应用", Name: "HTTP/DNS/FTP/SMTP/SNMP"}}},
	}},
	{Name: "网络安全", Sections: []SASection{
		{Name: "加密与签名", Points: []SAPoint{{Section: "加密", Name: "对称/非对称/数字签名"}}},
		{Name: "防火墙与 VPN", Points: []SAPoint{{Section: "安全", Name: "防火墙/IPSec/SSL VPN"}}},
		{Name: "入侵检测", Points: []SAPoint{{Section: "IDS", Name: "IDS/IPS 与蜜罐"}}},
	}},
	{Name: "网络管理与运维", Sections: []SASection{
		{Name: "网络管理", Points: []SAPoint{{Section: "管理", Name: "SNMP/RMON"}}},
		{Name: "故障排查", Points: []SAPoint{{Section: "故障", Name: "ping/traceroute/抓包分析"}}},
	}},
	{Name: "广域网与接入", Sections: []SASection{
		{Name: "WAN 技术", Points: []SAPoint{{Section: "WAN", Name: "PPP/HDLC/帧中继/MPLS"}}},
		{Name: "接入技术", Points: []SAPoint{{Section: "接入", Name: "xDSL/PON/无线接入"}}},
	}},
	{Name: "网络规划与设计", Sections: []SASection{
		{Name: "网络规划", Points: []SAPoint{{Section: "规划", Name: "需求分析与分级设计"}}},
		{Name: "结构化布线", Points: []SAPoint{{Section: "布线", Name: "综合布线与机房"}}},
	}},
	{Name: "服务器与操作系统", Sections: []SASection{
		{Name: "Windows Server", Points: []SAPoint{{Section: "Win", Name: "AD/DNS/DHCP 配置"}}},
		{Name: "Linux", Points: []SAPoint{{Section: "Linux", Name: "常用命令与网络配置"}}},
	}},
	{Name: "新技术与趋势", Sections: []SASection{
		{Name: "IPv6", Points: []SAPoint{{Section: "IPv6", Name: "IPv6 地址与过渡技术"}}},
		{Name: "SDN/SD-WAN", Points: []SAPoint{{Section: "SDN", Name: "OpenFlow 与软件定义网络"}}},
	}},
}

// subjectSyllabusMap 科目→大纲 映射。
// 命中此 map 的科目，prompt 会注入章→节→考点 三级详细清单。
// 未命中的科目走 fallbackSubjectGuard（按所选科目的典型考域硬约束 LLM）。
var subjectSyllabusMap = map[string][]SAChapter{
	"系统分析师":     systemAnalystChapters,
	"软件设计师":     softwareDesignerChapters,
	"信息系统项目管理师": projectManagerChapters,
	"系统架构设计师":   architectDesignerChapters,
	"网络工程师":     networkEngineerChapters,
}

// getChapterByName 按章节名查找大纲数据，未找到返回 nil。
// 名称做 TrimSpace 处理，避免 seed 数据带前后空白时匹配失败。
func getChapterByName(name string) *SAChapter {
	name = strings.TrimSpace(name)
	for i := range systemAnalystChapters {
		if strings.TrimSpace(systemAnalystChapters[i].Name) == name {
			return &systemAnalystChapters[i]
		}
	}
	return nil
}

// getChapterByNameForSubject 在指定科目的大纲里按章节名查找
func getChapterByNameForSubject(subject, chapter string) *SAChapter {
	chs := getChaptersForSubject(subject)
	chapter = strings.TrimSpace(chapter)
	for i := range chs {
		if strings.TrimSpace(chs[i].Name) == chapter {
			return &chs[i]
		}
	}
	return nil
}

// getAllChapterPoints 返回某章节下所有考点的扁平列表。
// 用于在不区分节的情况下，告知 LLM 整章的全部可选考点。
func getAllChapterPoints(chapter *SAChapter) []SAPoint {
	if chapter == nil {
		return nil
	}
	var points []SAPoint
	for _, sec := range chapter.Sections {
		points = append(points, sec.Points...)
	}
	return points
}

// formatChapterKPBlock 把考点清单格式化为 "- 节名 / 考点名" 形式，
// 直接注入 prompt。带缩进/换行，便于 LLM 阅读。
func formatChapterKPBlock(chapter *SAChapter) string {
	if chapter == nil {
		return "（未限定）"
	}
	var b []byte
	b = append(b, "本章考点清单（必须从下列考点中选取 1~3 个作为本题核心考点）：\n"...)
	for _, sec := range chapter.Sections {
		b = append(b, "  "...)
		b = append(b, sec.Name...)
		b = append(b, "：\n"...)
		for _, p := range sec.Points {
			b = append(b, "    - "...)
			b = append(b, p.Name...)
			b = append(b, '\n')
		}
	}
	return string(b)
}

// formatChapterOutlineBlock 格式化"章→节"两层结构（不展开考点）
// 用于"不限定章节"模式：避免 prompt 体积过大，又给 LLM 足够的范围参考。
func formatChapterOutlineBlock(chapters []SAChapter) string {
	var b []byte
	b = append(b, "未限定具体章节，请从以下章→节中任选知识点命制（knowledge_point 填写细分考点名）：\n"...)
	for _, ch := range chapters {
		b = append(b, "- "...)
		b = append(b, ch.Name...)
		b = append(b, '\n')
		for _, sec := range ch.Sections {
			b = append(b, "    · "...)
			b = append(b, sec.Name...)
			b = append(b, '\n')
		}
	}
	return string(b)
}

// hasChapterSyllabus 判断某科目是否已维护完整的大纲。
// 命中后 prompt 会注入章→节→考点 三级详细清单。
// 未命中时使用 fallbackSubjectGuard 给出"按所选科目考纲"硬约束。
func hasChapterSyllabus(subjectName string) bool {
	_, ok := subjectSyllabusMap[strings.TrimSpace(subjectName)]
	return ok
}

// getChaptersForSubject 返回某科目下的章列表。
// 大纲命中的科目返回完整数据；未命中的返回空切片。
func getChaptersForSubject(subjectName string) []SAChapter {
	ch, ok := subjectSyllabusMap[strings.TrimSpace(subjectName)]
	if !ok {
		return nil
	}
	return ch
}

// fallbackSubjectGuard 对未维护详细大纲的科目，给出"严格按所选科目"硬约束。
// 即使用户选了"软件设计师"，prompt 也会明确禁止 LLM 出系统分析师/网工等其他科目的题。
func fallbackSubjectGuard(subjectName string) string {
	s := strings.TrimSpace(subjectName)
	if s == "" {
		return `【科目录入】未选择科目。请立即返回一个空数组 []，不要生成任何题目。`
	}
	// 命中大纲的科目不走保底
	if hasChapterSyllabus(s) {
		return ""
	}
	return `【科目录入·硬约束·必读】
1. 本次出题科目为《` + s + `》。你必须按《` + s + `》官方考试大纲范围命制题目，**严禁**串到其他科目（如系统分析师、软件设计师、网络工程师等）的考点。
2. 典型考域（请据此判断题目是否在范围内）：
   - ` + s + ` 通常考察：` + typicalScopeFor(s) + `
3. 如果你无法确定某题属于《` + s + `》考纲，请改命其他题目；不要硬出与本科目无关的题。
4. knowledge_point 字段必须填写《` + s + `》考纲内能查到的细分考点名（如该科目无明确细分考点可填"综合应用"）。`
}

// typicalScopeFor 返回某科目（未维护详细大纲）的典型考域描述。
// 用于 fallbackSubjectGuard 作为 LLM 范围参考。
func typicalScopeFor(subjectName string) string {
	switch strings.TrimSpace(subjectName) {
	case "软件设计师":
		return "数据结构与算法、操作系统、计算机网络、数据库系统、软件工程（需求、设计、测试、面向对象）、面向对象程序设计（C++/Java）、信息安全基础、标准化与知识产权"
	case "网络工程师":
		return "计算机网络体系结构（OSI/TCP-IP）、局域网与广域网、TCP/IP 协议簇、路由与交换、网络安全、网络管理（SNMP）、网络规划与设计、Windows/Linux 系统管理"
	case "信息系统项目管理师":
		return "项目整体/范围/进度/成本/质量/资源/沟通/风险/采购/干系人 十大管理、PMBOK 过程组、挣值分析、配置管理、变更管理、信息系统生命周期、立项管理"
	case "系统架构设计师":
		return "架构风格（分层/微服务/事件驱动/C/S/SOA）、架构评估（ATAM/SAAM）、设计模式、信息安全架构、分布式系统架构、可靠性/可用性/可维护性、数据库架构、缓存与消息中间件"
	case "网络规划设计师":
		return "网络规划与设计流程、网络需求分析、逻辑与物理设计、网络优化、网络安全方案、IP 地址规划与子网划分、路由协议、园区网/广域网/数据中心网络"
	case "数据库系统工程师":
		return "关系数据库理论、SQL 语法、数据库设计（ER/范式）、事务与并发控制、备份与恢复、NoSQL、分布式数据库、数据仓库与数据挖掘、数据库性能优化"
	case "信息系统管理工程师":
		return "信息系统建设、运维管理（ITIL）、系统故障与备份、信息安全策略、人员与管理、监理与评估、数据库与网络管理基础"
	case "程序员":
		return "程序设计语言基础（C/Java）、数据结构、算法、软件工程基础、面向对象、操作系统基础、计算机网络基础、数据库基础"
	case "网络管理员":
		return "网络基础、网络设备配置（交换机/路由器）、操作系统管理（Windows/Linux）、网络安全基础、网络故障排除、局域网组建"
	case "信息处理技术员":
		return "信息技术基础、办公软件（Word/Excel/PPT）、信息处理实务、数据采集与分析、信息安全意识"
	default:
		return subjectName + " 官方考纲范围内的常见考点"
	}
}
