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
	log.Printf("读取到 %d 条题目记录", len(recs))

	subject, subSubject, err := resolveTargets(db)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("导入目标: 科目 %s(id=%d) / 子科目 %s(id=%d)",
		subject.Name, subject.ID, subSubject.Name, subSubject.ID)

	// 加载第二版教材综合知识9章（仅限该子科目下），建立章节名 -> 章节ID 的映射
	chapterIDByName, err := loadChapterMap(db, subject.ID, subSubject.ID)
	if err != nil {
		log.Fatalf("加载章节失败: %v", err)
	}

	// 题目要求：旧数据不再需要，先清空该子科目下所有题目再重新生成
	if err := clearOldQuestions(db, subject.ID, subSubject.ID); err != nil {
		log.Fatalf("清空旧题目失败: %v", err)
	}
	log.Printf("已清空「%s - %s」下旧题目，开始按第二版教材章节重新归类导入",
		subject.Name, subSubject.Name)

	inserted, skipped := 0, 0
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
			Answer:       r.Answer,
			Analysis:     r.Analysis,
			Year:         r.Year,
			Source:       buildSource(r),
			Status:       1,
		}
		if err := db.Create(&q).Error; err != nil {
			log.Printf("插入失败 content=%s: %v", truncate(r.Content, 40), err)
			skipped++
			continue
		}
		inserted++
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

// chapterKeywords 定义《系统分析师教程（第2版）》综合知识9章的关键词，
// 用于把题目自动归类到对应章节。顺序即匹配优先级（越靠前越优先）。
var chapterKeywords = []struct {
	Name     string
	Keywords []string
}{
	{"绪论", []string{"绪论", "系统分析师", "系统分析", "角色", "职业道德", "法律法规", "标准化", "知识产权", "标准"}},
	{"数学与工程基础", []string{"图论", "运筹", "线性规划", "排队论", "博弈论", "组合数学", "概率论", "数理统计", "矩阵", "离散", "数值", "误差", "最优化", "数学建模", "微积分", "集合", "逻辑代数", "布尔"}},
	{"计算机系统", []string{"CPU", "处理器", "指令系统", "冯·诺依曼", "存储系统", "Cache", "内存", "寄存器", "总线", "体系结构", "RISC", "CISC", "流水线", "并行处理", "多核", "嵌入式", "固件", "芯片", "操作系统", "进程", "线程", "死锁", "虚拟内存", "文件系统", "编译", "汇编"}},
	{"计算机网络与分布式系统", []string{"网络", "TCP", "IP", "HTTP", "DNS", "路由", "交换机", "OSI", "协议", "子网", "网关", "分布式", "微服务", "集群", "负载均衡", "CDN", "区块链", "P2P", "通信", "带宽", "加密传输"}},
	{"数据库系统", []string{"数据库", "关系模型", "SQL", "ER", "范式", "事务", "并发控制", "索引", "数据仓库", "数据挖掘", "OLAP", "OLTP", "NoSQL", "分布式数据库", "视图", "触发器", "游标"}},
	{"企业信息化", []string{"企业信息化", "ERP", "CRM", "SCM", "电子商务", "电子政务", "智能制造", "工业互联网", "数字化转型", "业务流程", "BPR", "知识管理", "决策支持", "DSS", "BI", "大数据", "数据治理"}},
	{"软件工程", []string{"软件工程", "需求", "用例", "UML", "面向对象", "设计模式", "软件测试", "单元测试", "集成测试", "维护", "重构", "敏捷", "瀑布", "CMMI", "质量", "软件度量", "配置管理", "版本", "建模", "状态图", "活动图", "类图"}},
	{"项目管理", []string{"项目管理", "进度", "关键路径", "CPM", "PERT", "甘特图", "成本", "估算", "挣值", " Risk", "风险", "WBS", "范围", "里程碑", "资源", "招投标", "合同", "质量保证"}},
	{"信息安全", []string{"信息安全", "加密", "对称", "非对称", "RSA", "DES", "AES", "哈希", "数字签名", "认证", "防火墙", "入侵", "病毒", "木马", "安全", "PKI", "证书", "访问控", "权限", "审计", "隐私"}},
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

// classifyChapter 根据题目的题干、选项与解析组合文本，匹配最合适的章节
func classifyChapter(r rawQuestion, chapterIDByName map[string]uint) uint {
	text := r.Content
	for k, v := range r.Options {
		text += " " + k + " " + v
	}
	text += " " + r.Analysis

	for _, ch := range chapterKeywords {
		for _, kw := range ch.Keywords {
			if strings.Contains(text, kw) {
				if id, ok := chapterIDByName[ch.Name]; ok {
					return id
				}
				break
			}
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
