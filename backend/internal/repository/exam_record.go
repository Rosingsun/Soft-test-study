package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type ExamRecordRepo struct {
	db *gorm.DB
}

func NewExamRecordRepo(db *gorm.DB) *ExamRecordRepo {
	return &ExamRecordRepo{db: db}
}

func (r *ExamRecordRepo) Create(record *model.ExamRecord) error {
	return r.db.Create(record).Error
}

func (r *ExamRecordRepo) FindByID(id uint) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	err := r.db.First(&rec, id).Error
	return &rec, err
}

func (r *ExamRecordRepo) FindByUser(userID uint) ([]model.ExamRecord, error) {
	var list []model.ExamRecord
	err := r.db.Where("user_id = ?", userID).
		Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *ExamRecordRepo) Save(record *model.ExamRecord) error {
	return r.db.Save(record).Error
}

func (r *ExamRecordRepo) FindPendingByUserAndTemplate(userID, templateID uint) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	err := r.db.Where("user_id = ? AND template_id = ? AND status = 'pending'",
		userID, templateID).First(&rec).Error
	return &rec, err
}

func (r *ExamRecordRepo) SaveAnswer(answer *model.ExamRecordAnswer) error {
	var existing model.ExamRecordAnswer
	err := r.db.Where("record_id = ? AND question_id = ?", answer.RecordID, answer.QuestionID).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(answer).Error
	}
	if err != nil {
		return err
	}
	existing.Answer = answer.Answer
	existing.IsCorrect = answer.IsCorrect
	existing.Score = answer.Score
	return r.db.Save(&existing).Error
}

func (r *ExamRecordRepo) CreateBatchAnswers(answers []model.ExamRecordAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return r.db.CreateInBatches(answers, 100).Error
}

func (r *ExamRecordRepo) FindAnswersByRecord(recordID uint) ([]model.ExamRecordAnswer, error) {
	var list []model.ExamRecordAnswer
	err := r.db.Where("record_id = ?", recordID).Order("id asc").Find(&list).Error
	return list, err
}
