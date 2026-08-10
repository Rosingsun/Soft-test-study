package model

import "time"

type AiGeneratedQuestion struct {
	ID             uint      `gorm:"primarykey"`
	UserID         uint      `gorm:"column:user_id;not null;index"`
	QuestionID     uint      `gorm:"column:question_id;default:0;index"`
	SubjectID      uint      `gorm:"column:subject_id;not null;index"`
	ChapterID      uint      `gorm:"column:chapter_id;default:0;index"`
	Type           string    `gorm:"column:type;type:varchar(20);not null"`
	Difficulty     string    `gorm:"column:difficulty;type:varchar(20);not null"`
	Content        string    `gorm:"column:content;type:text;not null"`
	CaseMaterial   string    `gorm:"column:case_material;type:text"`
	Options        string    `gorm:"column:options;type:json"`
	Answer         string    `gorm:"column:answer;type:varchar(500);not null"`
	Analysis       string    `gorm:"column:analysis;type:text"`
	KnowledgePoint string    `gorm:"column:knowledge_point;type:varchar(200);default:''"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (AiGeneratedQuestion) TableName() string {
	return "ai_generated_questions"
}
