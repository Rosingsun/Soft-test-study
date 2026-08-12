package model

import "time"

type KnowledgePoint struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"column:name;type:varchar(200);not null;uniqueIndex:uk_kp_name"`
	SubjectID   uint      `gorm:"column:subject_id;default:0;index"`
	Description string    `gorm:"column:description;type:text"`
	Source      string    `gorm:"column:source;type:varchar(20);not null;default:'ai'"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (KnowledgePoint) TableName() string {
	return "knowledge_points"
}

type UserKnowledgePoint struct {
	ID               uint       `gorm:"primarykey"`
	UserID           uint       `gorm:"column:user_id;not null;uniqueIndex:uk_user_kp"`
	KnowledgePointID uint       `gorm:"column:knowledge_point_id;not null;uniqueIndex:uk_user_kp"`
	SourceQuestionID uint       `gorm:"column:source_question_id;default:0;index"`
	IsImportant      bool       `gorm:"column:is_important;type:tinyint(1);not null;default:0"`
	IsMastered       bool       `gorm:"column:is_mastered;type:tinyint(1);not null;default:0"`
	MasteredAt       *time.Time `gorm:"column:mastered_at"`
	Note             string     `gorm:"column:note;type:text"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (UserKnowledgePoint) TableName() string {
	return "user_knowledge_points"
}
