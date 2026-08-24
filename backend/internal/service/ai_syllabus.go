package service

import "strings"

// systemAnalystSyllabus 系统分析师综合知识 22 章
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

// systemAnalystChapters 系统分析师综合知识 22 章完整大纲（章→节→考点）。
// 章节按官方考纲知识域拆分，与 seed 数据 chapter.name 一一对应；
// 节与考点按《系统分析师教程（第 2 版 · 2024）》清华大学出版社综合知识篇整理；
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
					{Section: "职业道德", Name: "系统分析师职业道德与职业规范", Tags: []string{"诚信", "公正", "保密", "职业操守"}},
				},
			},
			{
				Name: "信息化与法律法规基础",
				Points: []SAPoint{
					{Section: "知识产权", Name: "知识产权归属与保护", Tags: []string{"著作权", "专利权", "商标权", "职务作品", "委托开发归属", "保护期限"}},
					{Section: "标准化", Name: "标准化基础与常用标准", Tags: []string{"ISO", "GB", "国标", "行标", "强制性标准", "推荐性标准", "CMMI"}},
					{Section: "法律法规", Name: "信息化相关法律法规", Tags: []string{"招投标法", "政府采购法", "合同法", "网络安全法", "数据安全法", "个人信息保护法"}},
				},
			},
		},
	},
	{
		Name: "数学与工程基础",
		Sections: []SASection{
			{
				Name: "工程数学",
				Points: []SAPoint{
					{Section: "高等数学", Name: "微积分与函数基础", Tags: []string{"极限", "导数", "积分", "多元函数"}},
					{Section: "线性代数", Name: "矩阵运算与线性方程组", Tags: []string{"行列式", "逆矩阵", "秩", "特征值"}},
					{Section: "概率统计", Name: "概率论与数理统计", Tags: []string{"数学期望", "方差", "正态分布", "二项分布", "参数估计", "假设检验"}},
				},
			},
			{
				Name: "离散数学",
				Points: []SAPoint{
					{Section: "集合与关系", Name: "集合运算与关系", Tags: []string{"并", "交", "差", "笛卡尔积", "等价关系", "偏序"}},
					{Section: "数理逻辑", Name: "命题逻辑与谓词逻辑", Tags: []string{"合取", "析取", "蕴含", "真值表", "归结原理"}},
					{Section: "图论", Name: "图论基础", Tags: []string{"最小生成树", "Prim", "Kruskal", "最短路径 Dijkstra", "Floyd", "拓扑排序"}},
				},
			},
			{
				Name: "运筹学与决策分析",
				Points: []SAPoint{
					{Section: "线性规划", Name: "线性规划建模与单纯形法", Tags: []string{"约束", "目标函数", "单纯形", "对偶", "0-1 规划"}},
					{Section: "运输与指派", Name: "运输问题与指派问题", Tags: []string{"产销平衡", "表上作业法", "匈牙利法"}},
					{Section: "排队与博弈", Name: "排队论与博弈论", Tags: []string{"M/M/1", "Little 定律", "纳什均衡", "囚徒困境"}},
					{Section: "决策分析", Name: "决策树与不确定性决策", Tags: []string{"期望收益", "乐观法", "悲观法", "后悔值", "盈亏平衡"}},
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
					{Section: "CPU", Name: "CPU 结构与指令系统", Tags: []string{"ALU", "寄存器", "PC", "IR", "指令周期", "寻址方式"}},
					{Section: "流水线", Name: "指令流水线与吞吐率", Tags: []string{"IF", "ID", "EX", "WB", "冒险", "吞吐率", "加速比"}},
					{Section: "Cache", Name: "Cache 映射与命中率", Tags: []string{"直接映射", "组相联", "全相联", "命中率", "缺失率"}},
					{Section: "数据表示", Name: "进制转换与码制", Tags: []string{"原码", "反码", "补码", "移码", "浮点数", "规格化"}},
					{Section: "校验码", Name: "校验码原理", Tags: []string{"奇偶校验", "CRC", "海明码"}},
					{Section: "CISC/RISC", Name: "CISC 与 RISC 对比", Tags: []string{"复杂指令集", "精简指令集", "流水线"}},
					{Section: "Flynn 分类", Name: "Flynn 分类法", Tags: []string{"SISD", "SIMD", "MISD", "MIMD"}},
					{Section: "可靠性", Name: "系统可靠性模型", Tags: []string{"串联系统", "并联系统", "模冗余", "MTBF", "MTTR"}},
				},
			},
			{
				Name: "操作系统",
				Points: []SAPoint{
					{Section: "进程线程", Name: "进程与线程", Tags: []string{"PCB", "线程", "上下文切换", "前驱图"}},
					{Section: "同步互斥", Name: "进程同步与互斥", Tags: []string{"信号量", "管程", "PV 操作", "临界区"}},
					{Section: "死锁", Name: "死锁与处理", Tags: []string{"死锁条件", "银行家算法", "预防", "检测", "解除"}},
					{Section: "调度", Name: "处理机调度算法", Tags: []string{"FCFS", "SJF", "时间片轮转", "优先级", "多级反馈队列", "响应比"}},
					{Section: "内存管理", Name: "虚拟内存与页面置换", Tags: []string{"分页", "分段", "段页式", "FIFO", "LRU", "缺页率"}},
					{Section: "文件系统", Name: "文件管理与磁盘调度", Tags: []string{"FCFS", "SSTF", "SCAN", "索引节点", "位示图"}},
				},
			},
			{
				Name: "程序设计语言与语言处理",
				Points: []SAPoint{
					{Section: "语言范式", Name: "程序设计语言分类", Tags: []string{"命令式", "函数式", "逻辑式", "面向对象语言", "动态语言"}},
					{Section: "参数传递", Name: "传值调用与引用调用", Tags: []string{"值传递", "地址传递", "递归"}},
					{Section: "文法", Name: "文法分类与正规式", Tags: []string{"0 型", "1 型", "2 型", "3 型文法", "正则表达式", "有限自动机"}},
					{Section: "编译过程", Name: "编译过程与中间代码", Tags: []string{"词法分析", "语法分析", "语义分析", "三地址码", "逆波兰式"}},
					{Section: "解释与编译", Name: "解释器与编译器对比", Tags: []string{"解释执行", "编译执行", "JIT"}},
				},
			},
			{
				Name: "数据结构与算法",
				Points: []SAPoint{
					{Section: "线性结构", Name: "线性表、栈与队列", Tags: []string{"顺序表", "链表", "栈", "队列", "表达式求值", "KMP"}},
					{Section: "树", Name: "二叉树与特殊树", Tags: []string{"二叉树遍历", "哈夫曼", "AVL", "红黑树", "B 树", "B+ 树"}},
					{Section: "图", Name: "图的存储与遍历", Tags: []string{"邻接矩阵", "邻接表", "DFS", "BFS", "拓扑排序", "关键路径"}},
					{Section: "排序查找", Name: "排序与查找算法", Tags: []string{"冒泡", "快速排序", "归并排序", "堆排序", "折半查找", "散列表"}},
					{Section: "算法策略", Name: "算法设计策略", Tags: []string{"贪心", "动态规划", "分治法", "回溯法", "分支限界", "时间复杂度"}},
				},
			},
			{
				Name: "多媒体基础",
				Points: []SAPoint{
					{Section: "图像基础", Name: "图像存储容量计算", Tags: []string{"分辨率", "像素深度", "色深", "存储容量计算"}},
					{Section: "音视频", Name: "音频采样与视频编码", Tags: []string{"采样频率", "量化位数", "帧率", "MPEG", "H.264"}},
					{Section: "压缩编码", Name: "有损与无损压缩", Tags: []string{"JPEG", "哈夫曼编码", "行程编码", "预测编码"}},
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
					{Section: "IP 地址", Name: "IP 地址与子网划分", Tags: []string{"A 类", "B 类", "C 类", "子网掩码", "CIDR", "IPv6"}},
					{Section: "数据通信", Name: "数据通信基础", Tags: []string{"曼彻斯特编码", "奈奎斯特定理", "香农定理", "多路复用"}},
					{Section: "路由", Name: "路由协议与设备", Tags: []string{"RIP", "OSPF", "BGP", "静态路由", "交换机", "VLAN"}},
					{Section: "局域网", Name: "局域网技术", Tags: []string{"以太网", "CSMA/CD", "令牌环", "拓扑结构"}},
				},
			},
			{
				Name: "应用与传输层协议",
				Points: []SAPoint{
					{Section: "应用层", Name: "常用应用层协议", Tags: []string{"HTTP", "HTTPS", "DNS", "FTP", "SMTP", "POP3", "SNMP", "DHCP"}},
					{Section: "传输层", Name: "TCP 与 UDP 对比", Tags: []string{"三次握手", "四次挥手", "拥塞控制", "滑动窗口"}},
				},
			},
			{
				Name: "分布式系统理论",
				Points: []SAPoint{
					{Section: "CAP", Name: "CAP 定理与 BASE 理论", Tags: []string{"一致性", "可用性", "分区容错", "最终一致"}},
					{Section: "共识算法", Name: "分布式共识算法", Tags: []string{"Paxos", "Raft", "ZooKeeper", "两阶段提交", "三阶段提交"}},
					{Section: "一致性哈希", Name: "一致性哈希", Tags: []string{"哈希环", "虚拟节点", "数据分片"}},
				},
			},
			{
				Name: "中间件与分布式组件",
				Points: []SAPoint{
					{Section: "消息中间件", Name: "消息队列与发布订阅", Tags: []string{"Kafka", "RabbitMQ", "RocketMQ", "发布订阅"}},
					{Section: "缓存", Name: "分布式缓存与数据分片", Tags: []string{"Redis", "缓存穿透", "读写分离", "分库分表"}},
					{Section: "负载均衡", Name: "负载均衡与高可用", Tags: []string{"轮询", "加权轮询", "一致性哈希", "CDN"}},
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
					{Section: "范式", Name: "范式理论 1NF/2NF/3NF/BCNF", Tags: []string{"函数依赖", "码", "范式", "模式分解"}},
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
					{Section: "信息化战略", Name: "企业信息化战略与规划", Tags: []string{"战略", "规划", "数字化转型", "两化融合"}},
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
				Name: "软件过程与开发模型",
				Points: []SAPoint{
					{Section: "开发模型", Name: "软件开发模型", Tags: []string{"瀑布", "增量", "迭代", "原型", "螺旋", "敏捷", "RUP"}},
					{Section: "敏捷", Name: "敏捷开发 Scrum", Tags: []string{"Sprint", "Backlog", "Stand-up", "Retrospective", "极限编程"}},
					{Section: "CMMI", Name: "CMMI 能力成熟度模型", Tags: []string{"初始级", "可重复级", "已定义级", "定量管理", "优化级"}},
				},
			},
			{
				Name: "需求与设计基础",
				Points: []SAPoint{
					{Section: "需求基础", Name: "需求获取与需求分类", Tags: []string{"访谈", "问卷", "原型", "功能需求", "非功能需求"}},
					{Section: "结构化分析", Name: "数据流图与结构化方法", Tags: []string{"DFD", "数据字典", "加工", "外部实体"}},
					{Section: "设计原则", Name: "模块设计与耦合内聚", Tags: []string{"高内聚低耦合", "模块独立性", "信息隐藏"}},
				},
			},
			{
				Name: "面向对象方法",
				Points: []SAPoint{
					{Section: "OO 概念", Name: "封装继承多态", Tags: []string{"封装", "继承", "多态", "重载", "重写"}},
					{Section: "设计原则", Name: "面向对象设计原则", Tags: []string{"SOLID", "开闭原则", "里氏替换", "依赖倒置", "接口隔离"}},
					{Section: "UML 静态", Name: "UML 静态建模图", Tags: []string{"用例图", "类图", "对象图", "构件图", "部署图", "包图"}},
					{Section: "UML 动态", Name: "UML 动态建模图", Tags: []string{"时序图", "顺序图", "通信图", "活动图", "状态图"}},
					{Section: "用例关系", Name: "用例间关系", Tags: []string{"包含 include", "扩展 extend", "泛化"}},
				},
			},
			{
				Name: "设计模式",
				Points: []SAPoint{
					{Section: "创建型", Name: "创建型设计模式", Tags: []string{"单例", "工厂方法", "抽象工厂", "建造者", "原型"}},
					{Section: "结构型", Name: "结构型设计模式", Tags: []string{"适配器", "装饰器", "代理", "外观", "组合", "桥接", "享元"}},
					{Section: "行为型", Name: "行为型设计模式", Tags: []string{"观察者", "策略", "命令", "责任链", "中介者", "备忘录", "迭代器", "状态", "访问者"}},
				},
			},
			{
				Name: "软件测试",
				Points: []SAPoint{
					{Section: "测试分类", Name: "测试阶段与测试方法", Tags: []string{"单元测试", "集成测试", "系统测试", "验收测试", "回归测试", "白盒", "黑盒"}},
					{Section: "测试用例", Name: "测试用例设计方法", Tags: []string{"等价类", "边界值", "因果图", "判定表", "场景法"}},
					{Section: "测试覆盖", Name: "测试覆盖率与质量度量", Tags: []string{"语句覆盖", "分支覆盖", "路径覆盖", "McCabe 环形复杂度"}},
				},
			},
			{
				Name: "维护、配置与质量保证",
				Points: []SAPoint{
					{Section: "维护类型", Name: "软件维护类型", Tags: []string{"改正性", "适应性", "完善性", "预防性"}},
					{Section: "重构", Name: "软件重构与再工程", Tags: []string{"代码重构", "坏味道", "逆向工程", "正向工程"}},
					{Section: "配置管理", Name: "软件配置管理与基线", Tags: []string{"配置项", "基线", "版本控制", "变更控制"}},
					{Section: "质量管理", Name: "质量保证与质量度量", Tags: []string{"QA", "QC", "评审", "审计", "功能点估算", "COCOMO"}},
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
				Name: "网络安全协议",
				Points: []SAPoint{
					{Section: "SSL/TLS", Name: "SSL/TLS 握手过程", Tags: []string{"对称加密", "非对称加密", "证书", "握手"}},
					{Section: "IPSec/VPN", Name: "IPSec 与 VPN", Tags: []string{"隧道", "ESP", "AH", "VPN"}},
					{Section: "认证协议", Name: "身份认证协议", Tags: []string{"Kerberos", "单点登录 SSO", "双因素认证"}},
				},
			},
			{
				Name: "网络安全",
				Points: []SAPoint{
					{Section: "防火墙", Name: "防火墙类型与工作原理", Tags: []string{"包过滤", "状态检测", "代理防火墙", "WAF"}},
					{Section: "IDS/IPS", Name: "入侵检测与入侵防御", Tags: []string{"误用检测", "异常检测", "Snort", "签名"}},
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
	{
		Name: "系统规划与分析",
		Sections: []SASection{
			{
				Name: "系统规划",
				Points: []SAPoint{
					{Section: "总体规划", Name: "信息系统总体规划方法", Tags: []string{"战略目标", "业务规划", "技术规划", "信息化战略"}},
					{Section: "可行性研究", Name: "可行性研究内容", Tags: []string{"技术可行性", "经济可行性", "社会可行性", "运行可行性", "可行性研究报告"}},
				},
			},
			{
				Name: "系统分析",
				Points: []SAPoint{
					{Section: "业务流程分析", Name: "业务流程分析与建模", Tags: []string{"业务流程图", "BPMN", "流程优化", "业务流程重组"}},
					{Section: "数据分析", Name: "数据分析与数据流", Tags: []string{"数据字典", "数据流图 DFD", "数据存储"}},
					{Section: "系统分析报告", Name: "系统分析说明书编写", Tags: []string{"系统目标", "功能需求", "数据需求", "分析报告"}},
				},
			},
		},
	},
	{
		Name: "软件需求工程",
		Sections: []SASection{
			{
				Name: "需求获取与分析",
				Points: []SAPoint{
					{Section: "需求获取", Name: "需求获取技术", Tags: []string{"访谈", "问卷调查", "原型法", "观察法", "焦点小组", "联合需求计划 JRP"}},
					{Section: "需求分类", Name: "需求分类与层次", Tags: []string{"业务需求", "用户需求", "功能需求", "非功能需求"}},
				},
			},
			{
				Name: "需求建模与规格说明",
				Points: []SAPoint{
					{Section: "需求建模", Name: "需求建模方法", Tags: []string{"用例图", "活动图", "状态图", "类图"}},
					{Section: "SRS 编写", Name: "软件需求规格说明书 SRS", Tags: []string{"SRS 模板", "功能描述", "性能指标", "接口需求"}},
				},
			},
			{
				Name: "需求验证与管理",
				Points: []SAPoint{
					{Section: "需求验证", Name: "需求验证方法", Tags: []string{"需求评审", "需求测试", "原型验证", "一致性检查"}},
					{Section: "需求管理", Name: "需求变更与跟踪", Tags: []string{"基线", "需求跟踪矩阵", "变更控制", "版本管理"}},
				},
			},
		},
	},
	{
		Name: "软件架构设计",
		Sections: []SASection{
			{
				Name: "架构基础与视图",
				Points: []SAPoint{
					{Section: "架构概念", Name: "软件架构概念与生命周期", Tags: []string{"架构定义", "架构 4+1 视图", "架构与生命周期"}},
					{Section: "架构描述", Name: "架构描述语言 ADL", Tags: []string{"组件", "连接件", "配置", "约束"}},
				},
			},
			{
				Name: "架构风格",
				Points: []SAPoint{
					{Section: "经典风格", Name: "经典架构风格", Tags: []string{"分层", "管道-过滤器", "客户端-服务器", "MVC"}},
					{Section: "现代风格", Name: "现代架构风格", Tags: []string{"微服务", "事件驱动", "SOA", "微内核", "云原生"}},
				},
			},
			{
				Name: "架构评估",
				Points: []SAPoint{
					{Section: "质量属性", Name: "质量属性与架构权衡", Tags: []string{"性能", "可用性", "可维护性", "安全性", "可扩展性"}},
					{Section: "评估方法", Name: "架构评估方法", Tags: []string{"ATAM", "SAAM", "CBAM", "场景评估"}},
				},
			},
			{
				Name: "特定领域架构",
				Points: []SAPoint{
					{Section: "中间件", Name: "中间件与系统集成", Tags: []string{"消息中间件", "事务中间件", "对象请求代理"}},
					{Section: "J2EE/.NET", Name: "J2EE 与 .NET 平台", Tags: []string{"EJB", "Servlet", "Spring", ".NET Framework"}},
				},
			},
		},
	},
	{
		Name: "系统设计",
		Sections: []SASection{
			{
				Name: "概要设计",
				Points: []SAPoint{
					{Section: "模块设计", Name: "模块结构与功能划分", Tags: []string{"模块化", "功能分解", "模块独立性"}},
					{Section: "接口设计", Name: "接口与数据格式设计", Tags: []string{"外部接口", "内部接口", "用户接口", "数据格式"}},
				},
			},
			{
				Name: "详细设计",
				Points: []SAPoint{
					{Section: "数据结构", Name: "数据结构与数据库设计", Tags: []string{"逻辑结构", "物理结构", "数据规范化"}},
					{Section: "算法设计", Name: "关键算法详细设计", Tags: []string{"程序流程图", "N-S 图", "PAD 图", "PDL 伪代码"}},
				},
			},
			{
				Name: "设计评审与模式",
				Points: []SAPoint{
					{Section: "设计评审", Name: "设计评审内容", Tags: []string{"概要设计评审", "详细设计评审", "设计验证"}},
					{Section: "设计模式应用", Name: "设计模式在系统设计中的应用", Tags: []string{"GoF 模式", "架构模式", "反模式"}},
				},
			},
		},
	},
	{
		Name: "软件实现与测试",
		Sections: []SASection{
			{
				Name: "软件实现",
				Points: []SAPoint{
					{Section: "编码规范", Name: "编码规范与编程风格", Tags: []string{"命名规范", "注释规范", "代码风格", "代码审查"}},
					{Section: "编程技术", Name: "程序设计实现技术", Tags: []string{"结构化编程", "面向对象编程", "函数式编程", "泛型"}},
				},
			},
			{
				Name: "软件测试方法",
				Points: []SAPoint{
					{Section: "测试阶段", Name: "测试阶段与测试策略", Tags: []string{"单元测试", "集成测试", "确认测试", "系统测试", "验收测试"}},
					{Section: "测试技术", Name: "白盒与黑盒测试", Tags: []string{"语句覆盖", "分支覆盖", "路径覆盖", "等价类", "边界值", "判定表"}},
				},
			},
			{
				Name: "测试管理与度量",
				Points: []SAPoint{
					{Section: "测试管理", Name: "测试过程管理", Tags: []string{"测试计划", "测试用例", "缺陷管理", "测试报告"}},
					{Section: "测试度量", Name: "测试覆盖率与质量度量", Tags: []string{"覆盖率", "缺陷密度", "MTTF", "MTBF"}},
				},
			},
			{
				Name: "调试与发布",
				Points: []SAPoint{
					{Section: "调试", Name: "调试技术", Tags: []string{"蛮力法", "回溯法", "原因排除法", "演绎法", "归纳法"}},
					{Section: "发布", Name: "软件发布与持续集成", Tags: []string{"持续集成 CI", "持续交付 CD", "灰度发布", "版本管理"}},
				},
			},
		},
	},
	{
		Name: "系统运行与维护",
		Sections: []SASection{
			{
				Name: "系统运行管理",
				Points: []SAPoint{
					{Section: "运行制度", Name: "系统运行管理制度", Tags: []string{"操作规程", "运行记录", "值班制度", "用户管理"}},
					{Section: "运维管理", Name: "IT 服务与运维管理", Tags: []string{"ITIL", "ITSS", "服务级别协议 SLA", "事件管理"}},
				},
			},
			{
				Name: "系统维护",
				Points: []SAPoint{
					{Section: "维护类型", Name: "软件维护类型", Tags: []string{"改正性维护", "适应性维护", "完善性维护", "预防性维护"}},
					{Section: "维护过程", Name: "维护过程管理", Tags: []string{"维护申请", "维护分析", "维护实施", "维护评审"}},
				},
			},
			{
				Name: "系统评价与改进",
				Points: []SAPoint{
					{Section: "系统评价", Name: "系统评价指标", Tags: []string{"性能指标", "经济效益", "用户满意度", "系统可靠性"}},
					{Section: "系统改进", Name: "系统改进与重构", Tags: []string{"代码重构", "架构演进", "技术债务", "再工程"}},
				},
			},
		},
	},
	{
		Name: "Web 应用系统",
		Sections: []SASection{
			{
				Name: "Web 架构与技术",
				Points: []SAPoint{
					{Section: "Web 架构", Name: "Web 系统分层架构", Tags: []string{"表现层", "业务层", "持久层", "前后端分离"}},
					{Section: "前端技术", Name: "Web 前端开发技术", Tags: []string{"HTML5", "CSS3", "JavaScript", "Vue", "React", "Angular"}},
					{Section: "后端技术", Name: "Web 后端开发技术", Tags: []string{"Spring", "Django", "Node.js", "RESTful API", "GraphQL"}},
				},
			},
			{
				Name: "Web 性能与安全",
				Points: []SAPoint{
					{Section: "性能优化", Name: "Web 性能优化", Tags: []string{"CDN", "缓存", "压缩", "懒加载", "负载均衡"}},
					{Section: "Web 安全", Name: "Web 应用安全", Tags: []string{"SQL 注入", "XSS", "CSRF", "会话劫持", "WAF"}},
				},
			},
			{
				Name: "新兴 Web 技术",
				Points: []SAPoint{
					{Section: "云原生", Name: "云原生 Web 架构", Tags: []string{"容器", "Kubernetes", "Serverless", "Service Mesh"}},
					{Section: "PWA", Name: "渐进式 Web 应用 PWA", Tags: []string{"Service Worker", "离线缓存", "消息推送"}},
				},
			},
		},
	},
	{
		Name: "嵌入式系统",
		Sections: []SASection{
			{
				Name: "嵌入式系统基础",
				Points: []SAPoint{
					{Section: "嵌入式组成", Name: "嵌入式系统组成与特点", Tags: []string{"实时", "固件", "ARM", "DSP", "SoC", "微控制器"}},
					{Section: "RTOS", Name: "实时操作系统特性", Tags: []string{"硬实时", "软实时", "任务调度", "优先级反转", "看门狗"}},
				},
			},
			{
				Name: "嵌入式开发与调试",
				Points: []SAPoint{
					{Section: "开发流程", Name: "嵌入式系统开发流程", Tags: []string{"需求分析", "硬件选型", "驱动开发", "应用开发"}},
					{Section: "调试技术", Name: "嵌入式开发与调试", Tags: []string{"交叉编译", "JTAG", "在线仿真", "Bootloader"}},
				},
			},
			{
				Name: "嵌入式通信与接口",
				Points: []SAPoint{
					{Section: "通信接口", Name: "常用嵌入式通信接口", Tags: []string{"UART", "SPI", "I2C", "CAN", "USB"}},
					{Section: "无线通信", Name: "嵌入式无线通信", Tags: []string{"蓝牙", "Wi-Fi", "ZigBee", "LoRa", "NB-IoT"}},
				},
			},
		},
	},
	{
		Name: "移动应用系统",
		Sections: []SASection{
			{
				Name: "移动平台与开发",
				Points: []SAPoint{
					{Section: "平台架构", Name: "iOS 与 Android 架构", Tags: []string{"Cocoa Touch", "Activity", "Service", "BroadcastReceiver"}},
					{Section: "开发模式", Name: "移动应用开发模式", Tags: []string{"原生开发", "混合开发", "跨平台", "Flutter", "React Native"}},
				},
			},
			{
				Name: "移动网络与性能",
				Points: []SAPoint{
					{Section: "网络通信", Name: "移动网络通信", Tags: []string{"HTTP/HTTPS", "WebSocket", "推送服务", "心跳机制"}},
					{Section: "性能优化", Name: "移动端性能优化", Tags: []string{"启动速度", "内存优化", "电量优化", "卡顿监控"}},
				},
			},
			{
				Name: "移动安全与体验",
				Points: []SAPoint{
					{Section: "移动安全", Name: "移动应用安全", Tags: []string{"代码混淆", "数据加密", "防逆向", "权限控制"}},
					{Section: "用户体验", Name: "移动端 UX 设计原则", Tags: []string{"响应式布局", "Material Design", "人机交互", "无障碍"}},
				},
			},
		},
	},
	{
		Name: "大数据处理系统",
		Sections: []SASection{
			{
				Name: "大数据基础",
				Points: []SAPoint{
					{Section: "大数据特征", Name: "大数据 4V/5V 特征", Tags: []string{"Volume", "Velocity", "Variety", "Value", "Veracity"}},
					{Section: "Hadoop 生态", Name: "Hadoop 生态体系", Tags: []string{"HDFS", "MapReduce", "YARN", "HBase", "Hive"}},
				},
			},
			{
				Name: "大数据处理",
				Points: []SAPoint{
					{Section: "批处理", Name: "大数据批处理", Tags: []string{"MapReduce", "Spark", "Spark RDD", "Spark SQL"}},
					{Section: "流处理", Name: "大数据流处理", Tags: []string{"Storm", "Flink", "Spark Streaming", "Kafka Streams"}},
				},
			},
			{
				Name: "大数据应用",
				Points: []SAPoint{
					{Section: "数据湖", Name: "数据湖与数据中台", Tags: []string{"数据湖", "数据中台", "湖仓一体", "元数据管理"}},
					{Section: "数据挖掘", Name: "数据挖掘算法", Tags: []string{"分类", "聚类", "关联规则", "协同过滤", "机器学习"}},
				},
			},
		},
	},
	{
		Name: "微服务系统",
		Sections: []SASection{
			{
				Name: "微服务基础",
				Points: []SAPoint{
					{Section: "微服务概念", Name: "微服务架构特点", Tags: []string{"服务拆分", "独立部署", "服务治理", "去中心化", "轻量通信"}},
					{Section: "服务拆分", Name: "微服务拆分策略", Tags: []string{"业务能力", "子域", "康威定律", "DDD 领域驱动"}},
				},
			},
			{
				Name: "微服务通信与治理",
				Points: []SAPoint{
					{Section: "服务通信", Name: "微服务通信方式", Tags: []string{"RESTful", "gRPC", "消息队列", "事件总线"}},
					{Section: "服务治理", Name: "服务注册与发现", Tags: []string{"Eureka", "Consul", "Nacos", "ZooKeeper"}},
					{Section: "网关与限流", Name: "API 网关与限流熔断", Tags: []string{"Kong", "Zuul", "Spring Cloud Gateway", "Hystrix", "Sentinel"}},
				},
			},
			{
				Name: "微服务高级特性",
				Points: []SAPoint{
					{Section: "分布式事务", Name: "分布式事务解决方案", Tags: []string{"两阶段提交", "TCC", "Saga", "本地消息表", "最大努力通知"}},
					{Section: "容器化", Name: "容器化与服务网格", Tags: []string{"Docker", "Kubernetes", "Istio", "Service Mesh"}},
				},
			},
		},
	},
	{
		Name: "信息物理系统",
		Sections: []SASection{
			{
				Name: "CPS 基础",
				Points: []SAPoint{
					{Section: "CPS 概念", Name: "信息物理系统 CPS 概念", Tags: []string{"计算", "通信", "控制", "物理实体", "深度融合"}},
					{Section: "CPS 体系", Name: "CPS 体系结构", Tags: []string{"感知层", "网络层", "平台层", "应用层"}},
				},
			},
			{
				Name: "工业互联网与智能制造",
				Points: []SAPoint{
					{Section: "工业互联网", Name: "工业互联网架构", Tags: []string{"边缘计算", "工业云", "工业 PaaS", "工业 APP"}},
					{Section: "智能制造", Name: "智能制造与 CPS 应用", Tags: []string{"智能工厂", "数字孪生", "工业 4.0", "柔性制造"}},
				},
			},
			{
				Name: "CPS 安全与标准",
				Points: []SAPoint{
					{Section: "CPS 安全", Name: "CPS 安全挑战", Tags: []string{"设备认证", "数据加密", "固件安全", "供应链安全"}},
					{Section: "标准与协议", Name: "CPS 标准化与协议", Tags: []string{"OPC UA", "MQTT", "CoAP", "时间敏感网络 TSN"}},
				},
			},
		},
	},
	{
		Name: "系统分析师论文写作要点",
		Sections: []SASection{
			{
				Name: "论文结构与写作",
				Points: []SAPoint{
					{Section: "论文结构", Name: "论文标准结构", Tags: []string{"摘要", "正文", "项目背景", "论点展开", "总结"}},
					{Section: "写作要点", Name: "论文写作要点", Tags: []string{"结合实践", "论点明确", "论据充分", "条理清晰", "字数控制"}},
				},
			},
			{
				Name: "常考主题与素材",
				Points: []SAPoint{
					{Section: "常考主题", Name: "常考论题方向", Tags: []string{"需求工程", "系统架构", "项目管理", "信息安全", "新技术应用"}},
					{Section: "素材组织", Name: "论据与素材组织", Tags: []string{"项目背景", "技术难点", "解决方案", "效果评价", "经验总结"}},
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
