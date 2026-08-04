package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

// rawQuestion 对应爬虫输出 data/case_study.json 中的记录。
// 其中 subject_id / sub_subject_id / chapter_id 为可选提示字段，
// 实际导入时按名称（subject / subSubject）解析为数据库真实 id。
type rawQuestion struct {
	Subject      string `json:"subject"`
	SubSubject   string `json:"sub_subject"`
	Year         int    `json:"year"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	CaseMaterial string `json:"case_material"`
	Options      map[string]string `json:"options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Source       string `json:"source"`
}

func main() {
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	jsonPath := resolveDataPath("data/case_study.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("读取 JSON 失败(%s): %v", jsonPath, err)
	}
	var recs []rawQuestion
	if err := json.Unmarshal(data, &recs); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
	}
	log.Printf("读取到 %d 条案例分析记录", len(recs))

	inserted, skipped := 0, 0
	for _, r := range recs {
		if r.Content == "" {
			skipped++
			continue
		}
		subjectName := r.Subject
		if subjectName == "" {
			subjectName = "系统分析师"
		}
		subSubjectName := r.SubSubject
		if subSubjectName == "" {
			subSubjectName = "案例分析"
		}

		subject, subSubject, err := resolveTargets(db, subjectName, subSubjectName)
		if err != nil {
			log.Printf("跳过: %v（content=%s）", err, truncate(r.Content, 30))
			skipped++
			continue
		}

		// 按题干去重（同一科目下不重复插入相同内容）
		var cnt int64
		db.Model(&model.Question{}).
			Where("subject_id = ? AND content = ?", subject.ID, r.Content).
			Count(&cnt)
		if cnt > 0 {
			skipped++
			continue
		}

		q := model.Question{
			SubjectID:    subject.ID,
			SubSubjectID: subSubject.ID,
			Type:         model.TypeCaseStudy,
			Difficulty:   r.Difficulty,
			Content:      r.Content,
			CaseMaterial: r.CaseMaterial,
			Options:      buildOptions(r.Options),
			Answer:       r.Answer,
			Analysis:     r.Analysis,
			Year:         r.Year,
			Source:       r.Source,
			Status:       1,
		}
		if q.Difficulty == "" {
			q.Difficulty = "medium"
		}
		if err := db.Create(&q).Error; err != nil {
			log.Printf("插入失败 content=%s: %v", truncate(r.Content, 40), err)
			skipped++
			continue
		}
		inserted++
	}
	fmt.Printf("导入完成：新增 %d 题，跳过 %d 题\n", inserted, skipped)

	// 导入后把 chapter_id=0 的题分配到章节（复用 seed 逻辑）
	assignChapters(db)
}

// assignChapters 将新插入的 chapter_id=0 案例分析题分配到该子科目下的章节
func assignChapters(db *gorm.DB) {
	var questions []model.Question
	db.Where("chapter_id = 0 AND type = ? AND status = 1", model.TypeCaseStudy).
		Order("id asc").Find(&questions)
	if len(questions) == 0 {
		return
	}
	poolIndex := 0
	updated := 0
	for i := range questions {
		q := &questions[i]
		var chapters []model.Chapter
		db.Where("sub_subject_id = ? AND subject_id = ?", q.SubSubjectID, q.SubjectID).
			Order("id asc").Find(&chapters)
		if len(chapters) == 0 {
			db.Where("subject_id = ?", q.SubjectID).Order("id asc").Find(&chapters)
		}
		if len(chapters) == 0 {
			continue
		}
		idx := poolIndex % len(chapters)
		poolIndex++
		if err := db.Model(q).Update("chapter_id", chapters[idx].ID).Error; err != nil {
			log.Printf("分配章节失败 question_id=%d: %v", q.ID, err)
			continue
		}
		updated++
	}
	if updated > 0 {
		fmt.Printf("已为 %d 道案例分析题分配章节\n", updated)
	}
}

func resolveTargets(db *gorm.DB, subjectName, subSubjectName string) (model.Subject, model.SubSubject, error) {
	var subject model.Subject
	if err := db.Where("name = ?", subjectName).First(&subject).Error; err != nil {
		return model.Subject{}, model.SubSubject{}, fmt.Errorf("未找到科目「%s」", subjectName)
	}
	var ss model.SubSubject
	if err := db.Where("subject_id = ? AND name = ?", subject.ID, subSubjectName).First(&ss).Error; err != nil {
		return model.Subject{}, model.SubSubject{}, fmt.Errorf("未找到子科目「%s」", subSubjectName)
	}
	return subject, ss, nil
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

func buildOptions(opts map[string]string) string {
	// 案例分析题通常无选项；直接存空 JSON 数组
	b, _ := json.Marshal(opts)
	return string(b)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
