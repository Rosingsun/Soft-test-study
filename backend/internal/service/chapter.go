package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

type ChapterService struct {
	repo         *repository.ChapterRepo
	questionRepo *repository.QuestionRepo
}

func NewChapterService(repo *repository.ChapterRepo, questionRepo *repository.QuestionRepo) *ChapterService {
	return &ChapterService{repo: repo, questionRepo: questionRepo}
}

func (s *ChapterService) GetBySubSubjectID(subSubjectID uint) ([]dto.ChapterResp, error) {
	chapters, err := s.repo.FindBySubSubjectID(subSubjectID)
	if err != nil {
		return nil, err
	}
	counts, err := s.questionRepo.CountBySubSubjectGroupChapter(subSubjectID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ChapterResp, len(chapters))
	for i, v := range chapters {
		resp[i] = dto.ChapterResp{
			ID:            v.ID,
			SubjectID:     v.SubjectID,
			SubSubjectID:  v.SubSubjectID,
			ParentID:      v.ParentID,
			Name:          v.Name,
			SortOrder:     v.SortOrder,
			MaterialID:    v.MaterialID,
			QuestionCount: counts[v.ID],
		}
	}
	return resp, nil
}

func (s *ChapterService) GetBySubjectID(subjectID uint) ([]dto.ChapterResp, error) {
	chapters, err := s.repo.FindBySubjectID(subjectID)
	if err != nil {
		return nil, err
	}
	counts, err := s.questionRepo.CountBySubjectGroupChapter(subjectID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ChapterResp, len(chapters))
	for i, v := range chapters {
		resp[i] = dto.ChapterResp{
			ID:            v.ID,
			SubjectID:     v.SubjectID,
			SubSubjectID:  v.SubSubjectID,
			ParentID:      v.ParentID,
			Name:          v.Name,
			SortOrder:     v.SortOrder,
			MaterialID:    v.MaterialID,
			QuestionCount: counts[v.ID],
		}
	}
	return resp, nil
}
