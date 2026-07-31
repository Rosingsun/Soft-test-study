package model

type QuestionKnowledge struct {
	QuestionID       uint `gorm:"column:question_id;primaryKey"`
	KnowledgePointID uint `gorm:"column:knowledge_point_id;primaryKey"`
}

func (QuestionKnowledge) TableName() string {
	return "question_knowledge"
}
