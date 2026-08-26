package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/pkg/llm"
)

const maxKnowledgeDescriptionLen = 1000

type KnowledgePointService struct {
	repo         *repository.KnowledgePointRepo
	questionRepo *repository.QuestionRepo
	subjectRepo  *repository.SubjectRepo
}

func NewKnowledgePointService(
	repo *repository.KnowledgePointRepo,
	questionRepo *repository.QuestionRepo,
	subjectRepo *repository.SubjectRepo,
) *KnowledgePointService {
	return &KnowledgePointService{
		repo:         repo,
		questionRepo: questionRepo,
		subjectRepo:  subjectRepo,
	}
}

// ExtractKnowledgePoints AI 提取知识点（不入库）
func (s *KnowledgePointService) ExtractKnowledgePoints(req dto.ExtractKnowledgePointsReq) (*dto.ExtractKnowledgePointsResp, error) {
	if req.ApiConfig.ApiKey == "" {
		return nil, fmt.Errorf("请先配置 AI API Key")
	}
	if req.ApiConfig.BaseURL == "" {
		return nil, fmt.Errorf("请先配置 AI API 地址")
	}
	if req.ApiConfig.Model == "" {
		return nil, fmt.Errorf("请先配置 AI 模型")
	}

	prompt := buildExtractKnowledgePointsPrompt(req)

	resp, err := llm.Chat(context.Background(), req.ApiConfig.Provider, req.ApiConfig.BaseURL, req.ApiConfig.ApiKey, llm.ChatRequest{
		Model:       req.ApiConfig.Model,
		Messages:    []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.4,
		MaxTokens:   2048,
		Timeout:     120 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("AI 调用失败: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI 返回为空")
	}
	content := resp.Choices[0].Message.Content

	jsonStr := extractJSONArray(content)
	var items []dto.KnowledgePointItem
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, fmt.Errorf("AI 返回无法解析: %w", err)
	}

	cleaned := make([]dto.KnowledgePointItem, 0, len(items))
	for _, it := range items {
		name := strings.TrimSpace(it.Name)
		desc := strings.TrimSpace(it.Description)
		if name == "" {
			continue
		}
		if len([]rune(name)) > 60 {
			name = string([]rune(name)[:60])
		}
		if len([]rune(desc)) > maxKnowledgeDescriptionLen {
			desc = string([]rune(desc)[:maxKnowledgeDescriptionLen])
		}
		cleaned = append(cleaned, dto.KnowledgePointItem{Name: name, Description: desc})
		if len(cleaned) >= 1 {
			break
		}
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("AI 未提取到有效知识点")
	}
	return &dto.ExtractKnowledgePointsResp{Points: cleaned}, nil
}

// AddResult 加入单条返回结构
type AddResult struct {
	Status   string
	UserKPID uint
	KP       *model.KnowledgePoint
	Existing *model.UserKnowledgePoint
}

// Add 加入单条知识点（带冲突检测）
func (s *KnowledgePointService) Add(userID uint, req dto.AddKnowledgePointReq) (*AddResult, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("知识点名称不能为空")
	}
	desc := strings.TrimSpace(req.Description)
	if len([]rune(desc)) > maxKnowledgeDescriptionLen {
		desc = string([]rune(desc)[:maxKnowledgeDescriptionLen])
	}

	var kp *model.KnowledgePoint
	existing, err := s.repo.FindByName(name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		kp = &model.KnowledgePoint{
			Name:        name,
			SubjectID:   req.SubjectID,
			Description: desc,
			Source:      "ai",
		}
		if err := s.repo.CreateKnowledgePoint(kp); err != nil {
			return nil, err
		}
	} else {
		kp = existing
	}

	ukp, err := s.repo.FindUserKP(userID, kp.ID)
	if err == nil {
		return &AddResult{
			Status:   "duplicate",
			UserKPID: ukp.ID,
			KP:       kp,
			Existing: ukp,
		}, ErrConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	ukp = &model.UserKnowledgePoint{
		UserID:           userID,
		KnowledgePointID: kp.ID,
		SourceQuestionID: req.SourceQuestionID,
	}
	if err := s.repo.CreateUserKP(ukp); err != nil {
		return nil, err
	}
	return &AddResult{
		Status:   "added",
		UserKPID: ukp.ID,
		KP:       kp,
	}, nil
}

