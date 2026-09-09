package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/pkg/security"
)

// adminExamRecordLimit 用户详情里「答题信息」最多返回的考试记录条数（展示用，非统计口径）
const adminExamRecordLimit = 10

// adminRecentStudyDays 用户详情里「学习信息」回看的天数
const adminRecentStudyDays = 30

// maxGeneratePerRequest 单次允许生成的最大邀请码数量（GUI 场景远小于 CLI 批量的 1000）
const maxGeneratePerRequest = 100

// maxBatchQuestionIDs 单次批量启停允许的题目数量上限，防止误传超大 ID 列表拖垮数据库
const maxBatchQuestionIDs = 200

// AdminService 管理端业务编排层。
//
// 设计要点：
//   - 只读 + 少量写操作，不引入新的数据表
//   - 统计口径全部复用 StatsService / ExamRecordRepo，避免 SQL 逻辑重复漂移
//   - 所有 SQL 仍在 repository 层，本层只做编排与拼装
type AdminService struct {
	userRepo       *repository.UserRepo
	levelRepo      *repository.ExamLevelRepo
	subjectRepo    *repository.SubjectRepo
	subSubjectRepo *repository.SubSubjectRepo
	chapterRepo    *repository.ChapterRepo
	questionRepo   *repository.QuestionRepo
	invitationRepo *repository.InvitationCodeRepo
	invitationSvc  *InvitationCodeService
	statsSvc       *StatsService
	examRecordRepo *repository.ExamRecordRepo
	examTplRepo    *repository.ExamTemplateRepo
}

func NewAdminService(
	userRepo *repository.UserRepo,
	levelRepo *repository.ExamLevelRepo,
	subjectRepo *repository.SubjectRepo,
	invitationRepo *repository.InvitationCodeRepo,
	invitationSvc *InvitationCodeService,
	statsSvc *StatsService,
	examRecordRepo *repository.ExamRecordRepo,
	examTplRepo *repository.ExamTemplateRepo,
	questionRepo *repository.QuestionRepo,
	subSubjectRepo *repository.SubSubjectRepo,
	chapterRepo *repository.ChapterRepo,
) *AdminService {
	return &AdminService{
		userRepo:       userRepo,
		levelRepo:      levelRepo,
		subjectRepo:    subjectRepo,
		subSubjectRepo: subSubjectRepo,
		chapterRepo:    chapterRepo,
		questionRepo:   questionRepo,
		invitationRepo: invitationRepo,
		invitationSvc:  invitationSvc,
		statsSvc:       statsSvc,
		examRecordRepo: examRecordRepo,
		examTplRepo:    examTplRepo,
	}
}

// ===================== 邀请码管理 =====================

// ListInvitationCodes 管理端邀请码列表（筛选 + 分页）
func (s *AdminService) ListInvitationCodes(req dto.ListInvitationCodeReq) ([]dto.InvitationCodeResp, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filter := repository.AdminInvitationCodeFilter{
		State:     req.State,
		Keyword:   req.Keyword,
		CreatedBy: req.CreatedBy,
	}

	total, err := s.invitationRepo.AdminCount(filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []dto.InvitationCodeResp{}, 0, nil
	}

	rows, err := s.invitationRepo.AdminList(filter, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.InvitationCodeResp, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildInvitationCodeResp(row))
	}
	return list, total, nil
}

// GenerateInvitationCodes 生成邀请码，operatorID 记录为生成者
func (s *AdminService) GenerateInvitationCodes(operatorID uint, req dto.CreateInvitationCodeReq) ([]dto.InvitationCodeResp, error) {
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > maxGeneratePerRequest {
		return nil, errors.New("单次生成数量不得超过 100 个")
	}

	codes, err := s.invitationSvc.Generate(count, operatorID, req.Note)
	if err != nil {
		return nil, err
	}

	// 生成者名称只需查一次，避免每条码都回查 users（N+1）
	creatorName := ""
	if user, uErr := s.userRepo.FindByID(operatorID); uErr == nil {
		creatorName = user.Username
	}

	resp := make([]dto.InvitationCodeResp, 0, len(codes))
	for _, code := range codes {
		resp = append(resp, dto.InvitationCodeResp{
			ID:            code.ID,
			Code:          code.Code,
			Note:          code.Note,
			Status:        code.Status,
			State:         "unused",
			CreatedBy:     operatorID,
			CreatedByName: creatorName,
			UsedBy:        nil,
			UsedByName:    "",
			UsedAt:        "",
			CreatedAt:     formatTime(code.CreatedAt),
		})
	}
	return resp, nil
}

// ===================== 用户管理 =====================

// ListUsers 管理端用户列表（关键词 + 角色 + 状态筛选，分页）
func (s *AdminService) ListUsers(req dto.ListAdminUserReq) ([]dto.AdminUserListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filter := repository.AdminUserFilter{
		Keyword: req.Keyword,
		Role:    req.Role,
		Status:  req.Status,
	}

	total, err := s.userRepo.AdminCount(filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []dto.AdminUserListItem{}, 0, nil
	}

	rows, err := s.userRepo.AdminList(filter, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.AdminUserListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildAdminUserListItem(row))
	}
	return list, total, nil
}

