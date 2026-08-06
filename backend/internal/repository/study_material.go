package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type StudyMaterialRepo struct {
	db *gorm.DB
}

func NewStudyMaterialRepo(db *gorm.DB) *StudyMaterialRepo {
	return &StudyMaterialRepo{db: db}
}

// FindList 查询上线的学习资料，subjectID 为 0 时返回全部
func (r *StudyMaterialRepo) FindList(subjectID uint) ([]model.StudyMaterial, error) {
	var list []model.StudyMaterial
	query := r.db.Where("status = 1")
	if subjectID > 0 {
		query = query.Where("subject_id = ?", subjectID)
	}
	err := query.Order("sort_order desc, id desc").Find(&list).Error
	return list, err
}

func (r *StudyMaterialRepo) FindByID(id uint) (*model.StudyMaterial, error) {
	var m model.StudyMaterial
	err := r.db.First(&m, id).Error
	return &m, err
}

func (r *StudyMaterialRepo) IncrementDownloadCount(id uint) error {
	return r.db.Model(&model.StudyMaterial{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *StudyMaterialRepo) IncrementViewCount(id uint) error {
	return r.db.Model(&model.StudyMaterial{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
