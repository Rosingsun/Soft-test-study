package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

type SubjectService struct {
	repo *repository.SubjectRepo
}

func NewSubjectService(repo *repository.SubjectRepo) *SubjectService {
	return &SubjectService{repo: repo}
}

func (s *SubjectService) GetByLevelID(levelID uint) ([]dto.SubjectResp, error) {
	subjects, err := s.repo.FindByLevelID(levelID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.SubjectResp, len(subjects))
	for i, v := range subjects {
		resp[i] = dto.SubjectResp{
			ID:          v.ID,
			LevelID:     v.LevelID,
			Name:        v.Name,
			ShortName:   v.ShortName,
			Description: v.Description,
			Icon:        v.Icon,
			SortOrder:   v.SortOrder,
		}
	}
	return resp, nil
}

func (s *SubjectService) GetAll() ([]dto.SubjectResp, error) {
	subjects, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	resp := make([]dto.SubjectResp, len(subjects))
	for i, v := range subjects {
		resp[i] = dto.SubjectResp{
			ID:          v.ID,
			LevelID:     v.LevelID,
			Name:        v.Name,
			ShortName:   v.ShortName,
			Description: v.Description,
			Icon:        v.Icon,
			SortOrder:   v.SortOrder,
		}
	}
	return resp, nil
}
