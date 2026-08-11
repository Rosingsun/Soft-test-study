package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type InvitationCodeRepo struct {
	db *gorm.DB
}

func NewInvitationCodeRepo(db *gorm.DB) *InvitationCodeRepo {
	return &InvitationCodeRepo{db: db}
}

// Create 写入新邀请码（依赖 code 唯一约束去重，重复时返回错误）
func (r *InvitationCodeRepo) Create(ic *model.InvitationCode) error {
	return r.db.Create(ic).Error
}

// FindByCode 按 code 查询（不附加状态/时间过滤，业务自行判定）
func (r *InvitationCodeRepo) FindByCode(code string) (*model.InvitationCode, error) {
	var ic model.InvitationCode
	err := r.db.Where("code = ?", code).First(&ic).Error
	if err != nil {
		return nil, err
	}
	return &ic, nil
}

// FindByCodeForUpdate 行级锁定后查询，调用方需在事务内使用，避免并发消费
func (r *InvitationCodeRepo) FindByCodeForUpdate(tx *gorm.DB, code string) (*model.InvitationCode, error) {
	var ic model.InvitationCode
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("code = ?", code).
		First(&ic).Error
	if err != nil {
		return nil, err
	}
	return &ic, nil
}

// MarkUsed 标记邀请码已被指定用户使用
func (r *InvitationCodeRepo) MarkUsed(tx *gorm.DB, id, usedBy uint) error {
	now := time.Now()
	return tx.Model(&model.InvitationCode{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"used_by": usedBy,
			"used_at": now,
		}).Error
}

// List 按状态/时间倒序分页查询
func (r *InvitationCodeRepo) List(status *int, limit, offset int) ([]model.InvitationCode, int64, error) {
	var list []model.InvitationCode
	var total int64

	q := r.db.Model(&model.InvitationCode{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
