package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type EssayScoreRepo struct {
	db *gorm.DB
}

func NewEssayScoreRepo(db *gorm.DB) *EssayScoreRepo {
	return &EssayScoreRepo{db: db}
}

// Create 创建论文评分记录
func (r *EssayScoreRepo) Create(score *model.EssayScore) error {
	return r.db.Create(score).Error
}

// Save 更新论文评分记录
func (r *EssayScoreRepo) Save(score *model.EssayScore) error {
	return r.db.Save(score).Error
}

// FindByRecord 根据记录类型和ID查找评分（练习模式用 record_id）
func (r *EssayScoreRepo) FindByRecord(recordType string, recordID uint) (*model.EssayScore, error) {
	var score model.EssayScore
	err := r.db.Where("record_type = ? AND record_id = ?", recordType, recordID).
		Order("created_at desc").First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// FindByExamAnswer 根据考试答题详情ID查找评分（考试模式用 exam_answer_id）
func (r *EssayScoreRepo) FindByExamAnswer(examAnswerID uint) (*model.EssayScore, error) {
	var score model.EssayScore
	err := r.db.Where("exam_answer_id = ?", examAnswerID).
		Order("created_at desc").First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// FindByUserAndQuestion 根据用户和题目查找评分
func (r *EssayScoreRepo) FindByUserAndQuestion(userID, questionID uint) (*model.EssayScore, error) {
	var score model.EssayScore
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Order("created_at desc").First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}
