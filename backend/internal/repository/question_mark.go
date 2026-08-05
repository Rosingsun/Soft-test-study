package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type QuestionMarkRepo struct {
	db *gorm.DB
}

func NewQuestionMarkRepo(db *gorm.DB) *QuestionMarkRepo {
	return &QuestionMarkRepo{db: db}
}

func (r *QuestionMarkRepo) Add(mark *model.QuestionMark) error {
	return r.db.Create(mark).Error
}

func (r *QuestionMarkRepo) Remove(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.QuestionMark{}).Error
}

func (r *QuestionMarkRepo) FindByUser(userID uint) ([]model.QuestionMark, error) {
	var list []model.QuestionMark
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *QuestionMarkRepo) IsMarked(userID, questionID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.QuestionMark{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count).Error
	return count > 0, err
}
