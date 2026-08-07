package repository

import (
	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/model"
)

type ReviewCardRepo struct {
	db *gorm.DB
}

func NewReviewCardRepo(db *gorm.DB) *ReviewCardRepo {
	return &ReviewCardRepo{db: db}
}

func (r *ReviewCardRepo) FindByUserAndQuestion(userID, questionID uint) (*model.ReviewCard, error) {
	var card model.ReviewCard
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&card).Error
	return &card, err
}

func (r *ReviewCardRepo) Create(card *model.ReviewCard) error {
	return r.db.Create(card).Error
}

func (r *ReviewCardRepo) Save(card *model.ReviewCard) error {
	return r.db.Save(card).Error
}

func (r *ReviewCardRepo) FindDueByUser(userID uint, date string, limit int) ([]model.ReviewCard, error) {
	var list []model.ReviewCard
	err := r.db.Where("user_id = ? AND status = 1 AND due_date <= ?", userID, date).
		Order("due_date asc, id asc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *ReviewCardRepo) CountDue(userID uint, date string) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = 1 AND due_date <= ?", userID, date).Count(&n).Error
	return n, err
}

func (r *ReviewCardRepo) CountDueBetween(userID uint, startDate, endDate string) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = 1 AND due_date >= ? AND due_date <= ?", userID, startDate, endDate).
		Count(&n).Error
	return n, err
}

func (r *ReviewCardRepo) CountByStatus(userID uint, status int) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = ?", userID, status).Count(&n).Error
	return n, err
}