// CreateUser 管理员直接新建用户（不消耗邀请码）
//
// 校验顺序：用户名格式 → 密码强度 → 等级/科目合法性 → 用户名/邮箱唯一性
func (s *AdminService) CreateUser(req dto.CreateAdminUserReq) (*dto.AdminUserListItem, error) {
	username := strings.TrimSpace(req.Username)
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	if err := s.validateLevelSubject(req.LevelID, req.SubjectID); err != nil {
		return nil, err
	}
	if err := s.ensureUsernameAvailable(username, 0); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(req.Email)
	if err := s.ensureEmailAvailable(email, 0); err != nil {
		return nil, err
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "student"
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Nickname:     strings.TrimSpace(req.Nickname),
		Role:         role,
		Status:       status,
		LevelID:      req.LevelID,
		SubjectID:    req.SubjectID,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败，请稍后重试")
	}
	return s.findUserItem(user.ID)
}

// UpdateUser 管理员编辑用户资料（全字段可选）
//
// operatorID 用于阻止管理员把自己降级/禁用，避免系统失去唯一管理员
func (s *AdminService) UpdateUser(operatorID, userID uint, req dto.UpdateAdminUserReq) (*dto.AdminUserListItem, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Nickname != nil {
		user.Nickname = strings.TrimSpace(*req.Nickname)
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email != "" {
			if e := s.ensureEmailAvailable(email, userID); e != nil {
				return nil, e
			}
			user.Email = email
		}
	}
	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role == "" {
			return nil, errors.New("角色不能为空")
		}
		if userID == operatorID && user.Role == "admin" && role != "admin" {
			return nil, errors.New("不能修改自己的管理员角色")
		}
		if user.Role == "admin" && role != "admin" {
			if e := s.ensureAdminRemains(userID); e != nil {
				return nil, e
			}
		}
		user.Role = role
	}
	if req.Status != nil {
		if userID == operatorID && user.Status == 1 && *req.Status == 0 {
			return nil, errors.New("不能禁用当前登录的账号")
		}
		user.Status = *req.Status
	}
	if req.LevelID != nil {
		user.LevelID = *req.LevelID
	}
	if req.SubjectID != nil {
		user.SubjectID = *req.SubjectID
	}
	if err := s.validateLevelSubject(user.LevelID, user.SubjectID); err != nil {
		return nil, err
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("更新用户失败，请稍后重试")
	}
	return s.findUserItem(userID)
}

// UpdateUserStatus 启用 / 禁用用户
func (s *AdminService) UpdateUserStatus(operatorID, userID uint, status int) (*dto.AdminUserListItem, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if userID == operatorID && status == 0 {
		return nil, errors.New("不能禁用当前登录的账号")
	}
	if user.Status != status {
		user.Status = status
		// 启用时顺带清空登录失败锁定，避免用户刚被解禁仍无法登录
		if status == 1 {
			user.FailedAttempts = 0
			user.LockedUntil = nil
		}
		if err := s.userRepo.Update(user); err != nil {
			return nil, errors.New("更新状态失败，请稍后重试")
		}
	}
	return s.findUserItem(userID)
}

// ResetUserPassword 管理员重置指定用户密码（无需原密码）
func (s *AdminService) ResetUserPassword(userID uint, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.PasswordHash = hash
	user.FailedAttempts = 0
	user.LockedUntil = nil
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("重置密码失败，请稍后重试")
	}
	return nil
}

