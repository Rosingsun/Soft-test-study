package model

import "time"

type User struct {
	ID             uint       `gorm:"primarykey"`
	Username       string     `gorm:"column:username;type:varchar(50);uniqueIndex;not null"`
	Email          string     `gorm:"column:email;type:varchar(100);uniqueIndex;not null"`
	EmailVerified  bool       `gorm:"column:email_verified;type:tinyint(1);default:0"`
	PasswordHash   string     `gorm:"column:password_hash;type:varchar(255);not null"`
	Nickname       string     `gorm:"column:nickname;type:varchar(50)"`
	Avatar         string     `gorm:"column:avatar;type:varchar(255)"`
	Role           string     `gorm:"column:role;type:varchar(20);default:student"`
	Status         int        `gorm:"column:status;default:1"`
	LevelID        uint       `gorm:"column:level_id;default:0"`
	SubjectID      uint       `gorm:"column:subject_id;default:0"`
	Difficulty     string     `gorm:"column:difficulty;type:varchar(20);default:''"`
	FailedAttempts int        `gorm:"column:failed_attempts;default:0"`
	LockedUntil    *time.Time `gorm:"column:locked_until"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
