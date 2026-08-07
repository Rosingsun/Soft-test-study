package model

import "time"

// CheckIn 每日打卡汇总
// 注：accuracy/correct_count/duration 为「当日最新一次」数据（重打会更新），
// rank_* 为「首次打卡」快照（排名与个人平均正确率口径，重打保持不变）。
type CheckIn struct {
	ID           uint      `gorm:"primarykey"`
	UserID       uint      `gorm:"column:user_id;not null;uniqueIndex:uk_user_date"`
	CheckDate    time.Time `gorm:"column:check_date;type:date;not null;uniqueIndex:uk_user_date"`
	TotalCount   int       `gorm:"column:total_count;not null;default:10"`
	CorrectCount int       `gorm:"column:correct_count;not null;default:0"`
	Accuracy     float64   `gorm:"column:accuracy;not null;default:0"`
	Duration     int       `gorm:"column:duration;not null;default:0"`
	Status       int       `gorm:"column:status;not null;default:1"`
	// RankCorrectCount 首次打卡快照：答对数（重打保持不变）
	RankCorrectCount int `gorm:"column:rank_correct_count;not null;default:0"`
	// RankAccuracy 首次打卡快照：正确率（重打保持不变，排名/个人平均口径）
	RankAccuracy float64 `gorm:"column:rank_accuracy;not null;default:0"`
	// RankDuration 首次打卡快照：耗时（重打保持不变）
	RankDuration int       `gorm:"column:rank_duration;not null;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (CheckIn) TableName() string {
	return "check_ins"
}