// DeleteUser 删除用户（数据由数据库外键级联清理）
//
// 保护规则：不能删除自己；不能删除系统内最后一个管理员
func (s *AdminService) DeleteUser(operatorID, userID uint) error {
	if userID == operatorID {
		return errors.New("不能删除当前登录的账号")
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if user.Role == "admin" {
		if err := s.ensureAdminRemains(userID); err != nil {
			return err
		}
	}
	if err := s.userRepo.Delete(userID); err != nil {
		return errors.New("删除失败，该用户可能仍有关联数据未清理")
	}
	return nil
}

// ensureAdminRemains 保证降级/删除后系统中至少还剩一个管理员。
//
// excludeID 传入待变更（降级/删除）的用户 ID。若该用户本身是管理员，需在其
// 落库变更前调用，统计值仍包含该用户，因此当 admin 总数 <=1 时即代表会是最后一个。
func (s *AdminService) ensureAdminRemains(excludeID uint) error {
	total, err := s.userRepo.CountByRole("admin")
	if err != nil {
		return errors.New("校验管理员数量失败")
	}
	if total <= 1 {
		return errors.New("系统至少需要保留一个管理员")
	}
	_ = excludeID
	return nil
}

// ensureUsernameAvailable 校验用户名唯一（excludeID>0 时排除自身，用于编辑场景）
func (s *AdminService) ensureUsernameAvailable(username string, excludeID uint) error {
	existing, _ := s.userRepo.FindByUsername(username)
	if existing != nil && existing.ID > 0 && existing.ID != excludeID {
		return errors.New("该用户名已被占用")
	}
	return nil
}

// ensureEmailAvailable 校验邮箱唯一（excludeID>0 时排除自身，用于编辑场景）
func (s *AdminService) ensureEmailAvailable(email string, excludeID uint) error {
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil && existing.ID > 0 && existing.ID != excludeID {
		return errors.New("该邮箱已被占用")
	}
	return nil
}

// validateLevelSubject 校验等级/科目存在性与归属关系（均允许为 0，表示未选择）
func (s *AdminService) validateLevelSubject(levelID, subjectID uint) error {
	if levelID == 0 && subjectID == 0 {
		return nil
	}
	if levelID > 0 {
		if _, err := s.levelRepo.FindByID(levelID); err != nil {
			return errors.New("考试等级不存在")
		}
	}
	if subjectID > 0 {
		subject, err := s.subjectRepo.FindByID(subjectID)
		if err != nil {
			return errors.New("考试科目不存在")
		}
		if levelID > 0 && subject.LevelID != levelID {
			return errors.New("所选科目不属于该考试等级")
		}
	}
	return nil
}

// findUserItem 回查单个用户的列表行并转成 dto
func (s *AdminService) findUserItem(userID uint) (*dto.AdminUserListItem, error) {
	row, err := s.userRepo.AdminFindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	item := buildAdminUserListItem(*row)
	return &item, nil
}

// buildAdminUserListItem 列表行 → dto
func buildAdminUserListItem(row repository.AdminUserRow) dto.AdminUserListItem {
	return dto.AdminUserListItem{
		ID:            row.ID,
		Username:      row.Username,
		Nickname:      row.Nickname,
		Email:         row.Email,
		EmailVerified: row.EmailVerified,
		Avatar:        row.Avatar,
		Role:          row.Role,
		Status:        row.Status,
		LevelID:       row.LevelID,
		LevelName:     row.LevelName,
		SubjectID:     row.SubjectID,
		SubjectName:   row.SubjectName,
		CreatedAt:     formatTime(row.CreatedAt),
	}
}

// ===================== 用户详情 =====================

// GetUserDetail 聚合某用户的档案 + 学习 + 答题信息。
//
// 注意：响应体刻意不含 password_hash 等凭据字段。
func (s *AdminService) GetUserDetail(userID uint) (*dto.AdminUserDetailResp, error) {
	if userID == 0 {
		return nil, errors.New("用户不存在")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	detail := &dto.AdminUserDetailResp{
		User: s.buildUserProfile(user),
	}

	// 邀请码：白名单注册的用户不消耗邀请码，此处允许为空
	if row, icErr := s.invitationRepo.AdminFindByUsedBy(userID); icErr == nil {
		ic := buildInvitationCodeResp(*row)
		detail.InviteCode = &ic
	}

	study, sErr := s.buildStudy(userID)
	if sErr != nil {
		return nil, sErr
	}
	detail.Study = study

	answer, aErr := s.buildAnswer(userID)
	if aErr != nil {
		return nil, aErr
	}
	detail.Answer = answer

	return detail, nil
}

// buildUserProfile 拼装用户档案（补 level/subject 名称）
func (s *AdminService) buildUserProfile(user *model.User) dto.AdminUserProfile {
	profile := dto.AdminUserProfile{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Avatar:        user.Avatar,
		Role:          user.Role,
		Status:        user.Status,
		LevelID:       user.LevelID,
		SubjectID:     user.SubjectID,
		Difficulty:    user.Difficulty,
		CreatedAt:     formatTime(user.CreatedAt),
	}
	if user.LevelID > 0 {
		if level, err := s.levelRepo.FindByID(user.LevelID); err == nil {
			profile.LevelName = level.Name
		}
	}
	if user.SubjectID > 0 {
		if subject, err := s.subjectRepo.FindByID(user.SubjectID); err == nil {
			profile.SubjectName = subject.Name
		}
	}
	return profile
}

// buildStudy 学习信息：概览 + 近 N 日活跃 + 科目/章节掌握度
func (s *AdminService) buildStudy(userID uint) (dto.AdminUserStudyResp, error) {
	resp := dto.AdminUserStudyResp{
		Overview: dto.StatsOverviewResp{},
		Recent:   []dto.DailyStatsResp{},
		Subjects: []dto.SubjectProgressResp{},
		Chapters: []dto.ChapterProgressResp{},
	}

	overview, err := s.statsSvc.GetOverview(userID)
	if err != nil {
		return resp, err
	}
	resp.Overview = *overview

	recent, err := s.statsSvc.GetDailyStats(userID, adminRecentStudyDays)
	if err != nil {
		return resp, err
	}
	resp.Recent = recent

	subjects, err := s.statsSvc.GetSubjectProgress(userID)
	if err != nil {
		return resp, err
	}
	resp.Subjects = subjects

	chapters, err := s.statsSvc.GetChapterProgress(userID)
	if err != nil {
		return resp, err
	}
	resp.Chapters = chapters

	return resp, nil
}

// buildAnswer 答题信息：考试次数/平均分 + 最近若干条考试记录
func (s *AdminService) buildAnswer(userID uint) (dto.AdminUserAnswerResp, error) {
	resp := dto.AdminUserAnswerResp{Records: []dto.ExamRecordResp{}}

	overview, err := s.statsSvc.GetOverview(userID)
	if err != nil {
		return resp, err
	}
	resp.TotalExams = overview.TotalExams
	resp.AvgExamScore = overview.AvgExamScore

	records, err := s.examRecordRepo.FindByUser(userID)
	if err != nil {
		return resp, err
	}
	if len(records) == 0 {
		return resp, nil
	}

	if len(records) > adminExamRecordLimit {
		records = records[:adminExamRecordLimit]
	}

	list, err := s.buildExamRecordResps(records)
	if err != nil {
		return resp, err
	}
	resp.Records = list
	return resp, nil
}

// buildExamRecordResps 把考试记录 model 转成 dto（批量查模板，消除 N+1）
func (s *AdminService) buildExamRecordResps(records []model.ExamRecord) ([]dto.ExamRecordResp, error) {
	templateIDs := make([]uint, 0, len(records))
	aiRecordIDs := make([]uint, 0)
	for _, r := range records {
		if r.TemplateID == 0 {
			aiRecordIDs = append(aiRecordIDs, r.ID)
		} else {
			templateIDs = append(templateIDs, r.TemplateID)
		}
	}

	templates, err := s.examTplRepo.FindByIDs(templateIDs)
	if err != nil {
		return nil, err
	}
	aiSubjects, err := s.examRecordRepo.GetSubjectIDsByRecords(aiRecordIDs)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ExamRecordResp, 0, len(records))
	for _, r := range records {
		templateName := ""
		subjectID := uint(0)
		if r.TemplateID == 0 {
			templateName = "AI 智能组卷"
			subjectID = aiSubjects[r.ID]
		} else if t, ok := templates[r.TemplateID]; ok {
			templateName = t.Name
			subjectID = t.SubjectID
		}
		resp = append(resp, dto.ExamRecordResp{
			ID:           r.ID,
			TemplateID:   r.TemplateID,
			TemplateName: templateName,
			SubjectID:    subjectID,
			Score:        r.Score,
			TotalScore:   r.TotalScore,
			Duration:     r.Duration,
			CorrectCount: r.CorrectCount,
			Accuracy:     r.Accuracy,
			Status:       r.Status,
			StartedAt:    formatTime(r.StartedAt),
			FinishedAt:   formatTime(r.FinishedAt),
			CreatedAt:    formatTime(r.CreatedAt),
		})
	}
	return resp, nil
}

// ===================== 题库管理 =====================

// 合法题型 / 难度白名单，与 model 中的题型常量、前端 utils/question.ts 保持一致
var (
	adminQuestionTypes = map[string]bool{
		model.TypeSingle:        true,
		model.TypeMulti:         true,
		model.TypeJudge:         true,
		model.TypeFill:          true,
		model.TypeShort:         true,
		model.TypeComprehensive: true,
		model.TypeEssay:         true,
		model.TypeCaseStudy:     true,
		model.TypeMultiBlank:    true,
	}
	adminDifficulties = map[string]bool{"easy": true, "medium": true, "hard": true}
)

// ListQuestions 管理端题目列表（筛选 + 分页）
func (s *AdminService) ListQuestions(req dto.AdminQuestionListReq) ([]dto.AdminQuestionResp, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filter := repository.AdminQuestionFilter{
		SubjectID:    req.SubjectID,
		SubSubjectID: req.SubSubjectID,
		ChapterID:    req.ChapterID,
		Type:         req.Type,
		Difficulty:   req.Difficulty,
		Status:       req.Status,
		Year:         req.Year,
		Keyword:      req.Keyword,
	}

	total, err := s.questionRepo.AdminCountQuestions(filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []dto.AdminQuestionResp{}, 0, nil
	}

	rows, err := s.questionRepo.AdminListQuestions(filter, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]dto.AdminQuestionResp, 0, len(rows))
	for _, row := range rows {
		list = append(list, toAdminQuestionResp(row))
	}
	return list, total, nil
}

