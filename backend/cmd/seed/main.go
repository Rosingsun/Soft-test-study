package main

import (
	"fmt"
	"log"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	seedLevels(db)
	seedSubjects(db)
	seedSubSubjects(db)
	seedChapters(db)
	assignQuestionsToChapters(db)
	seedExamTemplates(db)

	fmt.Println("种子数据初始化完成")
}

func seedLevels(db *gorm.DB) {
	levels := []model.ExamLevel{
		{Name: "初级", SortOrder: 1},
		{Name: "中级", SortOrder: 2},
		{Name: "高级", SortOrder: 3},
	}
	for _, l := range levels {
		db.FirstOrCreate(&l, model.ExamLevel{Name: l.Name})
	}
}

func seedSubjects(db *gorm.DB) {
	subjects := []model.Subject{
		{LevelID: 1, Name: "程序员", ShortName: "程序员", Description: "计算机基础与程序设计", SortOrder: 1},
		{LevelID: 1, Name: "网络管理员", ShortName: "网管", Description: "网络基础与管理", SortOrder: 2},
		{LevelID: 1, Name: "信息处理技术员", ShortName: "信处", Description: "信息处理技术", SortOrder: 3},
		{LevelID: 2, Name: "软件设计师", ShortName: "软设", Description: "软件工程与设计", SortOrder: 1},
		{LevelID: 2, Name: "网络工程师", ShortName: "网工", Description: "网络工程与技术", SortOrder: 2},
		{LevelID: 2, Name: "数据库系统工程师", ShortName: "库工", Description: "数据库系统设计与管理", SortOrder: 3},
		{LevelID: 2, Name: "信息系统管理工程师", ShortName: "信管", Description: "信息系统管理", SortOrder: 4},
		{LevelID: 3, Name: "信息系统项目管理师", ShortName: "高项", Description: "信息系统项目管理", SortOrder: 1},
		{LevelID: 3, Name: "系统分析师", ShortName: "系分", Description: "系统分析与设计", SortOrder: 2},
		{LevelID: 3, Name: "系统架构设计师", ShortName: "架构", Description: "系统架构设计", SortOrder: 3},
		{LevelID: 3, Name: "网络规划设计师", ShortName: "网规", Description: "网络规划与设计", SortOrder: 4},
	}
	for _, s := range subjects {
		db.FirstOrCreate(&s, model.Subject{Name: s.Name, LevelID: s.LevelID})
	}
}

func seedSubSubjects(db *gorm.DB) {
	// 高级科目：综合知识 / 案例分析 / 论文
	for _, high := range []string{"信息系统项目管理师", "系统分析师", "系统架构设计师", "网络规划设计师"} {
		createSubSubjects(db, high, []string{"综合知识", "案例分析", "论文"})
	}
	// 初/中级科目：基础知识 / 应用技术
	for _, mid := range []string{"程序员", "网络管理员", "信息处理技术员", "软件设计师", "网络工程师", "数据库系统工程师", "信息系统管理工程师"} {
		createSubSubjects(db, mid, []string{"基础知识", "应用技术"})
	}
}

func createSubSubjects(db *gorm.DB, subjectName string, names []string) {
	var subject model.Subject
	if err := db.Where("name = ?", subjectName).First(&subject).Error; err != nil {
		log.Printf("科目 %s 不存在，跳过子科目", subjectName)
		return
	}
	for i, name := range names {
		db.FirstOrCreate(&model.SubSubject{
			SubjectID: subject.ID,
			Name:      name,
			SortOrder: i + 1,
		}, model.SubSubject{SubjectID: subject.ID, Name: name})
	}
}

