package model

import "time"

type WrongQuestion struct {
	ID           uint      `gorm:"primarykey"`
	UserID       uint      `gorm:"column:user_id;not null;uniqueIndex:uk_user_question"`
	QuestionID   uint      `gorm:"column:question_id;not null;uniqueIndex:uk_user_question"`
	WrongCount   int       `gorm:"column:wrong_count;default:1"`
	CorrectCount int       `gorm:"column:correct_count;default:0"`
	LastWrongAt  time.Time `gorm:"column:last_wrong_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (WrongQuestion) TableName() string {
	return "wrong_questions"
}
