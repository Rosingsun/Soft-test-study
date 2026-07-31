package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type SubjectRepo struct {
	db *gorm.DB
}

func NewSubjectRepo(db *gorm.DB) *SubjectRepo {
	return &SubjectRepo{db: db}
}

func (r *SubjectRepo) FindByLevelID(levelID uint) ([]model.Subject, error) {
	var subjects []model.Subject
	err := r.db.Where("level_id = ? AND status = 1", levelID).
		Order("sort_order asc").
		Find(&subjects).Error
	return subjects, err
}

func (r *SubjectRepo) FindByID(id uint) (*model.Subject, error) {
	var subject model.Subject
	err := r.db.First(&subject, id).Error
	return &subject, err
}

func (r *SubjectRepo) FindAll() ([]model.Subject, error) {
	var subjects []model.Subject
	err := r.db.Where("status = 1").Order("sort_order asc").Find(&subjects).Error
	return subjects, err
}