// GetQuestion 管理端题目详情：案例分析大题目会带回小题目
func (s *AdminService) GetQuestion(id uint) (*dto.AdminQuestionResp, error) {
	q, err := s.questionRepo.AdminFindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	resp := s.buildAdminQuestionResp(*q)

	// 案例分析 / 综合题可能存在小题目，一并带回便于管理员核对
	if children, cErr := s.questionRepo.AdminFindChildren(id); cErr == nil && len(children) > 0 {
		resp.Children = make([]dto.AdminQuestionChildResp, 0, len(children))
		for _, c := range children {
			resp.Children = append(resp.Children, dto.AdminQuestionChildResp{
				ID:       c.ID,
				Type:     c.Type,
				Content:  c.Content,
				Answer:   c.Answer,
				Analysis: c.Analysis,
			})
		}
	}
	return &resp, nil
}

// CreateQuestion 新增题目，写入后回查返回完整响应（含科目/章节名称）
func (s *AdminService) CreateQuestion(req dto.AdminQuestionUpsertReq) (*dto.AdminQuestionResp, error) {
	q, err := s.buildQuestionModel(req)
	if err != nil {
		return nil, err
	}
	// 新增默认启用；显式传 0 时为停用
	q.Status = 1
	if req.Status != nil && *req.Status == 0 {
		q.Status = 0
	}

	if err := s.questionRepo.Create(q); err != nil {
		return nil, err
	}

	resp := s.buildAdminQuestionResp(*q)
	return &resp, nil
}

