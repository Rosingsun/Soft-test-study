package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type ChapterRepo struct {
	db *gorm.DB
}

func NewChapterRepo(db *gorm.DB) *ChapterRepo {
	return &ChapterRepo{db: db}
}

func (r *ChapterRepo) FindBySubSubjectID(subSubjectID uint) ([]model.Chapter, error) {
	var chapters []model.Chapter
	err := r.db.Where("sub_subject_id = ?", subSubjectID).
		Order("sort_order asc").
		Find(&chapters).Error
	return chapters, err
}

func (r *ChapterRepo) FindBySubjectID(subjectID uint) ([]model.Chapter, error) {
	var chapters []model.Chapter
	err := r.db.Where("subject_id = ?", subjectID).
		Order("sort_order asc").
		Find(&chapters).Error
	return chapters, err
}
