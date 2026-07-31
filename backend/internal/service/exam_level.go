package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

type ExamLevelService struct {
	repo *repository.ExamLevelRepo
}

func NewExamLevelService(repo *repository.ExamLevelRepo) *ExamLevelService {
	return &ExamLevelService{repo: repo}
}

func (s *ExamLevelService) GetAll() ([]dto.ExamLevelResp, error) {
	levels, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ExamLevelResp, len(levels))
	for i, v := range levels {
		resp[i] = dto.ExamLevelResp{ID: v.ID, Name: v.Name}
	}
	return resp, nil
}