// UpdateQuestion 局部更新题目：字段白名单更新，未传的字段不覆盖
func (s *AdminService) UpdateQuestion(id uint, req dto.AdminQuestionUpsertReq) (*dto.AdminQuestionResp, error) {
	existing, err := s.questionRepo.AdminFindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	// 合并后整体校验：允许只传部分字段（如仅改状态），用现有值补齐后再判合法性
	merged := mergeQuestionReq(*existing, req)
	if _, err := s.buildQuestionModel(merged); err != nil {
		return nil, err
	}

	columns := map[string]any{}
	if req.SubjectID > 0 {
		columns["subject_id"] = req.SubjectID
	}
	columns["sub_subject_id"] = req.SubSubjectID
	columns["chapter_id"] = req.ChapterID
	if req.Type != "" {
		columns["type"] = req.Type
	}
	if req.Difficulty != "" {
		columns["difficulty"] = req.Difficulty
	}
	columns["content"] = req.Content
	columns["case_material"] = req.CaseMaterial
	columns["options"] = req.Options
	columns["blank_options"] = req.BlankOptions
	columns["answer"] = req.Answer
	columns["analysis"] = req.Analysis
	columns["year"] = req.Year
	columns["source"] = req.Source
	if req.Status != nil {
		columns["status"] = *req.Status
	}

	if err := s.questionRepo.UpdateColumns(id, columns); err != nil {
		return nil, err
	}

	updated, err := s.questionRepo.AdminFindByID(id)
	if err != nil {
		return nil, err
	}
	resp := s.buildAdminQuestionResp(*updated)
	return &resp, nil
}

// DeleteQuestion 删除题目：软删除（status=0），保留练习/考试记录的外键引用
func (s *AdminService) DeleteQuestion(id uint) error {
	if _, err := s.questionRepo.AdminFindByID(id); err != nil {
		return ErrNotFound
	}
	return s.questionRepo.SetStatus([]uint{id}, 0)
}

// BatchSetQuestionStatus 批量启用 / 停用题目
func (s *AdminService) BatchSetQuestionStatus(ids []uint, status int) error {
	if len(ids) == 0 {
		return errors.New("请至少选择一道题目")
	}
	if len(ids) > maxBatchQuestionIDs {
		return fmt.Errorf("单次操作不得超过 %d 道题目", maxBatchQuestionIDs)
	}
	if status != 0 && status != 1 {
		return errors.New("状态值非法")
	}
	return s.questionRepo.SetStatus(ids, status)
}

// QuestionStats 题库概览：总量 / 启用 / 停用 / 各题型分布
func (s *AdminService) QuestionStats() (*dto.AdminQuestionStatsResp, error) {
	total, err := s.questionRepo.CountAll()
	if err != nil {
		return nil, err
	}
	enabled, err := s.questionRepo.CountByStatus(1)
	if err != nil {
		return nil, err
	}
	rows, err := s.questionRepo.AdminTypeStats()
	if err != nil {
		return nil, err
	}

	byType := make([]dto.AdminQuestionTypeStat, 0, len(rows))
	for _, r := range rows {
		byType = append(byType, dto.AdminQuestionTypeStat{Type: r.Type, Count: r.Count})
	}

	return &dto.AdminQuestionStatsResp{
		Total:    total,
		Enabled:  enabled,
		Disabled: total - enabled,
		ByType:   byType,
	}, nil
}

// buildQuestionModel 校验请求并构造题目 model
//
// 校验项：科目、题型、难度、题干、答案；选择题额外校验选项 JSON 与答案是否落在选项内。
func (s *AdminService) buildQuestionModel(req dto.AdminQuestionUpsertReq) (*model.Question, error) {
	if req.SubjectID == 0 {
		return nil, errors.New("请选择所属科目")
	}
	if !adminQuestionTypes[req.Type] {
		return nil, errors.New("题型不合法")
	}
	if !adminDifficulties[req.Difficulty] {
		return nil, errors.New("难度不合法，可选：简单 / 中等 / 困难")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("题干不能为空")
	}
	if req.Year < 0 {
		return nil, errors.New("年份不能为负数")
	}

	answer := strings.TrimSpace(req.Answer)
	options := strings.TrimSpace(req.Options)
	blankOptions := strings.TrimSpace(req.BlankOptions)

	switch req.Type {
	case model.TypeSingle, model.TypeMulti, model.TypeJudge:
		ids, err := parseOptionIDs(options)
		if err != nil {
			return nil, errors.New("选项格式错误，需为合法 JSON 数组")
		}
		if len(ids) < 2 {
			return nil, errors.New("选择题至少需要两个选项")
		}
		if req.Type == model.TypeMulti {
			parts := splitMultiAnswer(answer)
			if len(parts) == 0 {
				return nil, errors.New("多选题答案不能为空")
			}
			for _, p := range parts {
				if !containsID(ids, p) {
					return nil, errors.New("多选题答案必须来自选项")
				}
			}
			// 多选题按选项顺序规范化（A,C），保证与判分逻辑的字符串比较一致
			answer = normalizeMultiAnswer(parts, ids)
		} else {
			if !containsID(ids, answer) {
				return nil, errors.New("答案必须来自选项")
			}
		}
	case model.TypeMultiBlank:
		blanks, err := parseBlankAnswers(answer)
		if err != nil || len(blanks) == 0 {
			return nil, errors.New("多空题答案需为非空 JSON 数组，如 [\"A\",\"C\"]")
		}
		if blankOptions != "" {
			groups, err := parseBlankOptionGroups(blankOptions)
			if err != nil {
				return nil, errors.New("多空题选项格式错误")
			}
			if len(groups) != len(blanks) {
				return nil, errors.New("多空题的空数量与答案数量不一致")
			}
			for i, g := range groups {
				if !containsID(g, blanks[i]) {
					return nil, fmt.Errorf("第 %d 空的答案不在该空选项中", i+1)
				}
			}
		}
	default:
		// 主观题（论文 / 案例分析）允许参考答案为空，仅作评分参考
		if req.Type != model.TypeEssay && req.Type != model.TypeCaseStudy && answer == "" {
			return nil, errors.New("答案不能为空")
		}
	}

	return &model.Question{
		SubjectID:    req.SubjectID,
		SubSubjectID: req.SubSubjectID,
		ChapterID:    req.ChapterID,
		Type:         req.Type,
		Difficulty:   req.Difficulty,
		Content:      content,
		CaseMaterial: req.CaseMaterial,
		Options:      options,
		BlankOptions: blankOptions,
		Answer:       answer,
		Analysis:     req.Analysis,
		Year:         req.Year,
		Source:       req.Source,
	}, nil
}

