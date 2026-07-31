package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type SubSubjectRepo struct {
	db *gorm.DB
}

func NewSubSubjectRepo(db *gorm.DB) *SubSubjectRepo {
	return &SubSubjectRepo{db: db}
}

func (r *SubSubjectRepo) FindBySubjectID(subjectID uint) ([]model.SubSubject, error) {
	var list []model.SubSubject
	err := r.db.Where("subject_id = ?", subjectID).
		Order("sort_order asc").
		Find(&list).Error
	return list, err
}

func (r *SubSubjectRepo) FindByID(id uint) (*model.SubSubject, error) {
	var item model.SubSubject
	err := r.db.First(&item, id).Error
	return &item, err
}
