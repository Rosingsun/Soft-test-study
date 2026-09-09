package repository

import (
	"strings"
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

// =============================================================
// 管理端列表：需要 JOIN users 带出使用人/生成者用户名，
// 并支持「状态 + 创建者 + 关键词」筛选，故单列一段查询逻辑。
// =============================================================

// adminInvitationSelect 管理端邀请码 SELECT 片段：
// 两次 LEFT JOIN users 分别取生成者与使用人用户名（created_by 允许为 0，JOIN 不到时回落到空串）
const adminInvitationSelect = `
	SELECT ic.id               AS id,
	       ic.code             AS code,
	       ic.note             AS note,
	       ic.status           AS status,
	       ic.created_by       AS created_by,
	       ic.used_by          AS used_by,
	       ic.used_at          AS used_at,
	       ic.created_at       AS created_at,
	       COALESCE(cu.username, '') AS created_by_name,
	       COALESCE(uu.username, '') AS used_by_name
	FROM invitation_codes ic
	LEFT JOIN users cu ON cu.id = ic.created_by
	LEFT JOIN users uu ON uu.id = ic.used_by`

// AdminInvitationCodeRow 管理端列表行（含 JOIN 出来的用户名）
type AdminInvitationCodeRow struct {
	ID            uint
	Code          string
	Note          string
	Status        int
	CreatedBy     uint
	CreatedByName string
	UsedBy        *uint
	UsedByName    string
	UsedAt        *time.Time
	CreatedAt     time.Time
}

// AdminInvitationCodeFilter 管理端筛选条件
type AdminInvitationCodeFilter struct {
	State     string // unused / used / disabled，其他值（含空串）视为全部
	Keyword   string // 匹配 code / note / 使用人 username
	CreatedBy uint   // 0 表示不限
}

// buildAdminInvitationWhere 拼装 WHERE 子句与绑定参数。
//
// 状态语义与注册逻辑保持一致：
//   - unused：used_by IS NULL 且 status=1
//   - used：used_by 非空
//   - disabled：status=0
func buildAdminInvitationWhere(f AdminInvitationCodeFilter) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, 5)

	switch f.State {
	case "unused":
		sb.WriteString(" WHERE ic.used_by IS NULL AND ic.status = 1")
	case "used":
		sb.WriteString(" WHERE ic.used_by IS NOT NULL")
	case "disabled":
		sb.WriteString(" WHERE ic.status = 0")
	default:
		sb.WriteString(" WHERE 1 = 1")
	}

	if f.CreatedBy > 0 {
		sb.WriteString(" AND ic.created_by = ?")
		args = append(args, f.CreatedBy)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		sb.WriteString(" AND (ic.code LIKE ? OR ic.note LIKE ? OR COALESCE(uu.username, '') LIKE ?)")
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like, like)
	}
	return sb.String(), args
}

// escapeLike 转义 LIKE 通配符，避免用户输入 %/_ 造成意外匹配
func escapeLike(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return replacer.Replace(s)
}

// AdminCount 按筛选条件统计总数
func (r *InvitationCodeRepo) AdminCount(f AdminInvitationCodeFilter) (int64, error) {
	where, args := buildAdminInvitationWhere(f)
	var total int64
	err := r.db.Raw("SELECT COUNT(*) FROM invitation_codes ic"+
		" LEFT JOIN users uu ON uu.id = ic.used_by"+where, args...).Scan(&total).Error
	return total, err
}

// AdminList 按筛选条件倒序取一页数据
func (r *InvitationCodeRepo) AdminList(f AdminInvitationCodeFilter, limit, offset int) ([]AdminInvitationCodeRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	where, args := buildAdminInvitationWhere(f)
	args = append(args, limit, offset)

	var rows []AdminInvitationCodeRow
	err := r.db.Raw(adminInvitationSelect+where+
		" ORDER BY ic.id DESC LIMIT ? OFFSET ?", args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []AdminInvitationCodeRow{}
	}
	return rows, nil
}

// AdminFindByUsedBy 取某个用户注册时使用的邀请码（正常场景唯一，取最新一条）
func (r *InvitationCodeRepo) AdminFindByUsedBy(usedBy uint) (*AdminInvitationCodeRow, error) {
	var rows []AdminInvitationCodeRow
	err := r.db.Raw(adminInvitationSelect+
		" WHERE ic.used_by = ? ORDER BY ic.used_at DESC LIMIT 1", usedBy).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rows[0], nil
}
