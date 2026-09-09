package repository

import (
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepo) FindByUsernameOrEmail(login string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ? OR email = ?", login, login).First(&user).Error
	return &user, err
}

func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 物理删除用户（依赖数据库外键级联清理其学习数据）
func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// CountByRole 统计指定角色的用户数量（用于「最后一个管理员」保护）
func (r *UserRepo) CountByRole(role string) (int64, error) {
	var total int64
	err := r.db.Model(&model.User{}).Where("role = ?", role).Count(&total).Error
	return total, err
}

// =============================================================
// 管理端用户列表：JOIN 出报考等级/科目名称，
// 支持「关键词 + 角色 + 状态」筛选与分页
// =============================================================

// AdminUserFilter 管理端用户筛选条件
type AdminUserFilter struct {
	Keyword string // 匹配 username / nickname / email
	Role    string // admin / student，其他值（含空串）视为全部
	Status  *int   // 1=正常 0=禁用，nil 视为全部
}

// AdminUserRow 管理端用户列表行（含 JOIN 出的等级/科目名）
type AdminUserRow struct {
	ID            uint
	Username      string
	Nickname      string
	Email         string
	EmailVerified bool
	Avatar        string
	Role          string
	Status        int
	LevelID       uint
	LevelName     string
	SubjectID     uint
	SubjectName   string
	CreatedAt     time.Time
}

// adminUserSelect 管理端用户 SELECT 片段（LEFT JOIN 取等级/科目名）
const adminUserSelect = `
	SELECT u.id             AS id,
	       u.username       AS username,
	       u.nickname       AS nickname,
	       u.email          AS email,
	       u.email_verified AS email_verified,
	       u.avatar         AS avatar,
	       u.role           AS role,
	       u.status         AS status,
	       u.level_id       AS level_id,
	       u.subject_id     AS subject_id,
	       u.created_at     AS created_at,
	       COALESCE(el.name, '') AS level_name,
	       COALESCE(s.name, '')  AS subject_name
	FROM users u
	LEFT JOIN exam_levels el ON el.id = u.level_id
	LEFT JOIN subjects s ON s.id = u.subject_id`

// buildAdminUserWhere 拼装 WHERE 子句与绑定参数
func buildAdminUserWhere(f AdminUserFilter) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, 3)

	sb.WriteString(" WHERE 1 = 1")
	if f.Role == "admin" || f.Role == "student" {
		sb.WriteString(" AND u.role = ?")
		args = append(args, f.Role)
	}
	if f.Status != nil {
		sb.WriteString(" AND u.status = ?")
		args = append(args, *f.Status)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		sb.WriteString(" AND (u.username LIKE ? OR COALESCE(u.nickname, '') LIKE ? OR u.email LIKE ?)")
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like, like)
	}
	return sb.String(), args
}

// AdminCount 按筛选条件统计用户总数
func (r *UserRepo) AdminCount(f AdminUserFilter) (int64, error) {
	where, args := buildAdminUserWhere(f)
	var total int64
	err := r.db.Raw("SELECT COUNT(*) FROM users u"+where, args...).Scan(&total).Error
	return total, err
}

// AdminFindByID 取单个用户的管理端列表行（新建/编辑后回查用于响应）
func (r *UserRepo) AdminFindByID(id uint) (*AdminUserRow, error) {
	var rows []AdminUserRow
	err := r.db.Raw(adminUserSelect+" WHERE u.id = ? LIMIT 1", id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rows[0], nil
}

// AdminList 按筛选条件倒序取一页用户
func (r *UserRepo) AdminList(f AdminUserFilter, limit, offset int) ([]AdminUserRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	where, args := buildAdminUserWhere(f)
	args = append(args, limit, offset)

	var rows []AdminUserRow
	err := r.db.Raw(adminUserSelect+where+
		" ORDER BY u.id DESC LIMIT ? OFFSET ?", args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []AdminUserRow{}
	}
	return rows, nil
}