// chapterNamesBySubject 各科目「基础知识/综合知识」的章节规划
var chapterNamesBySubject = map[string][]string{
	"程序员": {
		"计算机基础", "操作系统", "程序设计基础", "数据结构与算法", "数据库基础", "网络与信息安全",
	},
	"软件设计师": {
		"计算机组成与体系结构", "操作系统", "数据结构与算法", "程序设计语言",
		"软件工程基础", "数据库系统", "计算机网络", "信息安全",
	},
	"网络工程师": {
		"计算机网络概述", "数据通信基础", "局域网与以太网", "网络互连与路由", "网络安全", "网络管理与维护",
	},
	"数据库系统工程师": {
		"数据库系统概论", "关系数据模型", "SQL语言", "数据库设计", "事务管理与并发控制", "数据库管理与安全",
	},
	"信息系统项目管理师": {
		"项目管理基础", "项目立项与整体管理", "范围与时间管理", "成本与质量管理",
		"人力资源与沟通管理", "风险与采购管理", "配置与变更管理", "信息系统综合知识", "信息技术知识",
	},
	"网络管理员": {"网络基础", "网络设备与配置", "网络安全与管理"},
	"信息处理技术员": {"信息处理基础", "办公软件应用", "信息安全与法规"},
	"信息系统管理工程师": {"信息系统基础", "系统运维管理", "IT服务管理"},
	// 系统分析师综合知识按《系统分析师教程（第2版）》第一篇「基础知识」9章分类
	"系统分析师": {
		"绪论", "数学与工程基础", "计算机系统", "计算机网络与分布式系统",
		"数据库系统", "企业信息化", "软件工程", "项目管理", "信息安全",
	},
	"系统架构设计师": {"架构设计基础", "软件架构风格", "系统质量与性能", "云与大数据架构"},
	"网络规划设计师": {"网络规划基础", "网络拓扑设计", "网络性能与安全"},
}

func seedChapters(db *gorm.DB) {
	var subSubjects []model.SubSubject
	db.Order("id asc").Find(&subSubjects)

	for _, ss := range subSubjects {
		var subject model.Subject
		if err := db.First(&subject, ss.SubjectID).Error; err != nil {
			continue
		}

		// 系统分析师-综合知识按第二版教材9章重建：先强制删除该子科目下旧题目与旧章节，避免残留旧分类
		if subject.Name == "系统分析师" && ss.Name == "综合知识" {
			db.Unscoped().Where("subject_id = ? AND sub_subject_id = ?", subject.ID, ss.ID).
				Delete(&model.Question{})
			db.Unscoped().Where("subject_id = ? AND sub_subject_id = ?", subject.ID, ss.ID).
				Delete(&model.Chapter{})
		}

		names := chapterNames(subject.Name, ss.Name)
		for i, name := range names {
			db.FirstOrCreate(&model.Chapter{
				SubjectID:    subject.ID,
				SubSubjectID: ss.ID,
				Name:         name,
				SortOrder:    i + 1,
			}, model.Chapter{SubjectID: subject.ID, SubSubjectID: ss.ID, Name: name})
		}
	}
}

func chapterNames(subjectName, subSubjectName string) []string {
	switch subSubjectName {
	case "综合知识", "基础知识":
		if names, ok := chapterNamesBySubject[subjectName]; ok {
			return names
		}
		return []string{"基础理论（一）", "基础理论（二）", "基础理论（三）"}
	case "案例分析", "应用技术":
		return []string{"综合案例（一）", "综合案例（二）", "综合案例（三）"}
	case "论文":
		return []string{"论文写作专题"}
	default:
		return []string{"学习章节"}
	}
}

