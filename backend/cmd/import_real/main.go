package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

// rawQuestion 对应 data/sa_questions_standard.json 中的记录
type rawQuestion struct {
	Subject  string            `json:"subject"`
	Year     int               `json:"year"`
	Half     string            `json:"half"`
	Paper    string            `json:"paper"`
	Type     string            `json:"type"`
	Content  string            `json:"content"`
	Options  map[string]string `json:"options"`
	Answer   string            `json:"answer"`
	Analysis string            `json:"analysis"`
	Source   string            `json:"source"`
}

func main() {
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	jsonPath := resolveDataPath("data/sa_questions_standard.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("读取 JSON 失败(%s): %v", jsonPath, err)
	}
	var recs []rawQuestion
	if err := json.Unmarshal(data, &recs); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
	}
	mainCount := len(recs)
	log.Printf("读取到 %d 条主数据源题目记录", mainCount)

	// 合并补充数据源（考友回忆版真题），按归一化题干去重后追加
	recs = mergeExtraRecords(recs, resolveDataPath("data/sa_questions_lightsoft.json"))
	log.Printf("合并补充数据源后共 %d 条记录（新增 %d 条）", len(recs), len(recs)-mainCount)

	subject, subSubject, err := resolveTargets(db)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("导入目标: 科目 %s(id=%d) / 子科目 %s(id=%d)",
		subject.Name, subject.ID, subSubject.Name, subSubject.ID)

	// 加载考纲 18 章（仅限该子科目下），建立章节名 -> 章节ID 的映射
	chapterIDByName, err := loadChapterMap(db, subject.ID, subSubject.ID)
	if err != nil {
		log.Fatalf("加载章节失败: %v", err)
	}

	// 题目要求：旧数据不再需要，先清空该子科目下所有题目再重新生成
	if err := clearOldQuestions(db, subject.ID, subSubject.ID); err != nil {
		log.Fatalf("清空旧题目失败: %v", err)
	}
	log.Printf("已清空「%s - %s」下旧题目，开始按考纲 18 章重新归类导入",
		subject.Name, subSubject.Name)

	inserted, skipped := 0, 0

	// 先构造全部待插入题目：blank_options 为空的普通题型统一填合法 JSON "[]"，
	// 否则 MySQL JSON 列会拒绝空字符串（Error 3140 Invalid JSON text）
	type pending struct {
		rec rawQuestion
		q   model.Question
	}
	pendings := make([]pending, 0, len(recs))
	for _, r := range recs {
		if r.Content == "" || r.Answer == "" {
			skipped++
			continue
		}

		chapterID := classifyChapter(r, chapterIDByName)

		optionsJSON := buildOptions(r.Options)
		q := model.Question{
			SubjectID:    subject.ID,
			SubSubjectID: subSubject.ID,
			ChapterID:    chapterID,
			Type:         normalizeType(r.Type),
			Difficulty:   "medium",
			Content:      r.Content,
			Options:      optionsJSON,
			BlankOptions: "[]",
			Answer:       r.Answer,
			Analysis:     r.Analysis,
			Year:         r.Year,
			Source:       buildSource(r),
			Status:       1,
		}
		pendings = append(pendings, pending{rec: r, q: q})
	}

	// 分批事务提交：远程数据库逐条插入往返延迟高，批量事务可将耗时降低一个数量级；
	// 批内任一失败则整批回退为逐条插入，以便精准跳过问题记录并继续导入其余数据
	const batchSize = 50
	for start := 0; start < len(pendings); start += batchSize {
		end := start + batchSize
		if end > len(pendings) {
			end = len(pendings)
		}
		batch := pendings[start:end]

		err := db.Transaction(func(tx *gorm.DB) error {
			for i := range batch {
				if err := tx.Create(&batch[i].q).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err == nil {
			inserted += len(batch)
			continue
		}
		for i := range batch {
			if err := db.Create(&batch[i].q).Error; err != nil {
				log.Printf("插入失败 content=%s: %v", truncate(batch[i].rec.Content, 40), err)
				skipped++
				continue
			}
			inserted++
		}
	}
	fmt.Printf("导入完成：新增 %d 题，跳过 %d 题\n", inserted, skipped)
}

func resolveDataPath(rel string) string {
	for _, base := range []string{"", "..", "../.."} {
		candidate := filepath.Join(base, rel)
		if _, err := os.Stat(candidate); err == nil {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}
	return rel
}

// mergeExtraRecords 合并补充数据源中的题目记录：
// 按归一化题干（去除空白字符）判断重复，仅追加主数据源中没有的题目。
// 补充文件不存在时静默跳过，不影响主流程。
func mergeExtraRecords(recs []rawQuestion, extraPath string) []rawQuestion {
	data, err := os.ReadFile(extraPath)
	if err != nil {
		log.Printf("未找到补充数据源(%s)，跳过合并: %v", extraPath, err)
		return recs
	}
	var extra []rawQuestion
	if err := json.Unmarshal(data, &extra); err != nil {
		log.Printf("解析补充数据源失败(%s): %v", extraPath, err)
		return recs
	}

	seen := make(map[string]struct{}, len(recs))
	for _, r := range recs {
		seen[normalizeContent(r.Content)] = struct{}{}
	}

	added := 0
	for _, r := range extra {
		if r.Content == "" || r.Answer == "" {
			continue
		}
		key := normalizeContent(r.Content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		recs = append(recs, r)
		added++
	}
	log.Printf("补充数据源 %d 条记录，去重后新增 %d 条", len(extra), added)
	return recs
}

// normalizeContent 归一化题干文本作为去重键：
// 去除「考友回忆版」标记前缀与所有 Unicode 空白字符。
func normalizeContent(s string) string {
	s = strings.ReplaceAll(s, "【考友回忆版】", "")
	var b strings.Builder
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r', '\v', '\f', 0x00A0, 0x3000:
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func resolveTargets(db *gorm.DB) (model.Subject, model.SubSubject, error) {
	var subject model.Subject
	if err := db.Where("name = ?", "系统分析师").First(&subject).Error; err != nil {
		return model.Subject{}, model.SubSubject{}, fmt.Errorf("未找到科目「系统分析师」")
	}
	var ss model.SubSubject
	if err := db.Where("subject_id = ? AND name = ?", subject.ID, "综合知识").First(&ss).Error; err != nil {
		return model.Subject{}, model.SubSubject{}, fmt.Errorf("未找到子科目「综合知识」")
	}
	return subject, ss, nil
}

// chapterKeywords 定义系统分析师综合知识 18 章的关键词，
// 用于把题目自动归类到对应章节。
// 排列规则：特异性强的知识域在前（如法律法规、多媒体、嵌入式、信息安全），
// 泛化的知识域在后（如软件工程、项目管理），避免关键词误吸。
// 同组内顺序即匹配优先级（越靠前越优先）。
var chapterKeywords = []struct {
	Name     string
	Keywords []string
}{
	{"绪论", []string{"系统分析师", "生命周期", "软件危机", "职业道德", "角色与职责"}},
	{"法律法规与标准化", []string{"著作权", "专利权", "专利法", "商标", "招投标法", "招标投标", "政府采购法", "合同法", "网络安全法", "标准化法", "知识产权", "国家标准 GB", "GB/T", "强制性标准", "推荐性标准", "侵权", "软件保护条例"}},
	{"多媒体基础", []string{"多媒体", "图像存储容量", "采样频率", "量化位数", "色深", "像素", "MPEG", "JPEG", "H.26", "音视频编码", "色彩空间", "帧率", "有损压缩", "无损压缩"}},
	{"嵌入式系统", []string{"嵌入式", "实时操作系统 RTOS", "RTOS", "片上系统 SoC", "SoC", "交叉编译", "固件", "微控制器", "单片机", "JTAG", "看门狗"}},
	{"信息安全", []string{"信息安全", "加密", "对称密钥", "非对称密钥", "RSA", "DES", "AES", "ECC", "哈希", "摘要算法", "MD5", "SHA", "数字签名", "数字证书", "PKI", "CA 认证", "防火墙", "入侵检测", "IDS", "IPS", "病毒", "木马", "蠕虫", "拒绝服务", "DDoS", "SQL 注入", "XSS", "访问控制", "权限管理", "安全审计", "等级保护", "灾备", "SSL", "TLS", "IPSec", "VPN", "Kerberos", "认证技术", "生物特征", "公钥", "私钥", "报文摘要"}},
	{"数据结构与算法", []string{"哈夫曼", "邻接矩阵", "邻接表", "拓扑排序", "关键路径 AOE", "AOE 网", "时间复杂度", "空间复杂度", "二叉树", "平衡树", "红黑树", "线索二叉树", "顺序栈", "链栈", "共享栈", "循环队列", "双端队列", "优先队列", "队头", "队尾", "链表", "顺序表", "散列表", "冲突处理", "KMP", "模式串", "贪心策略", "动态规划", "分治法", "回溯法", "分支限界", "快速排序", "归并排序", "堆排序", "基数排序", "希尔排序", "直接插入排序", "折半查找", "顺序查找", "B 树", "B+ 树"}},
	{"运筹学基础", []string{"线性规划", "单纯形", "对偶问题", "整数规划", "0-1 规划", "运输问题", "指派问题", "最大流", "最小费用流", "排队论", "泊松流", "M/M/1", "博弈论", "纳什均衡", "决策树", "盈亏平衡", "期望收益", "最小生成树", "Prim", "Kruskal", "最短路径 Dijkstra", "Dijkstra 算法", "Floyd 算法"}},
	{"数学基础", []string{"离散数学", "集合论", "关系矩阵", "命题逻辑", "谓词逻辑", "推理规则", "等价关系", "偏序", "群论", "环论", "概率论", "数学期望", "方差", "正态分布", "二项分布", "排列组合", "鸽巢原理", "容斥原理", "抽屉原理", "误差", "插值", "数值积分", "迭代法", "牛顿迭代", "矩阵运算", "行列式"}},
	{"程序设计语言与语言处理", []string{"文法", "正规式", "正则表达式", "词法分析", "语法分析", "语义分析", "中间代码", "目标代码", "编译器", "解释器", "传值调用", "引用调用", "地址传递", "递归调用", "闭包", "作用域", "动态绑定", "静态绑定", "多态的实现", "标识符", "表达式求值", "后缀表达式", "逆波兰式"}},
	{"面向对象方法与设计模式", []string{"UML", "用例图", "类图", "对象图", "时序图", "顺序图", "通信图", "状态图", "活动图", "构件图", "部署图", "包图", "设计模式", "单例模式", "工厂方法", "抽象工厂", "观察者模式", "策略模式", "适配器模式", "装饰器模式", "代理模式", "外观模式", "组合模式", "桥接模式", "享元模式", "命令模式", "责任链", "中介者", "备忘录", "迭代器", "状态模式", "访问者", "封装", "继承", "多态", "SOLID", "开闭原则", "里氏替换", "依赖倒置", "接口隔离"}},
	{"操作系统", []string{"操作系统", "进程状态", "进程调度", "进程同步", "进程互斥", "进程与线程", "进程 P", "子进程", "父进程", "线程", "协程", "死锁", "银行家算法", "信号量", "PV 操作", "PV操作", "pv操作", "P、V 操作", "管程", "临界资源", "临界区", "页面置换", "LRU", "FIFO 置换", "缺页中断", "虚拟存储器", "分页存储", "分段存储", "段页式", "磁盘调度", "SCAN", "SSTF", "索引节点", "位示图", "文件物理结构", "Spooling", "作业调度", "响应比高者优先", "时间片轮转", "多级反馈队列", "前趋图", "处理机系统", "单处理机"}},
	{"计算机组成与体系结构", []string{"CPU", "运算器", "控制器", "寄存器", "指令周期", "指令流水线", "流水线吞吐率", "加速比", "Cache 命中率", "高速缓存", "主存", "辅存", "寻址方式", "立即寻址", "间接寻址", "总线仲裁", "冯·诺依曼", "哈佛结构", "RISC", "CISC", "Flynn 分类", "SISD", "SIMD", "MISD", "MIMD", "多核", "并行处理", "海明码", "CRC 校验", "原码", "反码", "补码", "移码", "浮点数表示", "规格化", "I/O 控制方式", "中断", "DMA", "通道", "可靠性 MTBF", "串联系统", "并联系统", "模冗余", "RAID", "冗余阵列", "磁盘阵列", "磁盘存取时间", "数据传输率", "字长", "机器字长"}},
	{"数据库系统", []string{"数据库", "关系代数", "选择投影连接", "元组", "候选键", "候选码", "主键", "外键", "函数依赖", "范式", "1NF", "2NF", "3NF", "BCNF", "4NF", "模式分解", "无损连接", "事务的 ACID", "原子性", "隔离级别", "封锁协议", "两段锁", "共享锁", "排他锁", "视图", "触发器", "存储过程", "游标", "数据仓库", "星型模型", "雪花模型", "OLAP", "OLTP", "数据挖掘", "NoSQL", "反规范化", "分布式数据库", "两阶段提交协议", "并发控制", "关系模式", "E-R 图", "ER 图", "实体联系"}},
	{"分布式系统与中间件", []string{"分布式系统", "CAP 定理", "BASE 理论", "最终一致性", "Paxos", "Raft", "ZooKeeper", "一致性哈希", "数据分片", "微服务", "服务治理", "服务网关", "注册中心", "消息队列", "发布订阅", "Kafka", "RabbitMQ", "负载均衡", "CDN 内容分发", "缓存穿透", "Redis 集群", "分布式事务", "TCC 补偿", "RPC 远程调用", "SOA 面向服务", "ESB 企业服务总线", "中间件", "容器编排", "Kubernetes", "区块链共识"}},
	{"计算机网络", []string{"计算机网络", "OSI 参考模型", "TCP/IP 协议簇", "三次握手", "四次挥手", "滑动窗口", "拥塞控制", "IP 地址", "IP地址", "子网划分", "子网掩码", "网络地址", "CIDR", "无类域间路由", "IPv6", "ARP", "ICMP", "DHCP", "DNS 域名解析", "HTTP", "HTTPS", "FTP", "SMTP", "POP3", "SNMP 网络管理", "路由协议", "RIP", "OSPF", "BGP", "交换机", "VLAN", "以太网", "CSMA/CD", "令牌环", "网桥", "集线器", "拓扑结构", "星型", "环型", "总线型", "曼彻斯特编码", "差分曼彻斯特", "奈奎斯特定理", "香农定理", "波特率", "比特率", "多路复用", "时分复用", "频分复用", "ATM 异步传输", "路由器", "数据包", "IP 报文", "IP报文", "报文首部", "网段", "QoS", "IntServ", "DiffServ", "RSVP", "资源预留", "无线接入", "Wi-Fi", "5G 网络"}},
	{"企业信息化", []string{"企业信息化", "信息化战略", "数字化转型", "ERP", "企业资源计划", "CRM 客户关系", "客户关系管理", "SCM 供应链", "供应链管理", "PLM 产品生命周期", "电子商务", "B2B", "B2C", "C2C", "O2O", "电子政务", "G2G", "G2B", "G2C", "智能制造", "工业互联网", "两化融合", "业务流程重组", "业务流程再造", "BPR", "BPMN", "决策支持系统", "DSS", "商业智能", "BI 可视化", "知识管理", "显性知识", "隐性知识", "数据治理", "主数据", "元数据管理", "IT 服务", "ITIL", "ITSS", "服务级别协议", "SLA", "供应商关系管理", "企业应用集成", "EAI", "商业过程建模"}},
	{"项目管理", []string{"项目管理", "挣值分析", "PV EV AC", "成本偏差 CV", "进度偏差 SV", "CPI 成本绩效", "SPI 进度绩效", "WBS 工作分解", "工作包", "范围基准", "里程碑", "甘特图", "PERT 三点估算", "计划评审技术", "CPM 关键路径法", "赶工", "资源平衡", "风险识别", "风险应对策略", "定性风险分析", "定量风险分析", "干系人管理", "沟通渠道数", "团队建设", "冲突解决", "合同类型", "总价合同", "成本补偿合同", "工料合同", "索赔", "工序", "紧前作业", "紧前工序", "总工期", "总时差", "自由时差", "最早开始时间", "最迟开始时间", "项目计划", "项目监控"}},
	{"软件工程", []string{"软件工程", "软件开发模型", "瀑布模型", "原型模型", "增量模型", "螺旋模型", "喷泉模型", "敏捷开发", "Scrum", "极限编程 XP", "统一过程 RUP", "需求工程", "需求获取", "需求分析", "需求验证", "需求跟踪矩阵", "软件设计原则", "高内聚低耦合", "模块独立性", "结构化设计", "结构化方法", "结构化分析", "数据流图", "DFD", "数据字典", "白盒测试", "黑盒测试", "等价类划分", "边界值分析", "判定表", "因果图", "单元测试", "集成测试", "确认测试", "系统测试", "回归测试", "语句覆盖", "判定覆盖", "条件覆盖", "路径覆盖", "McCabe", "环形复杂度", "改正性维护", "适应性维护", "完善性维护", "预防性维护", "重构", "配置管理", "基线", "版本控制", "质量保证 QA", "质量控制 QC", "CMMI", "成熟度等级", "软件度量", "功能点估算", "COCCOMO 模型", "COCOMO 估算", "逆向工程", "软件再工程", "软件工程三要素", "软件工具", "开发方法", "面向数据流", "工作流参考模型", "服务建模", "面向服务架构", "系统分析阶段", "系统设计阶段", "Requirements validation", "systems analysis"}},
}

// fallbackChapter 当题目无法命中任何关键词时的兜底章节
const fallbackChapter = "软件工程"

// loadChapterMap 读取该科目+子科目下的章节，返回 章节名 -> 章节ID
func loadChapterMap(db *gorm.DB, subjectID, subSubjectID uint) (map[string]uint, error) {
	var chapters []model.Chapter
	if err := db.Where("subject_id = ? AND sub_subject_id = ?", subjectID, subSubjectID).
		Find(&chapters).Error; err != nil {
		return nil, err
	}
	m := make(map[string]uint, len(chapters))
	for _, c := range chapters {
		m[c.Name] = c.ID
	}
	return m, nil
}

// clearOldQuestions 清空指定子科目下的全部旧题目（题目要求旧数据不再需要）
func clearOldQuestions(db *gorm.DB, subjectID, subSubjectID uint) error {
	return db.Where("subject_id = ? AND sub_subject_id = ?", subjectID, subSubjectID).
		Delete(&model.Question{}).Error
}

// classifyChapter 根据题目的题干、选项与解析组合文本进行章节归类。
// 采用评分制：统计每章命中的关键词数量，取最高分章节；
// 相比"先到先得"，可避免泛化关键词把题目提前截走、大量堆入兜底章。
func classifyChapter(r rawQuestion, chapterIDByName map[string]uint) uint {
	text := r.Content
	for k, v := range r.Options {
		text += " " + k + " " + v
	}
	text += " " + r.Analysis

	bestName, bestScore := "", 0
	for _, ch := range chapterKeywords {
		score := 0
		for _, kw := range ch.Keywords {
			if strings.Contains(text, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestName = ch.Name
		}
	}
	if bestScore > 0 {
		if id, ok := chapterIDByName[bestName]; ok {
			return id
		}
	}
	// 兜底：归入默认章节
	if id, ok := chapterIDByName[fallbackChapter]; ok {
		return id
	}
	// 极端情况：返回任意章节，避免 chapter_id 为 0
	for _, id := range chapterIDByName {
		return id
	}
	return 0
}

func buildOptions(opts map[string]string) string {
	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	arr := make([]string, 0, len(keys))
	for _, k := range keys {
		arr = append(arr, k+". "+opts[k])
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

func normalizeType(t string) string {
	switch t {
	case "single":
		return "single"
	case "multi":
		return "multi"
	default:
		return "single"
	}
}

func buildSource(r rawQuestion) string {
	if r.Source != "" {
		return fmt.Sprintf("%s | %s", r.Source, r.Paper)
	}
	return r.Paper
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