// mergeQuestionReq 用现有题目值补齐更新请求，便于整体校验
func mergeQuestionReq(q model.Question, req dto.AdminQuestionUpsertReq) dto.AdminQuestionUpsertReq {
	merged := req
	if merged.SubjectID == 0 {
		merged.SubjectID = q.SubjectID
	}
	if merged.SubSubjectID == 0 {
		merged.SubSubjectID = q.SubSubjectID
	}
	if merged.ChapterID == 0 {
		merged.ChapterID = q.ChapterID
	}
	if merged.Type == "" {
		merged.Type = q.Type
	}
	if merged.Difficulty == "" {
		merged.Difficulty = q.Difficulty
	}
	if merged.Content == "" {
		merged.Content = q.Content
	}
	if merged.CaseMaterial == "" {
		merged.CaseMaterial = q.CaseMaterial
	}
	if merged.Options == "" {
		merged.Options = q.Options
	}
	if merged.BlankOptions == "" {
		merged.BlankOptions = q.BlankOptions
	}
	if merged.Answer == "" {
		merged.Answer = q.Answer
	}
	if merged.Analysis == "" {
		merged.Analysis = q.Analysis
	}
	if merged.Year == 0 {
		merged.Year = q.Year
	}
	if merged.Source == "" {
		merged.Source = q.Source
	}
	return merged
}

// buildAdminQuestionResp 由 model 构造管理端响应（补科目/子科目/章节名称）
func (s *AdminService) buildAdminQuestionResp(q model.Question) dto.AdminQuestionResp {
	resp := dto.AdminQuestionResp{
		ID:           q.ID,
		SubjectID:    q.SubjectID,
		SubSubjectID: q.SubSubjectID,
		ChapterID:    q.ChapterID,
		ParentID:     q.ParentID,
		Type:         q.Type,
		Difficulty:   q.Difficulty,
		Content:      q.Content,
		CaseMaterial: q.CaseMaterial,
		Options:      q.Options,
		BlankOptions: q.BlankOptions,
		Answer:       q.Answer,
		Analysis:     q.Analysis,
		Year:         q.Year,
		Source:       q.Source,
		Status:       q.Status,
		CreatedAt:    formatTime(q.CreatedAt),
		UpdatedAt:    formatTime(q.UpdatedAt),
	}
	resp.SubjectName, resp.SubSubjectName, resp.ChapterName = s.loadQuestionNames(q)
	return resp
}

// loadQuestionNames 依次回查科目 / 子科目 / 章节名称；查不到时回落空串，不阻断主流程
func (s *AdminService) loadQuestionNames(q model.Question) (subject, subSubject, chapter string) {
	if q.SubjectID > 0 {
		if v, err := s.subjectRepo.FindByID(q.SubjectID); err == nil {
			subject = v.Name
		}
	}
	if q.SubSubjectID > 0 && s.subSubjectRepo != nil {
		if v, err := s.subSubjectRepo.FindByID(q.SubSubjectID); err == nil {
			subSubject = v.Name
		}
	}
	if q.ChapterID > 0 && s.chapterRepo != nil {
		if v, err := s.chapterRepo.FindByID(q.ChapterID); err == nil {
			chapter = v.Name
		}
	}
	return subject, subSubject, chapter
}

