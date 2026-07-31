package model

import "time"

type PracticeRecord struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"column:user_id;not null;index"`
	QuestionID uint      `gorm:"column:question_id;not null;index"`
	Mode       string    `gorm:"column:mode;type:varchar(20);not null"`
	Answer     string    `gorm:"column:answer;type:text"`
	IsCorrect  int       `gorm:"column:is_correct"`
	Duration   int       `gorm:"column:duration"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (PracticeRecord) TableName() string {
	return "practice_records"
}
