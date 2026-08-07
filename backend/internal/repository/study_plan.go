package repository

import (
	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/model"
)

type StudyPlanRepo struct {
	db *gorm.DB
}

func NewStudyPlanRepo(db *gorm.DB) *StudyPlanRepo {
	return &StudyPlanRepo{db: db}
}

func (r *StudyPlanRepo) FindByUser(userID uint) ([]model.StudyPlan, error) {
	var list []model.StudyPlan
	err := r.db.Where("user_id = ?", userID).Order("id desc").Find(&list).Error
	return list, err
}

func (r *StudyPlanRepo) FindByID(id uint) (*model.StudyPlan, error) {
	var p model.StudyPlan
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *StudyPlanRepo) FindActiveByUser(userID uint) ([]model.StudyPlan, error) {
	var list []model.StudyPlan
	err := r.db.Where("user_id = ? AND status = 1", userID).Find(&list).Error
	return list, err
}

func (r *StudyPlanRepo) Create(p *model.StudyPlan) error {
	return r.db.Create(p).Error
}

func (r *StudyPlanRepo) Save(p *model.StudyPlan) error {
	return r.db.Save(p).Error
}

func (r *StudyPlanRepo) Delete(p *model.StudyPlan) error {
	return r.db.Delete(p).Error
}
