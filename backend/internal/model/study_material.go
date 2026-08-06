package model

import "time"

// StudyMaterial 学习资料（xmind 笔记分享）
type StudyMaterial struct {
	ID             uint      `gorm:"primarykey"`
	Title          string    `gorm:"column:title;type:varchar(200);not null"`
	Description    string    `gorm:"column:description;type:text"`
	SubjectID      uint      `gorm:"column:subject_id;index"`
	CoverURL       string    `gorm:"column:cover_url;type:varchar(255)"`
	FileURL        string    `gorm:"column:file_url;type:varchar(255);not null"`
	FileName       string    `gorm:"column:file_name;type:varchar(255)"`
	FileSize       int64     `gorm:"column:file_size;default:0"`
	DownloadCount  int       `gorm:"column:download_count;default:0"`
	ViewCount      int       `gorm:"column:view_count;default:0"`
	SortOrder      int       `gorm:"column:sort_order;default:0"`
	Status         int       `gorm:"column:status;default:1"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

func (StudyMaterial) TableName() string {
	return "study_materials"
}
