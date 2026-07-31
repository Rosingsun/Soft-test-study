package model

import "time"

type ExamRecord struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"column:user_id;not null;index"`
	TemplateID uint      `gorm:"column:template_id;not null;index"`
	Score      int       `gorm:"column:score"`
	TotalScore int       `gorm:"column:total_score"`
	Duration   int       `gorm:"column:duration"`
	Status     string    `gorm:"column:status;type:varchar(20);default:pending"`
	StartedAt  time.Time `gorm:"column:started_at"`
	FinishedAt time.Time `gorm:"column:finished_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (ExamRecord) TableName() string {
	return "exam_records"
}

type ExamRecordAnswer struct {
	ID         uint   `gorm:"primarykey"`
	RecordID   uint   `gorm:"column:record_id;not null;index"`
	QuestionID uint   `gorm:"column:question_id;not null;index"`
	Answer     string `gorm:"column:answer;type:text"`
	IsCorrect  int    `gorm:"column:is_correct"`
	Score      int    `gorm:"column:score"`
}

func (ExamRecordAnswer) TableName() string {
	return "exam_record_answers"
}
