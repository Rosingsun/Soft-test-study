package model

import "time"

type QuestionMark struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"column:user_id;not null;uniqueIndex:uk_user_question"`
	QuestionID uint      `gorm:"column:question_id;not null;uniqueIndex:uk_user_question"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (QuestionMark) TableName() string {
	return "question_marks"
}
