package model

import "time"

// ReviewCard SM-2 遗忘曲线复习卡片
type ReviewCard struct {
	ID             uint       `gorm:"primarykey"`
	UserID         uint       `gorm:"column:user_id;not null;uniqueIndex:uk_user_question"`
	QuestionID     uint       `gorm:"column:question_id;not null;uniqueIndex:uk_user_question"`
	Repetition     int        `gorm:"column:repetition;not null;default:0"`
	IntervalDays   int        `gorm:"column:interval_days;not null;default:1"`
	EaseFactor     float64    `gorm:"column:ease_factor;type:decimal(4,2);not null;default:2.50"`
	DueDate        time.Time  `gorm:"column:due_date;type:date;not null"`
	LastReviewedAt *time.Time `gorm:"column:last_reviewed_at"`
	Status         int        `gorm:"column:status;not null;default:1"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (ReviewCard) TableName() string {
	return "review_cards"
}
