package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type AiRepo struct {
	db *gorm.DB
}

func NewAiRepo(db *gorm.DB) *AiRepo {
	return &AiRepo{db: db}
}

func (r *AiRepo) CreateQuestions(questions []model.AiGeneratedQuestion) error {
	return r.db.Create(&questions).Error
}

func (r *AiRepo) FindByUserID(userID uint, limit int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *AiRepo) FindRandomByUser(userID uint, count int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("RAND()").Limit(count).Find(&list).Error
	return list, err
}

// FindRandomByUserAndSubject 按用户+科目随机获取 AI 生成的题目，仅返回已关联 questions 表的记录
func (r *AiRepo) FindRandomByUserAndSubject(userID, subjectID uint, count int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ? AND subject_id = ? AND question_id > 0", userID, subjectID).
		Order("RAND()").Limit(count).Find(&list).Error
	return list, err
}

// CountByUserAndSubject 统计用户在某科目下已关联题库的 AI 题目数量
func (r *AiRepo) CountByUserAndSubject(userID, subjectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.AiGeneratedQuestion{}).
		Where("user_id = ? AND subject_id = ? AND question_id > 0", userID, subjectID).Count(&count).Error
	return count, err
}

func (r *AiRepo) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.AiGeneratedQuestion{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
