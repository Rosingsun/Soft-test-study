package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type QuestionRepo struct {
	db *gorm.DB
}

func NewQuestionRepo(db *gorm.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) FindByChapterID(chapterID uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("chapter_id = ? AND status = 1", chapterID).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByChapterIDFiltered(chapterID uint, difficulty string) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("chapter_id = ? AND status = 1", chapterID)
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindBySubSubjectID(subSubjectID uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("sub_subject_id = ? AND status = 1", subSubjectID).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByID(id uint) (*model.Question, error) {
	var q model.Question
	err := r.db.Where("id = ? AND status = 1", id).First(&q).Error
	return &q, err
}

func (r *QuestionRepo) FindRandom(subjectID uint, difficulty string, limit int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("subject_id = ? AND status = 1", subjectID)
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("RAND()").Limit(limit).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindSpecial(subjectID uint, qtype, difficulty string, limit int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("subject_id = ? AND status = 1", subjectID)
	if qtype != "" {
		query = query.Where("type = ?", qtype)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("RAND()").Limit(limit).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByIDs(ids []uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("id IN ? AND status = 1", ids).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) CountByChapterID(chapterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Question{}).
		Where("chapter_id = ? AND status = 1", chapterID).
		Count(&count).Error
	return count, err
}