// toAdminQuestionResp 列表行 → dto
func toAdminQuestionResp(row repository.AdminQuestionRow) dto.AdminQuestionResp {
	return dto.AdminQuestionResp{
		ID:             row.ID,
		SubjectID:      row.SubjectID,
		SubjectName:    row.SubjectName,
		SubSubjectID:   row.SubSubjectID,
		SubSubjectName: row.SubSubjectName,
		ChapterID:      row.ChapterID,
		ChapterName:    row.ChapterName,
		ParentID:       row.ParentID,
		Type:           row.Type,
		Difficulty:     row.Difficulty,
		Content:        row.Content,
		CaseMaterial:   row.CaseMaterial,
		Options:        row.Options,
		BlankOptions:   row.BlankOptions,
		Answer:         row.Answer,
		Analysis:       row.Analysis,
		Year:           row.Year,
		Source:         row.Source,
		Status:         row.Status,
		CreatedAt:      formatTime(row.CreatedAt),
		UpdatedAt:      formatTime(row.UpdatedAt),
	}
}

// ===================== 题库校验工具 =====================

// optionItem 选项 JSON 的两种形态之一：[{"id":"A","content":"..."}]
type optionItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// parseOptionIDs 解析选项 JSON，返回选项标识集合（A/B/C…）
//
// 兼容三种历史格式：
//  1. [{"id":"A","content":"..."}]（AI 生成题与新建题标准格式）
//  2. {"A":"...","B":"..."}
//  3. ["选项一","选项二"]（按 A/B/C 顺序补标识）
func parseOptionIDs(options string) ([]string, error) {
	if strings.TrimSpace(options) == "" {
		return nil, errors.New("选项为空")
	}
	raw := []byte(options)

	// 先尝试对象形态，再尝试数组形态
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err == nil {
		ids := make([]string, 0, len(obj))
		for k := range obj {
			ids = append(ids, k)
		}
		sort.Strings(ids)
		return ids, nil
	}

	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(arr))
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			if id, ok := v["id"].(string); ok && id != "" {
				ids = append(ids, id)
			} else {
				ids = append(ids, string(rune('A'+i)))
			}
		default:
			ids = append(ids, string(rune('A'+i)))
		}
	}
	return ids, nil
}

// parseBlankOptionGroups 解析多空题选项，返回每空的选项标识集合
func parseBlankOptionGroups(blankOptions string) ([][]string, error) {
	var groups []struct {
		BlankIndex int          `json:"blank_index"`
		Options    []optionItem `json:"options"`
	}
	if err := json.Unmarshal([]byte(blankOptions), &groups); err != nil {
		return nil, err
	}
	out := make([][]string, 0, len(groups))
	for _, g := range groups {
		ids := make([]string, 0, len(g.Options))
		for _, o := range g.Options {
			ids = append(ids, o.ID)
		}
		out = append(out, ids)
	}
	return out, nil
}

// parseBlankAnswers 解析多空题答案 JSON 数组，如 ["A","C"]
func parseBlankAnswers(answer string) ([]string, error) {
	if !strings.HasPrefix(strings.TrimSpace(answer), "[") {
		return nil, errors.New("答案非 JSON 数组")
	}
	var arr []string
	if err := json.Unmarshal([]byte(answer), &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// splitMultiAnswer 拆分多选题答案（支持 "A,C" 与 "A，C" 两种分隔符）
func splitMultiAnswer(answer string) []string {
	replaced := strings.NewReplacer("，", ",", " ", "", "、", ",").Replace(answer)
	parts := strings.Split(replaced, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, strings.ToUpper(p))
		}
	}
	return out
}

// normalizeMultiAnswer 按选项顺序排序多选题答案，保证存储格式统一（如 "A,C"）
func normalizeMultiAnswer(parts []string, optionIDs []string) string {
	order := make(map[string]int, len(optionIDs))
	for i, id := range optionIDs {
		order[id] = i
	}
	sort.Slice(parts, func(i, j int) bool { return order[parts[i]] < order[parts[j]] })
	return strings.Join(parts, ",")
}

// containsID 判断答案标识是否在选项集合中（忽略大小写与空格）
func containsID(ids []string, answer string) bool {
	target := strings.ToUpper(strings.TrimSpace(answer))
	for _, id := range ids {
		if strings.EqualFold(strings.TrimSpace(id), target) {
			return true
		}
	}
	return false
}

// ===================== 内部工具 =====================

// buildInvitationCodeResp 列表行 → dto，并统一推导 state 供前端直接渲染
func buildInvitationCodeResp(row repository.AdminInvitationCodeRow) dto.InvitationCodeResp {
	state := "unused"
	switch {
	case row.UsedBy != nil:
		state = "used"
	case row.Status != 1:
		state = "disabled"
	}
	return dto.InvitationCodeResp{
		ID:            row.ID,
		Code:          row.Code,
		Note:          row.Note,
		Status:        row.Status,
		State:         state,
		CreatedBy:     row.CreatedBy,
		CreatedByName: row.CreatedByName,
		UsedBy:        row.UsedBy,
		UsedByName:    row.UsedByName,
		UsedAt:        formatTimePtr(row.UsedAt),
		CreatedAt:     formatTime(row.CreatedAt),
	}
}

// ===================== 科目管理 =====================

