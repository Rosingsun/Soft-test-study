package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

// AdminHandler 管理端接口处理器。
//
// 本文件所有路由都必须挂在带 Auth + RequireAdmin 的 /admin 组下，
// handler 内只做参数绑定 → 调 service → 返回响应，不写业务逻辑。
type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// ListInvitationCodes GET /api/v1/admin/invitation-codes
// 支持 state / keyword / created_by 筛选与分页
func (h *AdminHandler) ListInvitationCodes(c *gin.Context) {
	var req dto.ListInvitationCodeReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.svc.ListInvitationCodes(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取邀请码列表失败")
		return
	}

	page, pageSize := normalizePageParams(req.Page, req.PageSize)
	response.SuccessPage(c, list, total, page, pageSize)
}

// CreateInvitationCode POST /api/v1/admin/invitation-codes
// 生成邀请码，生成者记为当前管理员
func (h *AdminHandler) CreateInvitationCode(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}

	var req dto.CreateInvitationCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	list, err := h.svc.GenerateInvitationCodes(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "生成邀请码失败: "+err.Error())
		return
	}
	response.Success(c, list)
}

// ListUsers GET /api/v1/admin/users
// 管理端用户列表：关键词 / 角色 / 状态筛选 + 分页
func (h *AdminHandler) ListUsers(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}

	var req dto.ListAdminUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.svc.ListUsers(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取用户列表失败")
		return
	}
	page, pageSize := normalizePageParams(req.Page, req.PageSize)
	response.SuccessPage(c, list, total, page, pageSize)
}

// CreateUser POST /api/v1/admin/users
// 管理员直接新建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}

	var req dto.CreateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	item, err := h.svc.CreateUser(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, item)
}

// UpdateUser PUT /api/v1/admin/users/:id
// 管理员编辑用户资料
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	operatorID, ok := mustUserID(c)
	if !ok {
		return
	}
	id, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	var req dto.UpdateAdminUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	item, err := h.svc.UpdateUser(operatorID, id, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, item)
}

// UpdateUserStatus PATCH /api/v1/admin/users/:id/status
// 启用 / 禁用用户
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	operatorID, ok := mustUserID(c)
	if !ok {
		return
	}
	id, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	var req dto.UpdateAdminUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	item, err := h.svc.UpdateUserStatus(operatorID, id, req.Status)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, item)
}

// ResetUserPassword POST /api/v1/admin/users/:id/reset-password
// 管理员重置用户密码
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}
	id, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	var req dto.ResetAdminUserPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if err := h.svc.ResetUserPassword(id, req.Password); err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// DeleteUser DELETE /api/v1/admin/users/:id
// 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	operatorID, ok := mustUserID(c)
	if !ok {
		return
	}
	id, ok := parseUserIDParam(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteUser(operatorID, id); err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// GetUserDetail GET /api/v1/admin/users/:id
// 聚合用户档案 + 学习信息 + 答题信息
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return
	}

	detail, err := h.svc.GetUserDetail(uint(id))
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取用户详情失败: "+err.Error())
		return
	}
	response.Success(c, detail)
}

// ListSubjects GET /api/v1/admin/subjects?level_id=&keyword=
func (h *AdminHandler) ListSubjects(c *gin.Context) {
	var req dto.ListAdminSubjectReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误: "+err.Error())
		return
	}
	list, err := h.svc.ListSubjects(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取科目列表失败")
		return
	}
	response.Success(c, list)
}

// CreateSubject POST /api/v1/admin/subjects
func (h *AdminHandler) CreateSubject(c *gin.Context) {
	var req dto.AdminSubjectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}
	subject, err := h.svc.CreateSubject(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "新增科目失败: "+err.Error())
		return
	}
	response.Success(c, subject)
}

// UpdateSubject PUT /api/v1/admin/subjects/:id
func (h *AdminHandler) UpdateSubject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req dto.AdminSubjectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}
	subject, err := h.svc.UpdateSubject(id, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "更新科目失败: "+err.Error())
		return
	}
	response.Success(c, subject)
}