// BatchAdd 批量加入
func (s *KnowledgePointService) BatchAdd(userID uint, req dto.BatchAddKnowledgePointReq) (*dto.BatchAddKnowledgePointResp, error) {
	added := make([]dto.AddKnowledgePointItemResp, 0)
	duplicates := make([]dto.AddKnowledgePointItemResp, 0)

	for _, p := range req.Points {
		r, err := s.Add(userID, dto.AddKnowledgePointReq{
			Name:             p.Name,
			Description:      p.Description,
			SubjectID:        p.SubjectID,
			SourceQuestionID: req.SourceQuestionID,
		})
		if err != nil {
			if errors.Is(err, ErrConflict) && r != nil {
				duplicates = append(duplicates, buildAddResp(r, "duplicate"))
				continue
			}
			return nil, err
		}
		added = append(added, buildAddResp(r, "added"))
	}
	return &dto.BatchAddKnowledgePointResp{Added: added, Duplicates: duplicates}, nil
}

// ResolveDuplicate 解决冲突
func (s *KnowledgePointService) ResolveDuplicate(userID uint, existingUserKPID uint, req dto.ResolveDuplicateReq) (*dto.AddKnowledgePointItemResp, error) {
	ukp, err := s.repo.FindUserKP(userID, existingUserKPID)
	if err != nil {
		return nil, fmt.Errorf("%w: 用户知识点不存在", ErrNotFound)
	}
	if ukp.KnowledgePointID == 0 {
		return nil, fmt.Errorf("关联知识点异常")
	}

	switch req.Action {
	case "keep_old":
		return &dto.AddKnowledgePointItemResp{
			UserKPID: ukp.ID,
			Status:   "kept",
		}, nil
	case "replace":
		newDesc := strings.TrimSpace(req.NewDesc)
		if len([]rune(newDesc)) > maxKnowledgeDescriptionLen {
			newDesc = string([]rune(newDesc)[:maxKnowledgeDescriptionLen])
		}
		if err := s.repo.UpdateKnowledgePoint(ukp.KnowledgePointID, map[string]interface{}{
			"description": newDesc,
		}); err != nil {
			return nil, err
		}
		return &dto.AddKnowledgePointItemResp{
			UserKPID: ukp.ID,
			Status:   "replaced",
		}, nil
	case "add_new":
		baseName := strings.TrimSpace(req.NewName)
		if baseName == "" {
			return nil, fmt.Errorf("新名称不能为空")
		}
		newName := baseName + " (副本 " + time.Now().Format("2006-01-02") + ")"
		newDesc := strings.TrimSpace(req.NewDesc)
		if len([]rune(newDesc)) > maxKnowledgeDescriptionLen {
			newDesc = string([]rune(newDesc)[:maxKnowledgeDescriptionLen])
		}
		newKP := &model.KnowledgePoint{
			Name:        newName,
			SubjectID:   req.SubjectID,
			Description: newDesc,
			Source:      "ai",
		}
		if err := s.repo.CreateKnowledgePoint(newKP); err != nil {
			return nil, err
		}
		newUKP := &model.UserKnowledgePoint{
			UserID:           userID,
			KnowledgePointID: newKP.ID,
		}
		if err := s.repo.CreateUserKP(newUKP); err != nil {
			return nil, err
		}
		kp := knowledgePointToResp(newKP)
		return &dto.AddKnowledgePointItemResp{
			UserKPID:  newUKP.ID,
			Status:    "added",
			Knowledge: &kp,
		}, nil
	default:
		return nil, fmt.Errorf("不支持的 action: %s", req.Action)
	}
}

