package dto

// =============================================================
// 管理端（/api/v1/admin/*）专用 dto
//
// 约定（与 AGENTS.md 一致）：
//   - 请求/响应字段一律 snake_case，且与前端 types/admin.ts 完全同名
//   - 用户详情聚合响应不得包含 password_hash 等敏感字段
// =============================================================

// ListInvitationCodeReq 管理端邀请码列表查询参数（GET query）
//
// State 取值：
//   - "" / all：全部
//   - unused：未使用（used_by IS NULL 且 status=1）
//   - used：已使用（used_by 非空）
//   - disabled：已停用（status=0）
//
// Keyword 同时匹配：邀请码 code、批次备注 note、使用人 username
type ListInvitationCodeReq struct {
	State     string `form:"state"`
	Keyword   string `form:"keyword"`
	CreatedBy uint   `form:"created_by"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

// InvitationCodeResp 邀请码列表行（含关联用户名）
//
// State 由后端统一判定（unused/used/disabled），前端不重复推演，直接用于渲染状态标签。
type InvitationCodeResp struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Note          string `json:"note"`
	Status        int    `json:"status"`
	State         string `json:"state"`
	CreatedBy     uint   `json:"created_by"`
	CreatedByName string `json:"created_by_name"`
	UsedBy        *uint  `json:"used_by"`
	UsedByName    string `json:"used_by_name"`
	UsedAt        string `json:"used_at"`
	CreatedAt     string `json:"created_at"`
}

// CreateInvitationCodeReq 生成邀请码请求（POST body）
//
// Count 省略或 <=0 时按 1 处理；单次上限由 service 层兜底。
type CreateInvitationCodeReq struct {
	Count int    `json:"count"`
	Note  string `json:"note" binding:"max=200"`
}

// =============================================================
// 用户管理
// =============================================================

// ListAdminUserReq 管理端用户列表查询参数（GET query）
//
// Keyword 同时匹配：用户名 username、昵称 nickname、邮箱 email
// Role 取值："" / all（全部）、admin、student
// Status 取值：nil（全部）、1（正常）、0（禁用）
type ListAdminUserReq struct {
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// AdminUserListItem 管理端用户列表行（不含任何凭据字段）
type AdminUserListItem struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Avatar        string `json:"avatar"`
	Role          string `json:"role"`
	Status        int    `json:"status"`
	LevelID       uint   `json:"level_id"`
	LevelName     string `json:"level_name"`
	SubjectID     uint   `json:"subject_id"`
	SubjectName   string `json:"subject_name"`
	CreatedAt     string `json:"created_at"`
}

// CreateAdminUserReq 管理员新建用户（POST body）
//
// Role 省略时默认 student；Status 省略时默认 1（正常）
type CreateAdminUserReq struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email,max=100"`
	Password  string `json:"password" binding:"required,min=8,max=128"`
	Nickname  string `json:"nickname" binding:"omitempty,max=50"`
	Role      string `json:"role" binding:"omitempty,oneof=student admin"`
	Status    *int   `json:"status"`
	LevelID   uint   `json:"level_id"`
	SubjectID uint   `json:"subject_id"`
}

// UpdateAdminUserReq 管理员编辑用户（PUT body）
//
// 全部字段可选，指针类型用于区分「未传」与「传零值」
type UpdateAdminUserReq struct {
	Nickname  *string `json:"nickname" binding:"omitempty,max=50"`
	Email     *string `json:"email" binding:"omitempty,email,max=100"`
	Role      *string `json:"role" binding:"omitempty,oneof=student admin"`
	Status    *int    `json:"status"`
	LevelID   *uint   `json:"level_id"`
	SubjectID *uint   `json:"subject_id"`
}

// UpdateAdminUserStatusReq 启/禁用用户（PATCH body）
type UpdateAdminUserStatusReq struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// ResetAdminUserPasswordReq 管理员重置用户密码（POST body）
type ResetAdminUserPasswordReq struct {
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// AdminUserProfile 管理端用户档案（不含任何凭据字段）
type AdminUserProfile struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Avatar        string `json:"avatar"`
	Role          string `json:"role"`
	Status        int    `json:"status"`
	LevelID       uint   `json:"level_id"`
	LevelName     string `json:"level_name"`
	SubjectID     uint   `json:"subject_id"`
	SubjectName   string `json:"subject_name"`
	Difficulty    string `json:"difficulty"`
	CreatedAt     string `json:"created_at"`
}

// AdminUserStudyResp 用户学习信息：概览 + 近期活跃 + 科目/章节掌握度
type AdminUserStudyResp struct {
	Overview StatsOverviewResp     `json:"overview"`
	Recent   []DailyStatsResp      `json:"recent"`
	Subjects []SubjectProgressResp `json:"subjects"`
	Chapters []ChapterProgressResp `json:"chapters"`
}

// AdminUserAnswerResp 用户答题/考试信息
type AdminUserAnswerResp struct {
	TotalExams   int64            `json:"total_exams"`
	AvgExamScore float64          `json:"avg_exam_score"`
	Records      []ExamRecordResp `json:"records"`
}

