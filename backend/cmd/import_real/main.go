package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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

	inserted, skipped := 0, 0
	for _, r := range recs {
		if r.Content == "" || r.Answer == "" {
			skipped++
			continue
		}
		// 按题干去重
		var cnt int64
		db.Model(&model.Question{}).
			Where("subject_id = ? AND content = ?", subject.ID, r.Content).
			Count(&cnt)
		if cnt > 0 {
			skipped++
			continue
		}

		optionsJSON := buildOptions(r.Options)
		q := model.Question{
			SubjectID:    subject.ID,
			SubSubjectID: subSubject.ID,
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
