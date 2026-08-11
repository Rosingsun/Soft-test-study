package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

type WrongQuestionService struct {
	repo         *repository.WrongQuestionRepo
	questionRepo *repository.QuestionRepo
	practiceRepo *repository.PracticeRecordRepo
	reviewSvc    *ReviewService
}

func NewWrongQuestionService(
	repo *repository.WrongQuestionRepo,
	questionRepo *repository.QuestionRepo,
	practiceRepo *repository.PracticeRecordRepo,
	reviewSvc *ReviewService,
) *WrongQuestionService {
	return &WrongQuestionService{repo: repo, questionRepo: questionRepo, practiceRepo: practiceRepo, reviewSvc: reviewSvc}
}

// List 错题本列表。
//
// 原实现：N 条错题 → 循环 N 次 questionRepo.FindByID（典型 N+1，远程 DB 下尤其致命）
// 改造：先一次 FindByIDs 批量拿全部题目，再 map 拼装，固定 2 次 DB 往返
func (s *WrongQuestionService) List(userID uint, subjectID *uint, source, sort string) ([]dto.WrongQuestionResp, error) {
	list, err := s.repo.Query(userID, subjectID, source, sort)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return []dto.WrongQuestionResp{}, nil
	}

	// 收集去重的题目 ID
	idSet := make(map[uint]struct{}, len(list))
	ids := make([]uint, 0, len(list))
	for _, wq := range list {
		if _, ok := idSet[wq.QuestionID]; ok {
			continue
		}
		idSet[wq.QuestionID] = struct{}{}
		ids = append(ids, wq.QuestionID)
	}
	questions, err := s.questionRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	qMap := make(map[uint]model.Question, len(questions))
	for _, q := range questions {
		qMap[q.ID] = q
	}

	resp := make([]dto.WrongQuestionResp, 0, len(list))
	for _, wq := range list {
		q, ok := qMap[wq.QuestionID]
		if !ok {
			continue
		}
		// 防御性过滤：主观题不应出现在错题本里。历史脏数据 / 直接 SQL 写入等情况
		// 可能混入主观题（essay / case_study），前端展示"永远答错"会破坏体验，故静默隐藏。
		if IsSubjectiveType(q.Type) {
			continue
		}
		resp = append(resp, dto.WrongQuestionResp{
			ID:           wq.ID,
			QuestionID:   wq.QuestionID,
			WrongCount:   wq.WrongCount,
			CorrectCount: wq.CorrectCount,
			LastWrongAt:  wq.LastWrongAt.Format("2006-01-02 15:04"),
			CreatedAt:    wq.CreatedAt.Format("2006-01-02 15:04"),
			Content:      q.Content,
			Type:         q.Type,
			SubjectID:    q.SubjectID,
			Source:       q.Source,
		})
	}
	return resp, nil
}

func (s *WrongQuestionService) PracticeSubmit(userID uint, req dto.WrongPracticeSubmitReq) (*dto.PracticeRecordResp, error) {
	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, err
	}

	// 主观题（论文 / 案例分析）不应进入错题本与复习卡流程：
	//   - 不做客观题判分（占位答案 vs 用户答案永远不等），否则会强制 0 分
	//   - 不触发 OnWrong()，否则每次练习都会把复习卡重置到 DueDate=today，
	//     形成"主观题永远待复习"的死循环
	//   - 仍记录 practice_records（保留作答痕迹），但 is_correct 置 0 表示"未判分"
	isCorrect := 0
	if !IsSubjectiveType(question.Type) && IsAnswerCorrect(question.Type, question.Answer, req.Answer) {
		isCorrect = 1
	}

	record := &model.PracticeRecord{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Mode:       "wrong",
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Duration:   req.Duration,
	}
	if err := s.practiceRepo.Create(record); err != nil {
		return nil, err
	}

	// 主观题：跳过所有错题本/复习卡联动，仅返回作答记录
	if IsSubjectiveType(question.Type) {
		return &dto.PracticeRecordResp{
			ID:         record.ID,
			QuestionID: record.QuestionID,
			Mode:       record.Mode,
			Answer:     record.Answer,
			IsCorrect:  record.IsCorrect,
			Duration:   record.Duration,
		}, nil
	}

	if isCorrect == 1 {
		if err := s.repo.IncrementCorrect(userID, req.QuestionID); err != nil {
			return nil, err
		}

		if wq, err := s.repo.FindByUserAndQuestion(userID, req.QuestionID); err == nil && wq.CorrectCount >= 2 {
			s.repo.Delete(userID, req.QuestionID)
		}
	} else {
		// 答错时累加错误次数、更新最近错误时间并刷新复习卡片
		if err := s.reviewSvc.OnWrong(userID, req.QuestionID); err != nil {
			return nil, err
		}
	}

	return &dto.PracticeRecordResp{
		ID:         record.ID,
		QuestionID: record.QuestionID,
		Mode:       record.Mode,
		Answer:     record.Answer,
		IsCorrect:  record.IsCorrect,
		Duration:   record.Duration,
	}, nil
}

func (s *WrongQuestionService) Count(userID uint) (int64, error) {
	return s.repo.CountByUser(userID)
}

func (s *WrongQuestionService) Remove(userID, questionID uint) error {
	return s.repo.Delete(userID, questionID)
}
