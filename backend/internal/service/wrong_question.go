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
}

func NewWrongQuestionService(
	repo *repository.WrongQuestionRepo,
	questionRepo *repository.QuestionRepo,
	practiceRepo *repository.PracticeRecordRepo,
) *WrongQuestionService {
	return &WrongQuestionService{repo: repo, questionRepo: questionRepo, practiceRepo: practiceRepo}
}

func (s *WrongQuestionService) List(userID uint, subjectID *uint) ([]dto.WrongQuestionResp, error) {
	var list []model.WrongQuestion
	var err error
	if subjectID != nil {
		list, err = s.repo.FindByUserAndSubject(userID, *subjectID)
	} else {
		list, err = s.repo.FindByUser(userID)
	}
	if err != nil {
		return nil, err
	}

	resp := make([]dto.WrongQuestionResp, 0, len(list))
	for _, wq := range list {
		q, err := s.questionRepo.FindByID(wq.QuestionID)
		if err != nil {
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
		})
	}
	return resp, nil
}

func (s *WrongQuestionService) PracticeSubmit(userID uint, req dto.WrongPracticeSubmitReq) (*dto.PracticeRecordResp, error) {
	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, err
	}

	isCorrect := 0
	if question.Answer == req.Answer {
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

	if isCorrect == 1 {
		if err := s.repo.IncrementCorrect(userID, req.QuestionID); err != nil {
			return nil, err
		}

		if wq, err := s.repo.FindByUserAndQuestion(userID, req.QuestionID); err == nil && wq.CorrectCount >= 2 {
			s.repo.Delete(userID, req.QuestionID)
		}
	} else {
		// 答错时累加错误次数并更新最近错误时间
		if err := s.repo.Upsert(userID, req.QuestionID); err != nil {
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