// AdminUserDetailResp 管理端用户详情聚合：邀请码 → 用户 → 学习/答题全貌
//
// InviteCode 可能为 nil（管理员白名单注册的用户不消耗邀请码）。
type AdminUserDetailResp struct {
	User       AdminUserProfile    `json:"user"`
	InviteCode *InvitationCodeResp `json:"invite_code"`
	Study      AdminUserStudyResp  `json:"study"`
	Answer     AdminUserAnswerResp `json:"answer"`
}

// =============================================================
// 题库管理
// =============================================================

// AdminQuestionListReq 管理端题目列表查询参数（GET query）
//
// Status 取值：""（全部）/ "1"（启用）/ "0"（停用）。
// 之所以用字符串而非 int：0 本身是合法的停用状态值，用 int 无法区分"未传"与"查停用"。
type AdminQuestionListReq struct {
	SubjectID    uint   `form:"subject_id"`
	SubSubjectID uint   `form:"sub_subject_id"`
	ChapterID    uint   `form:"chapter_id"`
	Type         string `form:"type"`
	Difficulty   string `form:"difficulty"`
	Status       string `form:"status"`
	Year         int    `form:"year"`
	Keyword      string `form:"keyword"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// AdminQuestionChildResp 案例分析大题目下的小题目（仅详情接口填充）
type AdminQuestionChildResp struct {
	ID       uint   `json:"id"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Answer   string `json:"answer"`
	Analysis string `json:"analysis"`
}

// AdminQuestionResp 管理端题目行（含 JOIN 出的科目/子科目/章节名称）
//
// Options / BlankOptions 原样返回 JSON 字符串，由前端解析后渲染，
// 避免后端与前端维护两套结构导致格式漂移。
type AdminQuestionResp struct {
	ID             uint                     `json:"id"`
	SubjectID      uint                     `json:"subject_id"`
	SubjectName    string                   `json:"subject_name"`
	SubSubjectID   uint                     `json:"sub_subject_id"`
	SubSubjectName string                   `json:"sub_subject_name"`
	ChapterID      uint                     `json:"chapter_id"`
	ChapterName    string                   `json:"chapter_name"`
	ParentID       uint                     `json:"parent_id"`
	Type           string                   `json:"type"`
	Difficulty     string                   `json:"difficulty"`
	Content        string                   `json:"content"`
	CaseMaterial   string                   `json:"case_material"`
	Options        string                   `json:"options"`
	BlankOptions   string                   `json:"blank_options"`
	Answer         string                   `json:"answer"`
	Analysis       string                   `json:"analysis"`
	Year           int                      `json:"year"`
	Source         string                   `json:"source"`
	Status         int                      `json:"status"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
	Children       []AdminQuestionChildResp `json:"children,omitempty"`
}

// AdminQuestionUpsertReq 新增 / 更新题目（POST / PUT body）
//
// Status 为指针：新增时省略默认启用（1）；更新时省略表示不改动当前状态。
// SubjectID / Type / Difficulty / Content / Answer 由 service 层校验。
type AdminQuestionUpsertReq struct {
	SubjectID    uint   `json:"subject_id"`
	SubSubjectID uint   `json:"sub_subject_id"`
	ChapterID    uint   `json:"chapter_id"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	CaseMaterial string `json:"case_material"`
	Options      string `json:"options"`
	BlankOptions string `json:"blank_options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Year         int    `json:"year"`
	Source       string `json:"source"`
	Status       *int   `json:"status"`
}

// AdminBatchQuestionStatusReq 批量启用 / 停用题目
type AdminBatchQuestionStatusReq struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status" binding:"oneof=0 1"`
}

// AdminQuestionTypeStat 按题型统计题量
type AdminQuestionTypeStat struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// AdminQuestionStatsResp 题库概览：总量 / 启用 / 停用 / 各题型分布
type AdminQuestionStatsResp struct {
	Total    int64                   `json:"total"`
	Enabled  int64                   `json:"enabled"`
	Disabled int64                   `json:"disabled"`
	ByType   []AdminQuestionTypeStat `json:"by_type"`
}

// =============================================================
// 科目管理
// =============================================================

// ListAdminSubjectReq 管理端科目列表查询参数（GET query）
// LevelID 省略或 0 表示全部等级。
type ListAdminSubjectReq struct {
	LevelID uint   `form:"level_id"`
	Keyword string `form:"keyword"`
}

// AdminSubjectResp 管理端科目行（含等级名称、状态与引用统计）
type AdminSubjectResp struct {
	ID              uint   `json:"id"`
	LevelID         uint   `json:"level_id"`
	LevelName       string `json:"level_name"`
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	Description     string `json:"description"`
	Icon            string `json:"icon"`
	SortOrder       int    `json:"sort_order"`
	Status          int    `json:"status"`
	SubSubjectCount int64  `json:"sub_subject_count"`
	QuestionCount   int64  `json:"question_count"`
}

// AdminSubjectUpsertReq 新增 / 更新科目（POST / PUT body）
// 更新时省略 LevelID 不修改所属等级。
type AdminSubjectUpsertReq struct {
	LevelID     uint   `json:"level_id"`
	Name        string `json:"name" binding:"required,max=100"`
	ShortName   string `json:"short_name" binding:"max=20"`
	Description string `json:"description"`
	Icon        string `json:"icon" binding:"max=255"`
	SortOrder   int    `json:"sort_order"`
	Status      *int   `json:"status"`
}