// assignQuestionsToChapters 将 chapter_id=0 的存量题目按子科目分配到章节
func assignQuestionsToChapters(db *gorm.DB) {
	var questions []model.Question
	db.Where("chapter_id = 0 AND status = 1").Order("subject_id asc, sub_subject_id asc, id asc").Find(&questions)

	// 系统分析师「综合知识」由 import_real 按第二版教材9章精确归类，seed 不随机分配
	skipSubject, skipSub := resolveSAZongheIDs(db)

	// 按 (subject_id, sub_subject_id) 缓存的章节列表
	chapterPool := map[string][]model.Chapter{}
	poolIndex := map[string]int{}

	// 记录实际更新的数量
	updated := 0
	for i := range questions {
		q := &questions[i]
		// 跳过由 import_real 专属处理的系统分析师综合知识题目
		if skipSubject != 0 && q.SubjectID == skipSubject && q.SubSubjectID == skipSub {
			continue
		}
		key := fmt.Sprintf("%d:%d", q.SubjectID, q.SubSubjectID)

		chapters, ok := chapterPool[key]
		if !ok {
			// 优先使用该子科目下的章节
			if q.SubSubjectID != 0 {
				db.Where("sub_subject_id = ? AND subject_id = ?", q.SubSubjectID, q.SubjectID).
					Order("id asc").Find(&chapters)
			}
			// 无子科目(sub_subject_id=0)的孤立题目：只归入该科目的「综合知识/基础知识」主章节，
			// 避免被随机分派到「论文/案例分析」等非客观题目章节，污染专项练习
			if len(chapters) == 0 {
				var primary model.SubSubject
				if err := db.Where("subject_id = ?", q.SubjectID).
					Order("sort_order asc, id asc").First(&primary).Error; err == nil {
					db.Where("sub_subject_id = ? AND subject_id = ?", primary.ID, q.SubjectID).
						Order("id asc").Find(&chapters)
				}
			}
			chapterPool[key] = chapters
			poolIndex[key] = 0
		}
		if len(chapters) == 0 {
			continue
		}

		idx := poolIndex[key] % len(chapters)
		poolIndex[key] = idx + 1

		if err := db.Model(q).Update("chapter_id", chapters[idx].ID).Error; err != nil {
			log.Printf("分配章节失败 question_id=%d: %v", q.ID, err)
			continue
		}
		updated++
	}
	if updated > 0 {
		fmt.Printf("已为 %d 道题目分配章节\n", updated)
	}
}

// resolveSAZongheIDs 返回系统分析师「综合知识」的 subject_id / sub_subject_id，
// 用于让 seed 跳过该子科目，避免随机分配章节覆盖 import_real 的精确归类
func resolveSAZongheIDs(db *gorm.DB) (uint, uint) {
	var subject model.Subject
	if err := db.Where("name = ?", "系统分析师").First(&subject).Error; err != nil {
		return 0, 0
	}
	var ss model.SubSubject
	if err := db.Where("subject_id = ? AND name = ?", subject.ID, "综合知识").First(&ss).Error; err != nil {
		return 0, 0
	}
	return subject.ID, ss.ID
}

func seedExamTemplates(db *gorm.DB) {
	// 为有题目的科目生成模拟试卷
	var subjects []model.Subject
	db.Find(&subjects)

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
			log.Printf("创建试卷模板失败: %v", err)
			continue
		}

		totalScore := 0
		for i, q := range questions {
			score := questionScore(q.Type)
			totalScore += score
			db.Create(&model.ExamTemplateQuestion{
				TemplateID: template.ID,
				QuestionID: q.ID,
				SortOrder:  i + 1,
				Score:      score,
			})
		}
		db.Model(&template).Update("total_score", totalScore)
		fmt.Printf("已创建试卷: %s（%d 题，%d 分）\n", templateName, len(questions), totalScore)
	}

	seedEssayTemplates(db)
}

// seedEssayTemplates 为含论文题的科目生成「论文写作卷」（幂等）。
func seedEssayTemplates(db *gorm.DB) {
	var subjects []model.Subject
	db.Find(&subjects)

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
			log.Printf("创建论文卷失败: %v", err)
			continue
		}
		for i, q := range essayQuestions {
			db.Create(&model.ExamTemplateQuestion{
				TemplateID: template.ID,
				QuestionID: q.ID,
				SortOrder:  i + 1,
				Score:      1,
			})
		}
		fmt.Printf("已创建论文卷: %s（%d 题）\n", templateName, len(essayQuestions))
	}
}

func questionScore(qtype string) int {
	switch qtype {
	case "multi":
		return 2
	case "short", "comprehensive", "case_study":
		return 5
	case "single", "judge", "fill":
		return 1
	default:
		return 1
	}
}
