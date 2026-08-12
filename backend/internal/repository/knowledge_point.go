package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/model"
)

type KnowledgePointRepo struct {
	db *gorm.DB
}

func NewKnowledgePointRepo(db *gorm.DB) *KnowledgePointRepo {
	return &KnowledgePointRepo{db: db}
}

// FindByName 按 name 查找全局知识点
func (r *KnowledgePointRepo) FindByName(name string) (*model.KnowledgePoint, error) {
	var kp model.KnowledgePoint
	err := r.db.Where("name = ?", name).First(&kp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &kp, nil
}

// CreateKnowledgePoint 创建全局知识点
func (r *KnowledgePointRepo) CreateKnowledgePoint(kp *model.KnowledgePoint) error {
	return r.db.Create(kp).Error
}

// UpdateKnowledgePoint 更新全局知识点（仅 description / subject_id）
func (r *KnowledgePointRepo) UpdateKnowledgePoint(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.KnowledgePoint{}).Where("id = ?", id).Updates(fields).Error
}

// FindUserKP 查询用户维度知识点
func (r *KnowledgePointRepo) FindUserKP(userID, knowledgePointID uint) (*model.UserKnowledgePoint, error) {
	var ukp model.UserKnowledgePoint
	err := r.db.Where("user_id = ? AND knowledge_point_id = ?", userID, knowledgePointID).First(&ukp).Error
	if err != nil {
		return nil, err
	}
	return &ukp, nil
}

// CreateUserKP 创建用户维度知识点
func (r *KnowledgePointRepo) CreateUserKP(ukp *model.UserKnowledgePoint) error {
	return r.db.Create(ukp).Error
}

// DeleteUserKP 删除用户维度知识点
func (r *KnowledgePointRepo) DeleteUserKP(userID, id uint) error {
	return r.db.Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.UserKnowledgePoint{}).Error
}

// UpdateUserKPFields 更新用户维度知识点字段
func (r *KnowledgePointRepo) UpdateUserKPFields(id uint, userID uint, fields map[string]interface{}) error {
	return r.db.Model(&model.UserKnowledgePoint{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields).Error
}

// ListUserKP 列表查询
type ListUserKPParams struct {
	UserID     uint
	Filter     string
	SubjectID  uint
	Keyword    string
	Page       int
	PageSize   int
}

func (r *KnowledgePointRepo) ListUserKP(p ListUserKPParams) ([]model.UserKnowledgePoint, int64, error) {
	q := r.db.Model(&model.UserKnowledgePoint{}).Where("user_id = ?", p.UserID)
	switch p.Filter {
	case "important":
		q = q.Where("is_important = ?", 1)
	case "mastered":
		q = q.Where("is_mastered = ?", 1)
	case "unmastered":
		q = q.Where("is_mastered = ?", 0)
	}
	if p.SubjectID > 0 {
		q = q.Where("id IN (SELECT id FROM (SELECT ukp.id FROM user_knowledge_points ukp INNER JOIN knowledge_points kp ON ukp.knowledge_point_id = kp.id WHERE kp.subject_id = ? AND ukp.user_id = ?) AS t)", p.SubjectID, p.UserID)
	}
	if p.Keyword != "" {
		q = q.Where("knowledge_point_id IN (SELECT id FROM knowledge_points WHERE name LIKE ?)", "%"+p.Keyword+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	var list []model.UserKnowledgePoint
	err := q.Order("created_at desc").
		Offset((p.Page - 1) * p.PageSize).
		Limit(p.PageSize).
		Find(&list).Error
	return list, total, err
}

// FindKnowledgePoints 批量取全局知识点
func (r *KnowledgePointRepo) FindKnowledgePoints(ids []uint) ([]model.KnowledgePoint, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.KnowledgePoint
	err := r.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// StatsByUser 统计
func (r *KnowledgePointRepo) StatsByUser(userID uint) (total, important, mastered, unmastered, recentWeek int64, err error) {
	if err = r.db.Model(&model.UserKnowledgePoint{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return
	}
	if err = r.db.Model(&model.UserKnowledgePoint{}).Where("user_id = ? AND is_important = 1", userID).Count(&important).Error; err != nil {
		return
	}
	if err = r.db.Model(&model.UserKnowledgePoint{}).Where("user_id = ? AND is_mastered = 1", userID).Count(&mastered).Error; err != nil {
		return
	}
	unmastered = total - mastered
	if err = r.db.Model(&model.UserKnowledgePoint{}).
		Where("user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)", userID).
		Count(&recentWeek).Error; err != nil {
		return
	}
	return
}
