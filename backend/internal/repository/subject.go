package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type SubjectRepo struct {
	db *gorm.DB
}

func NewSubjectRepo(db *gorm.DB) *SubjectRepo {
	return &SubjectRepo{db: db}
}

func (r *SubjectRepo) FindByLevelID(levelID uint) ([]model.Subject, error) {
	var subjects []model.Subject
	err := r.db.Where("level_id = ? AND status = 1", levelID).
		Order("sort_order asc").
		Find(&subjects).Error
	return subjects, err
}

func (r *SubjectRepo) FindByID(id uint) (*model.Subject, error) {
	var subject model.Subject
	err := r.db.First(&subject, id).Error
	return &subject, err
}

func (r *SubjectRepo) FindAll() ([]model.Subject, error) {
	var subjects []model.Subject
	err := r.db.Where("status = 1").Order("sort_order asc").Find(&subjects).Error
	return subjects, err
}

// AdminFindAll 管理端科目全量列表（含停用科目），可按键名模糊筛选，按等级与排序返回。
func (r *SubjectRepo) AdminFindAll(levelID uint, keyword string) ([]model.Subject, error) {
	q := r.db.Model(&model.Subject{})
	if levelID > 0 {
		q = q.Where("level_id = ?", levelID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR short_name LIKE ?", like, like)
	}
	var subjects []model.Subject
	err := q.Order("level_id asc, sort_order asc").Find(&subjects).Error
	return subjects, err
}

// CountByName 统计同等级下同名（含同简称）科目数量，用于唯一性校验（可排除自身 id）。
func (r *SubjectRepo) CountByName(levelID uint, name string, excludeID uint) (int64, error) {
	var count int64
	q := r.db.Model(&model.Subject{}).Where("level_id = ? AND name = ?", levelID, name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&count).Error
	return count, err
}

// CountByIDs 返回指定科目 id 的现存数量（用于批量软删前校验）。
func (r *SubjectRepo) CountByIDs(ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&model.Subject{}).Where("id IN ?", ids).Count(&count).Error
	return count, err
}

// Create 新增科目。
func (r *SubjectRepo) Create(subject *model.Subject) error {
	return r.db.Create(subject).Error
}

// Update 更新科目字段。
func (r *SubjectRepo) Update(subject *model.Subject) error {
	return r.db.Save(subject).Error
}

// UpdateStatus 软删/恢复：1=启用 0=停用。
func (r *SubjectRepo) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Subject{}).Where("id = ?", id).
		Update("status", status).Error
}

// CountSubSubjects 统计科目下的子科目数量（用于删除前引用检查）。
func (r *SubjectRepo) CountSubSubjects(subjectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SubSubject{}).Where("subject_id = ?", subjectID).Count(&count).Error
	return count, err
}

// CountQuestions 统计科目下的题目数量（含其子科目/章节的题目）。
func (r *SubjectRepo) CountQuestions(subjectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Question{}).Where("subject_id = ?", subjectID).Count(&count).Error
	return count, err
}

// Delete 物理删除科目（调用方需先保证无子科目 / 题目等外键引用）。
func (r *SubjectRepo) Delete(id uint) error {
	return r.db.Delete(&model.Subject{}, id).Error
}