// UpdateUserKP 更新重点/掌握
func (s *KnowledgePointService) UpdateUserKP(userID uint, id uint, req dto.UpdateUserKnowledgePointReq) error {
	fields := map[string]interface{}{}
	if req.IsImportant != nil {
		fields["is_important"] = boolToInt(*req.IsImportant)
	}
	if req.IsMastered != nil {
		fields["is_mastered"] = boolToInt(*req.IsMastered)
		if *req.IsMastered {
			fields["mastered_at"] = time.Now()
		} else {
			fields["mastered_at"] = nil
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateUserKPFields(id, userID, fields)
}

// Remove 删除
func (s *KnowledgePointService) Remove(userID uint, id uint) error {
	return s.repo.DeleteUserKP(userID, id)
}

// List 列表
func (s *KnowledgePointService) List(userID uint, req dto.ListKnowledgePointReq) ([]dto.UserKnowledgePointResp, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	list, total, err := s.repo.ListUserKP(repository.ListUserKPParams{
		UserID:    userID,
		Filter:    req.Filter,
		SubjectID: req.SubjectID,
		Keyword:   req.Keyword,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	if len(list) == 0 {
		return []dto.UserKnowledgePointResp{}, total, nil
	}
	ids := make([]uint, 0, len(list))
	for _, v := range list {
		ids = append(ids, v.KnowledgePointID)
	}
	kps, err := s.repo.FindKnowledgePoints(ids)
	if err != nil {
		return nil, 0, err
	}
	kpMap := make(map[uint]model.KnowledgePoint, len(kps))
	for _, kp := range kps {
		kpMap[kp.ID] = kp
	}
	resp := make([]dto.UserKnowledgePointResp, 0, len(list))
	for _, v := range list {
		kp := kpMap[v.KnowledgePointID]
		resp = append(resp, dto.UserKnowledgePointResp{
			ID:               v.ID,
			UserID:           v.UserID,
			KnowledgePointID: v.KnowledgePointID,
			SourceQuestionID: v.SourceQuestionID,
			IsImportant:      v.IsImportant,
			IsMastered:       v.IsMastered,
			MasteredAt:       formatTimePtr(v.MasteredAt),
			Note:             v.Note,
			CreatedAt:        v.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        v.UpdatedAt.Format("2006-01-02 15:04:05"),
			Knowledge:        knowledgePointToResp(&kp),
		})
	}
	return resp, total, nil
}

// Stats 统计
func (s *KnowledgePointService) Stats(userID uint) (*dto.KnowledgePointStatsResp, error) {
	total, important, mastered, unmastered, recent, err := s.repo.StatsByUser(userID)
	if err != nil {
		return nil, err
	}
	return &dto.KnowledgePointStatsResp{
		Total:      total,
		Important:  important,
		Mastered:   mastered,
		Unmastered: unmastered,
		RecentWeek: recent,
	}, nil
}

// ===== helpers =====

func buildExtractKnowledgePointsPrompt(req dto.ExtractKnowledgePointsReq) string {
	var sb strings.Builder
	sb.WriteString(extractKnowledgePointsPrompt)
	sb.WriteString("\n\n【本次题目】\n")
	if req.QuestionType != "" {
		sb.WriteString("题型：")
		sb.WriteString(typeLabelCN(req.QuestionType))
		sb.WriteString("\n")
	}
	if req.QuestionContent != "" {
		sb.WriteString("题干：")
		sb.WriteString(req.QuestionContent)
		sb.WriteString("\n")
	}
	return sb.String()
}

func buildAddResp(r *AddResult, status string) dto.AddKnowledgePointItemResp {
	item := dto.AddKnowledgePointItemResp{
		UserKPID: r.UserKPID,
		Status:   status,
	}
	if r.KP != nil {
		kpResp := knowledgePointToResp(r.KP)
		item.Knowledge = &kpResp
	}
	if r.Existing != nil {
		item.Existing = &dto.ExistingKnowledgePoint{
			UserKPID:    r.Existing.ID,
			Name:        r.KP.Name,
			Description: r.KP.Description,
			IsImportant: r.Existing.IsImportant,
			IsMastered:  r.Existing.IsMastered,
			CreatedAt:   r.Existing.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return item
}

func knowledgePointToResp(kp *model.KnowledgePoint) dto.KnowledgePointResp {
	return dto.KnowledgePointResp{
		ID:          kp.ID,
		Name:        kp.Name,
		Description: kp.Description,
		SubjectID:   kp.SubjectID,
		Source:      kp.Source,
		CreatedAt:   kp.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   kp.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
