package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type ExamLevelRepo struct {
	db *gorm.DB
}

func NewExamLevelRepo(db *gorm.DB) *ExamLevelRepo {
	return &ExamLevelRepo{db: db}
}

func (r *ExamLevelRepo) FindAll() ([]model.ExamLevel, error) {
	var levels []model.ExamLevel
	err := r.db.Order("sort_order asc").Find(&levels).Error
	return levels, err
}

func (r *ExamLevelRepo) FindByID(id uint) (*model.ExamLevel, error) {
	var level model.ExamLevel
	err := r.db.First(&level, id).Error
	return &level, err
}
