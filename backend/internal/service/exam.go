package service

import (
	"errors"
	"sort"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"gorm.io/gorm"
)

type ExamService struct {
	templateRepo *repository.ExamTemplateRepo
	recordRepo   *repository.ExamRecordRepo
	questionRepo *repository.QuestionRepo
}

func NewExamService(
	templateRepo *repository.ExamTemplateRepo,
	recordRepo *repository.ExamRecordRepo,
	questionRepo *repository.QuestionRepo,
) *ExamService {
	return &ExamService{templateRepo: templateRepo, recordRepo: recordRepo, questionRepo: questionRepo}
}

func (s *ExamService) ListTemplates() ([]dto.ExamTemplateResp, error) {
	templates, err := s.templateRepo.FindPublic()
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ExamTemplateResp, len(templates))
	for i, t := range templates {
		count, err := s.templateRepo.CountQuestions(t.ID)
		if err != nil {
			return nil, err
		}
		resp[i] = dto.ExamTemplateResp{
			ID:            t.ID,
			SubjectID:     t.SubjectID,
			Name:          t.Name,
			Duration:      t.Duration,
			TotalScore:    t.TotalScore,
			Year:          t.Year,
			QuestionCount: int(count),
		}
	}
	return resp, nil
}

func (s *ExamService) StartExam(userID, templateID uint) (*dto.StartExamResp, error) {
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, ErrNotFound
	}

	pending, err := s.recordRepo.FindPendingByUserAndTemplate(userID, templateID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if pending != nil && pending.ID > 0 {
		return s.loadExam(pending.ID, userID, template)
	}

	now := time.Now()
	record := &model.ExamRecord{
		UserID:     userID,
		TemplateID: templateID,
		TotalScore: template.TotalScore,
		Status:     "pending",
		StartedAt:  now,
	}
	if err := s.recordRepo.Create(record); err != nil {
		return nil, err
	}

	return s.loadExam(record.ID, userID, template)
}

func (s *ExamService) loadExam(recordID, userID uint, template *model.ExamTemplate) (*dto.StartExamResp, error) {
	record, err := s.recordRepo.FindByID(recordID)
	if err != nil {
		return nil, ErrNotFound
	}
	if record.UserID != userID {
		return nil, ErrForbidden
	}

	tplQuestions, err := s.templateRepo.FindQuestionsByTemplate(template.ID)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]uint, len(tplQuestions))
	scoreMap := make(map[uint]int)
	for i, tq := range tplQuestions {
		questionIDs[i] = tq.QuestionID
		scoreMap[tq.QuestionID] = tq.Score
	}

	questions, err := s.questionRepo.FindByIDs(questionIDs)
	if err != nil {
		return nil, err
	}

	qMap := make(map[uint]model.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	questionList := make([]dto.ExamQuesResp, 0, len(tplQuestions))
	for _, tq := range tplQuestions {
		q, ok := qMap[tq.QuestionID]
		if !ok {
			continue
		}
		questionList = append(questionList, dto.ExamQuesResp{
			ID:         q.ID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    q.Options,
			Difficulty: q.Difficulty,
			Score:      scoreMap[q.ID],
			SortOrder:  tq.SortOrder,
		})
	}

	answers, err := s.recordRepo.FindAnswersByRecord(recordID)
	if err != nil {
		return nil, err
	}
	answerMap := make(map[uint]string)
	for _, a := range answers {
		answerMap[a.QuestionID] = a.Answer
	}

	elapsed := int(time.Since(record.StartedAt).Seconds())

	return &dto.StartExamResp{
		RecordID:     recordID,
		TemplateID:   template.ID,
		TemplateName: template.Name,
		Duration:     template.Duration,
		TotalScore:   template.TotalScore,
		Questions:    questionList,
		Answers:      answerMap,
		Elapsed:      elapsed,
		Status:       record.Status,
	}, nil
}

func (s *ExamService) SubmitAnswer(userID, recordID uint, req dto.SubmitAnswerReq) error {
	record, err := s.recordRepo.FindByID(recordID)
	if err != nil {
		return ErrNotFound
	}
	if record.UserID != userID {
		return ErrForbidden
	}
	if record.Status != "pending" {
		return errors.New("考试已交卷，无法修改答案")
	}

	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return err
	}

	isCorrect := 0
	if question.Answer == req.Answer {
		isCorrect = 1
	}

	tplQuestions, err := s.templateRepo.FindQuestionsByTemplate(record.TemplateID)
	if err != nil {
		return err
	}
	score := 0
	for _, tq := range tplQuestions {
		if tq.QuestionID == req.QuestionID {
			if isCorrect == 1 {
				score = tq.Score
			}
			break
		}
	}

	answer := &model.ExamRecordAnswer{
		RecordID:   recordID,
		QuestionID: req.QuestionID,
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Score:      score,
	}
	return s.recordRepo.SaveAnswer(answer)
}

