package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type PracticeRecordRepo struct {
	db *gorm.DB
}

func NewPracticeRecordRepo(db *gorm.DB) *PracticeRecordRepo {
	return &PracticeRecordRepo{db: db}
}

func (r *PracticeRecordRepo) Create(record *model.PracticeRecord) error {
	return r.db.Create(record).Error
}

func (r *PracticeRecordRepo) FindByUserAndQuestion(userID, questionID uint) (*model.PracticeRecord, error) {
	var record model.PracticeRecord
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Order("created_at desc").First(&record).Error
	return &record, err
}

func (r *PracticeRecordRepo) FindByUserAndChapter(userID, chapterID uint) ([]model.PracticeRecord, error) {
	var list []model.PracticeRecord
	err := r.db.Where("user_id = ? AND question_id IN (SELECT id FROM questions WHERE chapter_id = ?)",
		userID, chapterID).Find(&list).Error
	return list, err
}
