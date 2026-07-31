package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

type SubSubjectService struct {
	repo *repository.SubSubjectRepo
}

func NewSubSubjectService(repo *repository.SubSubjectRepo) *SubSubjectService {
	return &SubSubjectService{repo: repo}
}

func (s *SubSubjectService) GetBySubjectID(subjectID uint) ([]dto.SubSubjectResp, error) {
	list, err := s.repo.FindBySubjectID(subjectID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.SubSubjectResp, len(list))
	for i, v := range list {
		resp[i] = dto.SubSubjectResp{
			ID:        v.ID,
			SubjectID: v.SubjectID,
			Name:      v.Name,
			SortOrder: v.SortOrder,
		}
	}
	return resp, nil
}
