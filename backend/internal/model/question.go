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
