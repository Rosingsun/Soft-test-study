package main

import (
	"fmt"
	"log"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

// 为所有有题目的科目生成模拟试卷（幂等），使用批量插入减少往返。
func main() {
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	var subjects []model.Subject
	if err := db.Order("id asc").Find(&subjects).Error; err != nil {
		log.Fatalf("查询科目失败: %v", err)
	}

	for _, subject := range subjects {
		var questions []model.Question
		db.Where("subject_id = ? AND status = 1", subject.ID).Order("RAND()").Limit(20).Find(&questions)
		if len(questions) < 10 {
			continue
		}

		templateName := fmt.Sprintf("%s模拟试卷（一）", subject.Name)
		var existing model.ExamTemplate
		err := db.Where("subject_id = ? AND name = ?", subject.ID, templateName).First(&existing).Error
		if err == nil {
			fmt.Printf("已存在: %s (id=%d)，跳过\n", templateName, existing.ID)
			continue
		}

		template := model.ExamTemplate{
			SubjectID:  subject.ID,
			Name:       templateName,
			Duration:   90,
			TotalScore: 0,
			IsPublic:   1,
			Year:       2024,
			Status:     1,
		}
		if err := db.Create(&template).Error; err != nil {
			log.Printf("创建试卷模板失败 %s: %v", templateName, err)
			continue
		}

		totalScore := 0
		etqs := make([]model.ExamTemplateQuestion, 0, len(questions))
		for i, q := range questions {
			score := questionScore(q.Type)
			totalScore += score
			etqs = append(etqs, model.ExamTemplateQuestion{
				TemplateID: template.ID,
				QuestionID: q.ID,
				SortOrder:  i + 1,
				Score:      score,
			})
		}
		if err := db.Create(&etqs).Error; err != nil {
			log.Printf("写入试卷题目失败 %s: %v", templateName, err)
			continue
		}
		db.Model(&template).Update("total_score", totalScore)
		fmt.Printf("已创建试卷: %s（%d 题，%d 分）\n", templateName, len(questions), totalScore)
	}

	createEssayTemplates(db)
	fmt.Println("试卷生成完成")
}

// createEssayTemplates 为含论文题的科目生成「论文写作卷」（幂等）。
func createEssayTemplates(db *gorm.DB) {
	var subjects []model.Subject
	if err := db.Order("id asc").Find(&subjects).Error; err != nil {
		log.Printf("查询科目失败: %v", err)
		return
	}

	for _, subject := range subjects {
		var essayQuestions []model.Question
		db.Where("subject_id = ? AND type = ? AND status = 1", subject.ID, model.TypeEssay).
			Order("year asc").Find(&essayQuestions)
		if len(essayQuestions) == 0 {
			continue
		}

		templateName := fmt.Sprintf("%s论文写作卷", subject.Name)
		var existing model.ExamTemplate
		err := db.Where("subject_id = ? AND name = ?", subject.ID, templateName).First(&existing).Error
		if err == nil {
			fmt.Printf("已存在: %s (id=%d)，跳过\n", templateName, existing.ID)
			continue
		}

		template := model.ExamTemplate{
			SubjectID:    subject.ID,
			Name:         templateName,
			Duration:     150,
			TotalScore:   1,
			QuestionType: model.TypeEssay,
			IsPublic:     1,
			Year:         2024,
			Status:       1,
		}
		if err := db.Create(&template).Error; err != nil {
			log.Printf("创建论文卷失败 %s: %v", templateName, err)
			continue
		}

		etqs := make([]model.ExamTemplateQuestion, 0, len(essayQuestions))
		for i, q := range essayQuestions {
			etqs = append(etqs, model.ExamTemplateQuestion{
				TemplateID: template.ID,
				QuestionID: q.ID,
				SortOrder:  i + 1,
				Score:      1,
			})
		}
		if err := db.Create(&etqs).Error; err != nil {
			log.Printf("写入论文卷题目失败 %s: %v", templateName, err)
			continue
		}
		fmt.Printf("已创建论文卷: %s（%d 题）\n", templateName, len(essayQuestions))
	}
}

func questionScore(qtype string) int {
	switch qtype {
	case "multi":
		return 2
	case "short", "comprehensive":
		return 5
	case "single", "judge", "fill":
		return 1
	default:
		return 1
	}
}
