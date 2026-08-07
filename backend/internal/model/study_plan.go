package model

import "time"

// StudyPlan 学习计划
type StudyPlan struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"column:user_id;not null;index"`
	SubjectID uint      `gorm:"column:subject_id;not null;default:0"`
	Title     string    `gorm:"column:title;type:varchar(100);not null"`
	DailyGoal int       `gorm:"column:daily_goal;not null;default:20"`
	StartDate time.Time `gorm:"column:start_date;type:date;not null"`
	EndDate   time.Time `gorm:"column:end_date;type:date;not null"`
	Status    int       `gorm:"column:status;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (StudyPlan) TableName() string {
	return "study_plans"
}
