package model

import "time"

type FavoriteFolder struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"column:user_id;not null;index"`
	Name      string    `gorm:"column:name;type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (FavoriteFolder) TableName() string {
	return "favorite_folders"
}

type QuestionFavorite struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"column:user_id;not null;uniqueIndex:uk_user_question"`
	QuestionID uint      `gorm:"column:question_id;not null;uniqueIndex:uk_user_question"`
	FolderID   uint      `gorm:"column:folder_id;default:0;index"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (QuestionFavorite) TableName() string {
	return "question_favorites"
}
