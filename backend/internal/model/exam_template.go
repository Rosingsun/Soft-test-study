package model

import "time"

type ExamTemplate struct {
	ID            uint      `gorm:"primarykey"`
	SubjectID     uint      `gorm:"column:subject_id;not null;index"`
	Name          string    `gorm:"column:name;type:varchar(200);not null"`
	Duration      int       `gorm:"column:duration;not null"`
	TotalScore    int       `gorm:"column:total_score;not null"`
	QuestionType  string    `gorm:"column:question_type;type:varchar(20);default:''"`
	IsPublic      int       `gorm:"column:is_public;default:1"`
	Year          int       `gorm:"column:year"`
	Status        int       `gorm:"column:status;default:1"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (ExamTemplate) TableName() string {
	return "exam_templates"
}

type ExamTemplateQuestion struct {
	TemplateID uint `gorm:"column:template_id;primaryKey"`
	QuestionID uint `gorm:"column:question_id;primaryKey"`
	SortOrder  int  `gorm:"column:sort_order;default:0"`
	Score      int  `gorm:"column:score;not null"`
}

func (ExamTemplateQuestion) TableName() string {
	return "exam_template_questions"
}