func (s *ExamService) SubmitExam(userID, recordID uint) (*dto.ExamResultResp, error) {
	record, err := s.recordRepo.FindByID(recordID)
	if err != nil {
		return nil, ErrNotFound
	}
	if record.UserID != userID {
		return nil, ErrForbidden
	}
	if record.Status == "finished" {
		return s.GetResult(userID, recordID)
	}

	answers, err := s.recordRepo.FindAnswersByRecord(recordID)
	if err != nil {
		return nil, err
	}
	totalScore := 0
	correctCount := 0
	for _, a := range answers {
		totalScore += a.Score
		if a.IsCorrect == 1 {
			correctCount++
		}
	}

	record.Score = totalScore
	record.Status = "finished"
	record.FinishedAt = time.Now()
	record.Duration = int(time.Since(record.StartedAt).Seconds())
	if err := s.recordRepo.Save(record); err != nil {
		return nil, err
	}

	return s.GetResult(userID, recordID)
}

func (s *ExamService) GetResult(userID, recordID uint) (*dto.ExamResultResp, error) {
	record, err := s.recordRepo.FindByID(recordID)
	if err != nil {
		return nil, ErrNotFound
	}
	if record.UserID != userID {
		return nil, ErrForbidden
	}

	template, err := s.templateRepo.FindByID(record.TemplateID)
	if err != nil {
		return nil, err
	}

	answers, err := s.recordRepo.FindAnswersByRecord(recordID)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]uint, len(answers))
	for i, a := range answers {
		questionIDs[i] = a.QuestionID
	}

	questions, err := s.questionRepo.FindByIDs(questionIDs)
	if err != nil {
		return nil, err
	}
	qMap := make(map[uint]model.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	correctCount := 0
	sectionMap := make(map[string]*dto.SectionAccuracyResp)
	details := make([]dto.AnswerDetailResp, 0, len(answers))
	for _, a := range answers {
		if a.IsCorrect == 1 {
			correctCount++
		}
		q := qMap[a.QuestionID]

		sec := sectionMap[q.Type]
		if sec == nil {
			sec = &dto.SectionAccuracyResp{Type: q.Type}
			sectionMap[q.Type] = sec
		}
		sec.TotalCount++
		if a.IsCorrect == 1 {
			sec.CorrectCount++
		}

		details = append(details, dto.AnswerDetailResp{
			QuestionID: a.QuestionID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    q.Options,
			YourAnswer: a.Answer,
			CorrectAns: q.Answer,
			IsCorrect:  a.IsCorrect,
			Score:      a.Score,
			Analysis:   q.Analysis,
		})
	}

	sectionAcc := make([]dto.SectionAccuracyResp, 0, len(sectionMap))
	for _, sec := range sectionMap {
		if sec.TotalCount > 0 {
			sec.Accuracy = float64(sec.CorrectCount) / float64(sec.TotalCount) * 100
		}
		sectionAcc = append(sectionAcc, *sec)
	}
	sort.Slice(sectionAcc, func(i, j int) bool { return sectionAcc[i].Type < sectionAcc[j].Type })

	return &dto.ExamResultResp{
		RecordID:        recordID,
		TemplateName:    template.Name,
		SubjectID:       template.SubjectID,
		Score:           record.Score,
		TotalScore:      template.TotalScore,
		Duration:        record.Duration,
		CorrectCount:    correctCount,
		TotalCount:      len(answers),
		Status:          record.Status,
		StartedAt:       formatTime(record.StartedAt),
		FinishedAt:      formatTime(record.FinishedAt),
		SectionAccuracy: sectionAcc,
		AnswerDetails:   details,
	}, nil
}

func (s *ExamService) ListRecords(userID uint) ([]dto.ExamRecordResp, error) {
	records, err := s.recordRepo.FindByUser(userID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ExamRecordResp, 0, len(records))
	for _, r := range records {
		templateName := ""
		subjectID := r.TemplateID
		if t, err := s.templateRepo.FindByID(r.TemplateID); err == nil {
			templateName = t.Name
			subjectID = t.SubjectID
		}
		resp = append(resp, dto.ExamRecordResp{
			ID:           r.ID,
			TemplateID:   r.TemplateID,
			TemplateName: templateName,
			SubjectID:    subjectID,
			Score:        r.Score,
			TotalScore:   r.TotalScore,
			Duration:     r.Duration,
			Status:       r.Status,
			StartedAt:    formatTime(r.StartedAt),
			FinishedAt:   formatTime(r.FinishedAt),
			CreatedAt:    formatTime(r.CreatedAt),
		})
	}
	return resp, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
