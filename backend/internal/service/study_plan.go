package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

// subjectLabelOf 错误信息里用的科目名：subjectID=0 走"全科目"通识文案
func subjectLabelOf(subjectID uint) string {
	if subjectID == 0 {
		return "全科目"
	}
	return fmt.Sprintf("科目#%d", subjectID)
}

// StudyPlanService 学习计划
type StudyPlanService struct {
	repo         *repository.StudyPlanRepo
	practiceRepo *repository.PracticeRecordRepo
}

func NewStudyPlanService(
	repo *repository.StudyPlanRepo,
	practiceRepo *repository.PracticeRecordRepo,
) *StudyPlanService {
	return &StudyPlanService{repo: repo, practiceRepo: practiceRepo}
}

// List 计划列表（含进度，自动完成已到期且达标的计划）
func (s *StudyPlanService) List(userID uint) ([]dto.StudyPlanResp, error) {
	plans, err := s.repo.FindByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.StudyPlanResp, 0, len(plans))
	for _, p := range plans {
		item := s.toResp(&p)
		if item.Status == 1 && time.Now().After(p.EndDate) && item.CompletionPct >= 100 {
			p.Status = 2
			if err := s.repo.Save(&p); err == nil {
				item.Status = 2
			}
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// Create 创建计划（同一用户同一科目仅允许 1 个进行中；不同科目 / 全科目互不影响）
// subjectID=0 表示"全科目"通识计划，独立维度去重
func (s *StudyPlanService) Create(userID uint, req dto.StudyPlanCreateReq) (*dto.StudyPlanResp, error) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("开始日期格式错误")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("结束日期格式错误")
	}
	if end.Before(start) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	// 同 subject 防冲突：允许"软设 + 高项 + 通识"多计划并行
	active, err := s.repo.FindActiveByUserAndSubject(userID, req.SubjectID)
	if err != nil {
		return nil, err
	}
	if len(active) > 0 {
		subjectLabel := subjectLabelOf(req.SubjectID)
		return nil, fmt.Errorf("「%s」已存在进行中的计划，请先完成或删除", subjectLabel)
	}

	plan := &model.StudyPlan{
		UserID:    userID,
		SubjectID: req.SubjectID,
		Title:     req.Title,
		DailyGoal: req.DailyGoal,
		StartDate: start,
		EndDate:   end,
		Status:    1,
	}
	if err := s.repo.Create(plan); err != nil {
		return nil, err
	}
	item := s.toResp(plan)
	return &item, nil
}

// Update 更新计划
func (s *StudyPlanService) Update(userID, planID uint, req dto.StudyPlanUpdateReq) (*dto.StudyPlanResp, error) {
	plan, err := s.repo.FindByID(planID)
	if err != nil {
		return nil, ErrNotFound
	}
	if plan.UserID != userID {
		return nil, ErrForbidden
	}
	if plan.Status != 1 {
		return nil, errors.New("计划已结束，无法修改")
	}

	if req.Title != "" {
		plan.Title = req.Title
	}
	// 修改 subject 时，校验新 subject 不存在其他进行中计划（排除自身）
	if req.SubjectID != nil && *req.SubjectID != plan.SubjectID {
		active, err := s.repo.FindActiveByUserAndSubject(userID, *req.SubjectID)
		if err != nil {
			return nil, err
		}
		for _, p := range active {
			if p.ID != plan.ID {
				subjectLabel := subjectLabelOf(*req.SubjectID)
				return nil, fmt.Errorf("「%s」已存在进行中的计划，请先完成或删除", subjectLabel)
			}
		}
		plan.SubjectID = *req.SubjectID
	}
	if req.DailyGoal != nil && *req.DailyGoal > 0 {
		plan.DailyGoal = *req.DailyGoal
	}
	if req.StartDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, errors.New("开始日期格式错误")
		}
		plan.StartDate = start
	}
	if req.EndDate != "" {
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, errors.New("结束日期格式错误")
		}
		plan.EndDate = end
	}
	if plan.EndDate.Before(plan.StartDate) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	if err := s.repo.Save(plan); err != nil {
		return nil, err
	}
	item := s.toResp(plan)
	return &item, nil
}

// Delete 删除计划
func (s *StudyPlanService) Delete(userID, planID uint) error {
	plan, err := s.repo.FindByID(planID)
	if err != nil {
		return ErrNotFound
	}
	if plan.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(plan)
}

// toResp 计算计划进度
func (s *StudyPlanService) toResp(p *model.StudyPlan) dto.StudyPlanResp {
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	todayDone, _ := s.practiceRepo.CountOnDate(p.UserID, todayStr, p.SubjectID)

	overallDone, _ := s.practiceRepo.CountInRange(
		p.UserID, p.StartDate.Format("2006-01-02"), p.EndDate.Format("2006-01-02"), p.SubjectID,
	)

	totalDays := int(p.EndDate.Sub(p.StartDate).Hours()/24) + 1
	overallGoal := p.DailyGoal * totalDays
	if overallGoal <= 0 {
		overallGoal = p.DailyGoal
	}

	completionPct := 0.0
	if overallGoal > 0 {
		completionPct = float64(overallDone) / float64(overallGoal) * 100
	}

	remaining := int(p.EndDate.Sub(today).Hours()/24) + 1
	remainingDays := 0
	if remaining > 0 {
		remainingDays = remaining
	}

	return dto.StudyPlanResp{
		ID:            p.ID,
		SubjectID:     p.SubjectID,
		Title:         p.Title,
		DailyGoal:     p.DailyGoal,
		StartDate:     p.StartDate.Format("2006-01-02"),
		EndDate:       p.EndDate.Format("2006-01-02"),
		Status:        p.Status,
		TodayDone:     todayDone,
		TodayGoal:     p.DailyGoal,
		OverallDone:   overallDone,
		OverallGoal:   overallGoal,
		CompletionPct: completionPct,
		RemainingDays: remainingDays,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04"),
	}
}
