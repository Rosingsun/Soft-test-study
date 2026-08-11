package service

import (
	"encoding/json"

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

func (s *QuestionService) GetEssayQuestions(subjectID uint, years int) ([]dto.QuestionResp, error) {
	questions, err := s.repo.FindEssayQuestions(subjectID, years)
	if err != nil {
		return nil, err
	}
	return toQuestionRespList(questions), nil
}

// GetCaseStudies 按子科目查询案例分析题
func (s *QuestionService) GetCaseStudies(subSubjectID uint, year int) ([]dto.QuestionResp, error) {
	questions, err := s.repo.FindCaseStudies(subSubjectID, year)
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
		CaseMaterial: q.CaseMaterial,
		Options:      q.Options,
		BlankOptions: q.BlankOptions,
		Answer:       q.Answer,
		Analysis:     q.Analysis,
		Year:         q.Year,
		Source:       q.Source,
	}
}

// IsSubjectiveType 是否为主观题（论文 / 案例分析）：不自动判分，不收录错题本
func IsSubjectiveType(questionType string) bool {
	return questionType == model.TypeEssay || questionType == model.TypeCaseStudy
}

// IsAnswerCorrect 客观题通用判分：多空题按 JSON 数组逐空比对，全部正确才算整题正确；其他题型字符串相等
// 供 service 包内所有判分场景复用（练习 / 考试 / 打卡 / 错题重做 / 复习）
func IsAnswerCorrect(questionType, correctAnswer, userAnswer string) bool {
	if questionType == model.TypeMultiBlank {
		var correctArr, userArr []string
		if err := json.Unmarshal([]byte(correctAnswer), &correctArr); err != nil {
			return false
		}
		if err := json.Unmarshal([]byte(userAnswer), &userArr); err != nil {
			return false
		}
		if len(correctArr) != len(userArr) {
			return false
		}
		for i := range correctArr {
			if correctArr[i] != userArr[i] {
				return false
			}
		}
		return true
	}
	return correctAnswer == userAnswer
}