// ListSubjects 管理端科目列表（按等级筛选 + 关键词，含停用科目，附引用统计）
func (s *AdminService) ListSubjects(req dto.ListAdminSubjectReq) ([]dto.AdminSubjectResp, error) {
	subjects, err := s.subjectRepo.AdminFindAll(req.LevelID, strings.TrimSpace(req.Keyword))
	if err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		return []dto.AdminSubjectResp{}, nil
	}

	resp := make([]dto.AdminSubjectResp, 0, len(subjects))
	for _, subject := range subjects {
		levelName := ""
		if subject.LevelID > 0 {
			if level, lErr := s.levelRepo.FindByID(subject.LevelID); lErr == nil {
				levelName = level.Name
			}
		}
		subCount, _ := s.subjectRepo.CountSubSubjects(subject.ID)
		questionCount, _ := s.subjectRepo.CountQuestions(subject.ID)
		resp = append(resp, dto.AdminSubjectResp{
			ID:              subject.ID,
			LevelID:         subject.LevelID,
			LevelName:       levelName,
			Name:            subject.Name,
			ShortName:       subject.ShortName,
			Description:     subject.Description,
			Icon:            subject.Icon,
			SortOrder:       subject.SortOrder,
			Status:          subject.Status,
			SubSubjectCount: subCount,
			QuestionCount:   questionCount,
		})
	}
	return resp, nil
}

// CreateSubject 新增科目，默认启用；显式传 status=0 时停用
func (s *AdminService) CreateSubject(req dto.AdminSubjectUpsertReq) (*dto.AdminSubjectResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("科目名称不能为空")
	}
	if req.LevelID == 0 {
		return nil, errors.New("请选择所属等级")
	}
	if _, err := s.levelRepo.FindByID(req.LevelID); err != nil {
		return nil, errors.New("考试等级不存在")
	}
	if n, err := s.subjectRepo.CountByName(req.LevelID, name, 0); err == nil && n > 0 {
		return nil, errors.New("该等级下已存在同名科目")
	}

	subject := &model.Subject{
		LevelID:     req.LevelID,
		Name:        name,
		ShortName:   strings.TrimSpace(req.ShortName),
		Description: req.Description,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	if req.Status != nil && *req.Status == 0 {
		subject.Status = 0
	}
	if err := s.subjectRepo.Create(subject); err != nil {
		return nil, err
	}
	return s.getSubjectResp(subject.ID)
}

// UpdateSubject 更新科目基本信息（不含所属等级的改名场景，等级不可变更）
func (s *AdminService) UpdateSubject(id uint, req dto.AdminSubjectUpsertReq) (*dto.AdminSubjectResp, error) {
	existing, err := s.subjectRepo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	levelID := existing.LevelID
	if req.LevelID > 0 && req.LevelID != levelID {
		// 允许迁级，但需校验新等级存在且该等级下不重名
		if _, lErr := s.levelRepo.FindByID(req.LevelID); lErr != nil {
			return nil, errors.New("考试等级不存在")
		}
		if n, cErr := s.subjectRepo.CountByName(req.LevelID, name, id); cErr == nil && n > 0 {
			return nil, errors.New("目标等级下已存在同名科目")
		}
		levelID = req.LevelID
	} else {
		// 改名需在本等级内查重（排除自身）
		if n, cErr := s.subjectRepo.CountByName(levelID, name, id); cErr == nil && n > 0 {
			return nil, errors.New("该等级下已存在同名科目")
		}
	}

	existing.LevelID = levelID
	existing.Name = name
	existing.ShortName = strings.TrimSpace(req.ShortName)
	existing.Description = req.Description
	existing.Icon = req.Icon
	existing.SortOrder = req.SortOrder

	if err := s.subjectRepo.Update(existing); err != nil {
		return nil, err
	}
	return s.getSubjectResp(id)
}

// SetSubjectStatus 启用 / 停用科目（停用即软删：前端不再展示，历史数据保留）
func (s *AdminService) SetSubjectStatus(id uint, status int) error {
	if status != 0 && status != 1 {
		return errors.New("状态值非法")
	}
	if _, err := s.subjectRepo.FindByID(id); err != nil {
		return ErrNotFound
	}
	return s.subjectRepo.UpdateStatus(id, status)
}

// DeleteSubject 删除科目
//
// 保护规则：科目下仍存在子科目或题目时禁止删除（这些数据有外键引用，级联删除代价大），
// 需先清理子科目与题目后重试，或改为「停用」。
func (s *AdminService) DeleteSubject(id uint) error {
	subject, err := s.subjectRepo.FindByID(id)
	if err != nil {
		return ErrNotFound
	}
	subCount, err := s.subjectRepo.CountSubSubjects(id)
	if err != nil {
		return err
	}
	if subCount > 0 {
		return fmt.Errorf("科目「%s」下仍有 %d 个子科目，请先删除子科目或改为停用", subject.Name, subCount)
	}
	questionCount, err := s.subjectRepo.CountQuestions(id)
	if err != nil {
		return err
	}
	if questionCount > 0 {
		return fmt.Errorf("科目「%s」下仍有 %d 道题目，请先删除题目或改为停用", subject.Name, questionCount)
	}
	return s.subjectRepo.Delete(id)
}

// getSubjectResp 按 id 回查并构造单个管理端科目响应
func (s *AdminService) getSubjectResp(id uint) (*dto.AdminSubjectResp, error) {
	list, err := s.ListSubjects(dto.ListAdminSubjectReq{})
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, ErrNotFound
}
