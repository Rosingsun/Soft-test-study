package service

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
	Name   string     // 节名
	Points []SAPoint  // 节内的考点
}

// SAChapter 章
type SAChapter struct {
	Name     string       // 章名（与数据库 chapter.name 一致）
	Sections []SASection  // 章下的节
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

// getChapterByName 按章节名查找大纲数据，未找到返回 nil。
func getChapterByName(name string) *SAChapter {
	for i := range systemAnalystChapters {
		if systemAnalystChapters[i].Name == name {
			return &systemAnalystChapters[i]
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

// hasChapterSyllabus 判断某科目下是否已经维护了完整的大纲。
// 未来若扩展到「软件设计师」「系统架构设计师」等其他科目，
// 可在本文件继续追加 map，让 buildGeneratePrompt 自动命中。
func hasChapterSyllabus(subjectName string) bool {
	if subjectName != "系统分析师" {
		return false
	}
	return true
}