// SetSubjectStatus PATCH /api/v1/admin/subjects/:id/status  body: {"status":0|1}
func (h *AdminHandler) SetSubjectStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "status 必须为 0 或 1")
		return
	}
	if err := h.svc.SetSubjectStatus(id, req.Status); err != nil {
		response.Error(c, config.CodeBadRequest, "更新状态失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteSubject DELETE /api/v1/admin/subjects/:id
func (h *AdminHandler) DeleteSubject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteSubject(id); err != nil {
		response.Error(c, config.CodeBadRequest, "删除科目失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// ===================== 题库管理 =====================

// ListQuestions GET /api/v1/admin/questions
// 支持 科目/子科目/章节/题型/难度/状态/年份/关键词 筛选与分页
func (h *AdminHandler) ListQuestions(c *gin.Context) {
	var req dto.AdminQuestionListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误: "+err.Error())
		return
	}
	// 题型 / 难度非法值直接忽略（等同"不限"），避免误传导致空列表
	if req.Type != "" && !allowedQuestionTypes[req.Type] {
		req.Type = ""
	}
	if req.Difficulty != "" && !allowedDifficulties[req.Difficulty] {
		req.Difficulty = ""
	}

	list, total, err := h.svc.ListQuestions(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取题目列表失败")
		return
	}

	page, pageSize := normalizePageParams(req.Page, req.PageSize)
	response.SuccessPage(c, list, total, page, pageSize)
}

// GetQuestion GET /api/v1/admin/questions/:id
func (h *AdminHandler) GetQuestion(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	q, err := h.svc.GetQuestion(id)
	if err != nil {
		response.Error(c, config.CodeNotFound, "题目不存在")
		return
	}
	response.Success(c, q)
}

// CreateQuestion POST /api/v1/admin/questions
func (h *AdminHandler) CreateQuestion(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}

	var req dto.AdminQuestionUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	q, err := h.svc.CreateQuestion(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, q)
}

// UpdateQuestion PUT /api/v1/admin/questions/:id
func (h *AdminHandler) UpdateQuestion(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req dto.AdminQuestionUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	q, err := h.svc.UpdateQuestion(id, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, q)
}

// DeleteQuestion DELETE /api/v1/admin/questions/:id
// 软删除：仅将 status 置 0，保留历史练习/考试记录引用
func (h *AdminHandler) DeleteQuestion(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteQuestion(id); err != nil {
		respondError(c, err, "删除题目失败")
		return
	}
	response.Success(c, gin.H{"id": id})
}

// BatchUpdateQuestionStatus PATCH /api/v1/admin/questions/status
// 批量启用（status=1）/ 停用（status=0）
func (h *AdminHandler) BatchUpdateQuestionStatus(c *gin.Context) {
	if _, ok := mustUserID(c); !ok {
		return
	}

	var req dto.AdminBatchQuestionStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if err := h.svc.BatchSetQuestionStatus(req.IDs, req.Status); err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": len(req.IDs)})
}

// QuestionStats GET /api/v1/admin/stats/questions
// 题库概览：总量 / 启用 / 停用 / 各题型分布
func (h *AdminHandler) QuestionStats(c *gin.Context) {
	stats, err := h.svc.QuestionStats()
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取题库统计失败")
		return
	}
	response.Success(c, stats)
}

// allowedQuestionTypes / allowedDifficulties 与 service 层白名单保持一致，
// 用于在参数绑定阶段过滤前端误传的非法枚举值
var (
	allowedQuestionTypes = map[string]bool{
		"single": true, "multi": true, "judge": true, "fill": true,
		"short": true, "comprehensive": true, "essay": true,
		"case_study": true, "multi_blank": true,
	}
	allowedDifficulties = map[string]bool{"easy": true, "medium": true, "hard": true}
)

// parseIDParam 解析路径中的 :id，失败时直接写 400 响应
func parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return 0, false
	}
	return uint(id), true
}
