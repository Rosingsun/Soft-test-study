package service

import (
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

type QuestionService struct {
	repo *repository.QuestionRepo
}

func NewQuestionService(repo *repository.QuestionRepo) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) GetByChapterID(chapterID uint, difficulty string) ([]dto.QuestionResp, error) {
	questions, err := s.repo.FindByChapterIDFiltered(chapterID, difficulty)
	if err != nil {
		return nil, err
	}
	return toQuestionRespList(questions), nil
}

func (s *QuestionService) GetByID(id uint) (*dto.QuestionResp, error) {
	q, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := toQuestionResp(*q)
	return &resp, nil
}

// GetRandomQuestions 随机抽题，subjectID 必填
func (s *QuestionService) GetRandomQuestions(subjectID uint, count int, difficulty string) ([]dto.QuestionResp, error) {
	if count <= 0 {
		count = 10
	}
	if count > 50 {
		count = 50
	}
	questions, err := s.repo.FindRandom(subjectID, difficulty, count)
	if err != nil {
		return nil, err
	}
	return toQuestionRespList(questions), nil
}

// GetSpecialQuestions 专项练习：按题型/难度抽题
func (s *QuestionService) GetSpecialQuestions(subjectID uint, qtype, difficulty string, count int) ([]dto.QuestionResp, error) {
	if count <= 0 {
		count = 10
	}
	if count > 50 {
		count = 50
	}
	questions, err := s.repo.FindSpecial(subjectID, qtype, difficulty, count)
	if err != nil {
		return nil, err
	}
	return toQuestionRespList(questions), nil
}

func toQuestionRespList(questions []model.Question) []dto.QuestionResp {
	resp := make([]dto.QuestionResp, len(questions))
	for i, q := range questions {
		resp[i] = toQuestionResp(q)
	}
	return resp
}

func toQuestionResp(q model.Question) dto.QuestionResp {
	return dto.QuestionResp{
		ID:           q.ID,
		SubjectID:    q.SubjectID,
		SubSubjectID: q.SubSubjectID,
		ChapterID:    q.ChapterID,
		Type:         q.Type,
		Difficulty:   q.Difficulty,
		Content:      q.Content,
		Options:      q.Options,
		Answer:       q.Answer,
		Analysis:     q.Analysis,
		Year:         q.Year,
	}
}
