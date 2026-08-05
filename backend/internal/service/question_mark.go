package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

type QuestionMarkService struct {
	repo *repository.QuestionMarkRepo
}

func NewQuestionMarkService(repo *repository.QuestionMarkRepo) *QuestionMarkService {
	return &QuestionMarkService{repo: repo}
}

func (s *QuestionMarkService) Mark(userID, questionID uint) error {
	marked, err := s.repo.IsMarked(userID, questionID)
	if err != nil {
		return err
	}
	if marked {
		return nil
	}
	return s.repo.Add(&model.QuestionMark{
		UserID:     userID,
		QuestionID: questionID,
	})
}

func (s *QuestionMarkService) Unmark(userID, questionID uint) error {
	return s.repo.Remove(userID, questionID)
}

func (s *QuestionMarkService) IsMarked(userID, questionID uint) (bool, error) {
	return s.repo.IsMarked(userID, questionID)
}

func (s *QuestionMarkService) List(userID uint) ([]dto.QuestionMarkResp, error) {
	list, err := s.repo.FindByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.QuestionMarkResp, 0, len(list))
	for _, v := range list {
		resp = append(resp, dto.QuestionMarkResp{
			ID:         v.ID,
			QuestionID: v.QuestionID,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}
