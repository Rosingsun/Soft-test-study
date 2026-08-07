package model

import "time"

// CheckInQuestion 每日打卡题目快照
type CheckInQuestion struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"column:user_id;not null;index:idx_user_date"`
	CheckDate  time.Time `gorm:"column:check_date;type:date;not null;index:idx_user_date"`
	QuestionID uint      `gorm:"column:question_id;not null"`
	Answer     string    `gorm:"column:answer;type:text"`
	IsCorrect  int       `gorm:"column:is_correct;not null;default:0"`
	Duration   int       `gorm:"column:duration;not null;default:0"`
}

func (CheckInQuestion) TableName() string {
	return "check_in_questions"
}
