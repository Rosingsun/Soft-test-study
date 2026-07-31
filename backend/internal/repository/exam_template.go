package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type ExamTemplateRepo struct {
	db *gorm.DB
}

func NewExamTemplateRepo(db *gorm.DB) *ExamTemplateRepo {
	return &ExamTemplateRepo{db: db}
}

func (r *ExamTemplateRepo) FindPublic() ([]model.ExamTemplate, error) {
	var list []model.ExamTemplate
	err := r.db.Where("is_public = 1 AND status = 1").
		Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *ExamTemplateRepo) FindByID(id uint) (*model.ExamTemplate, error) {
	var t model.ExamTemplate
	err := r.db.First(&t, id).Error
	return &t, err
}

func (r *ExamTemplateRepo) FindQuestionsByTemplate(templateID uint) ([]model.ExamTemplateQuestion, error) {
	var list []model.ExamTemplateQuestion
	err := r.db.Where("template_id = ?", templateID).
		Order("sort_order asc").Find(&list).Error
	return list, err
}

func (r *ExamTemplateRepo) CountQuestions(templateID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ExamTemplateQuestion{}).
		Where("template_id = ?", templateID).Count(&count).Error
	return count, err
}
