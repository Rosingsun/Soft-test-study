package model

import "time"

// 题型常量
const (
	TypeSingle       = "single"
	TypeMulti        = "multi"
	TypeJudge        = "judge"
	TypeFill         = "fill"
	TypeShort        = "short"
	TypeComprehensive = "comprehensive"
	TypeEssay        = "essay"
	TypeCaseStudy    = "case_study"
	// TypeMultiBlank 多空题：题干含多个填空位，每个空为独立单选；整题一次提交、按空独立判分
	TypeMultiBlank = "multi_blank"
)

type Question struct {
	ID           uint      `gorm:"primarykey"`
	SubjectID    uint      `gorm:"column:subject_id;not null;index"`
	SubSubjectID uint      `gorm:"column:sub_subject_id;default:0;index"`
	ChapterID    uint      `gorm:"column:chapter_id;default:0;index"`
	Type         string    `gorm:"column:type;type:varchar(20);not null"`
	Difficulty   string    `gorm:"column:difficulty;type:varchar(20);not null"`
	Content      string    `gorm:"column:content;type:text;not null"`
	CaseMaterial string    `gorm:"column:case_material;type:text"`
	Options      string    `gorm:"column:options;type:json"`
	// BlankOptions 多空题专用：每空的独立选项 JSON 数组
	// 格式：[{"blank_index":1,"options":[{"id":"A","content":"..."},...]}, ...]
	// 仅当 Type=TypeMultiBlank 时使用；其他题型为空
	BlankOptions string    `gorm:"column:blank_options;type:json"`
	// Answer 多空题时存 JSON 数组字符串，如 "[\"A\",\"C\"]"；其他题型存原始答案
	Answer       string    `gorm:"column:answer;type:text;not null"`
	Analysis     string    `gorm:"column:analysis;type:text"`
	Year         int       `gorm:"column:year"`
	Source       string    `gorm:"column:source;type:varchar(100);default:''"`
	Status       int       `gorm:"column:status;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Question) TableName() string {
	return "questions"
}
