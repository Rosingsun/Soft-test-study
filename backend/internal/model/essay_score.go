package model

import "time"

// EssayScore 论文 AI 评分记录
type EssayScore struct {
	ID               uint      `gorm:"primarykey"`
	UserID           uint      `gorm:"column:user_id;not null;index:idx_user_record"`
	QuestionID       uint      `gorm:"column:question_id;not null"`
	RecordType       string    `gorm:"column:record_type;type:varchar(20);not null;comment:practice/exam"`
	RecordID         uint      `gorm:"column:record_id;not null;index:idx_user_record;comment:练习记录ID或考试记录ID"`
	ExamAnswerID     uint      `gorm:"column:exam_answer_id;comment:考试答题详情ID(仅考试模式)"`
	UserAnswer       string    `gorm:"column:user_answer;type:text"`
	TotalScore       int       `gorm:"column:total_score;comment:AI 总分(满分75)"`
	ArgumentScore    int       `gorm:"column:argument_score;comment:论点与立意得分"`
	StructureScore   int       `gorm:"column:structure_score;comment:结构与逻辑得分"`
	LanguageScore    int       `gorm:"column:language_score;comment:语言表达得分"`
	DepthScore       int       `gorm:"column:depth_score;comment:深度与广度得分"`
	Comment          string    `gorm:"column:comment;type:text;comment:AI 评语"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (EssayScore) TableName() string {
	return "essay_scores"
}
