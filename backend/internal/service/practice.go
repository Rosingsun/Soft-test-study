package service

import (
	"log"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

type PracticeService struct {
	recordRepo   *repository.PracticeRecordRepo
	questionRepo *repository.QuestionRepo
	reviewSvc    *ReviewService
}

func NewPracticeService(
	recordRepo *repository.PracticeRecordRepo,
	questionRepo *repository.QuestionRepo,
	reviewSvc *ReviewService,
) *PracticeService {
	return &PracticeService{recordRepo: recordRepo, questionRepo: questionRepo, reviewSvc: reviewSvc}
}

func (s *PracticeService) Submit(userID uint, req dto.PracticeSubmitReq) (*dto.PracticeRecordResp, error) {
	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, err
	}

	isCorrect := 0
	if question.Type != model.TypeEssay && IsAnswerCorrect(question.Type, question.Answer, req.Answer) {
		isCorrect = 1
	}

	record := &model.PracticeRecord{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Mode:       req.Mode,
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Duration:   req.Duration,
	}
	if err := s.recordRepo.Create(record); err != nil {
		return nil, err
	}

	// 答错自动收录错题本并进入遗忘曲线复习（论文题不判分，不收录）
	if isCorrect == 0 && question.Type != model.TypeEssay {
		if err := s.reviewSvc.OnWrong(userID, req.QuestionID); err != nil {
			log.Printf("错题本/复习卡片收录失败 user_id=%d question_id=%d: %v", userID, req.QuestionID, err)
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
