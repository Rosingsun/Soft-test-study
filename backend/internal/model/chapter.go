package model

import "time"

type Chapter struct {
	ID           uint      `gorm:"primarykey"`
	SubjectID    uint      `gorm:"column:subject_id;not null;index"`
	SubSubjectID uint      `gorm:"column:sub_subject_id;default:0;index"`
	ParentID     uint      `gorm:"column:parent_id;default:0"`
	Name         string    `gorm:"column:name;type:varchar(200);not null"`
	SortOrder    int       `gorm:"column:sort_order;default:0"`
	Weight       float64   `gorm:"column:weight;type:decimal(5,2);default:1.00"`
	MaterialID   uint      `gorm:"column:material_id;default:0;index"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (Chapter) TableName() string {
	return "chapters"
}
