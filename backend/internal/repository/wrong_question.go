package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type WrongQuestionRepo struct {
	db *gorm.DB
}

func NewWrongQuestionRepo(db *gorm.DB) *WrongQuestionRepo {
	return &WrongQuestionRepo{db: db}
}

func (r *WrongQuestionRepo) Upsert(userID, questionID uint) error {
	var wq model.WrongQuestion
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&wq).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.WrongQuestion{
			UserID:      userID,
			QuestionID:  questionID,
			WrongCount:  1,
			LastWrongAt: time.Now(),
		}).Error
	}
	if err != nil {
		return err
	}
	wq.WrongCount++
	wq.LastWrongAt = time.Now()
	return r.db.Save(&wq).Error
}

func (r *WrongQuestionRepo) IncrementCorrect(userID, questionID uint) error {
	return r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Update("correct_count", gorm.Expr("correct_count + 1")).Error
}

func (r *WrongQuestionRepo) Delete(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.WrongQuestion{}).Error
}

func (r *WrongQuestionRepo) FindByUser(userID uint) ([]model.WrongQuestion, error) {
	var list []model.WrongQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("last_wrong_at desc").Find(&list).Error
	return list, err
}

func (r *WrongQuestionRepo) FindByUserAndSubject(userID, subjectID uint) ([]model.WrongQuestion, error) {
	var list []model.WrongQuestion
	err := r.db.Where("user_id = ? AND question_id IN (SELECT id FROM questions WHERE subject_id = ?)",
		userID, subjectID).Order("last_wrong_at desc").Find(&list).Error
	return list, err
}

func (r *WrongQuestionRepo) FindByUserAndQuestion(userID, questionID uint) (*model.WrongQuestion, error) {
	var wq model.WrongQuestion
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&wq).Error
	return &wq, err
}

func (r *WrongQuestionRepo) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
