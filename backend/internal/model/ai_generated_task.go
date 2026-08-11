package model

import "time"

// AiGeneratedTask AI 异步出题任务状态（持久化版）。
// 历史上由 service.aiTaskStore 内存 map 维护，重启即丢；
// 改为 DB 持久化后，内存 store 降级为热缓存（参见 service/ai_task.go）。
// 复合索引 (user_id, status, updated_at) 与 (status, updated_at) 由
// 迁移文件 backend/migrations/000007_ai_generated_tasks.up.sql 创建。
type AiGeneratedTask struct {
	ID              uint       `gorm:"primarykey"`
	TaskID          string     `gorm:"column:task_id;size:64;not null;uniqueIndex:uk_task_id"`
	UserID          uint       `gorm:"column:user_id;not null;index:idx_user_status_updated,priority:1"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:'pending';index:idx_user_status_updated,priority:2;index:idx_status_updated,priority:1"`
	QuestionType    string     `gorm:"column:question_type;type:varchar(20);not null"`
	ChapterID       uint       `gorm:"column:chapter_id;default:0"`
	ChapterName     string     `gorm:"column:chapter_name;type:varchar(200);default:''"`
	Difficulty      string     `gorm:"column:difficulty;type:varchar(20);default:''"`
	Count           int        `gorm:"column:count;default:0"`
	Error           string     `gorm:"column:error;type:text"`
	ResultQuestions string     `gorm:"column:result_questions;type:json"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;index:idx_status_updated,priority:2"`
	FinishedAt      *time.Time `gorm:"column:finished_at"`
}

func (AiGeneratedTask) TableName() string {
	return "ai_generated_tasks"
}
