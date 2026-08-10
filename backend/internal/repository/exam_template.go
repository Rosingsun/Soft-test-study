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

// FindByIDs 批量按 ID 查询模板（消除 N+1）
func (r *ExamTemplateRepo) FindByIDs(ids []uint) (map[uint]model.ExamTemplate, error) {
	out := make(map[uint]model.ExamTemplate, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var list []model.ExamTemplate
	if err := r.db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, t := range list {
		out[t.ID] = t
	}
	return out, nil
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
