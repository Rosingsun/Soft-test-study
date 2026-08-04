package service

import (
	"errors"
	"math"
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

const examPaperSize = 75

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
		questionCount := examPaperSize
		if t.QuestionType == model.TypeEssay {
			questionCount = 1
		} else if t.QuestionType != "" {
			if n, err := s.questionRepo.CountBySubjectAndType(t.SubjectID, t.QuestionType); err == nil {
				questionCount = int(n)
			}
		}
		resp[i] = dto.ExamTemplateResp{
			ID:            t.ID,
			SubjectID:     t.SubjectID,
			Name:          t.Name,
			Duration:      t.Duration,
			TotalScore:    t.TotalScore,
			Year:          t.Year,
			QuestionType:  t.QuestionType,
			QuestionCount: questionCount,
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
		if int(time.Since(pending.StartedAt).Seconds()) >= template.Duration*60 {
			if _, err := s.finishExam(userID, pending.ID); err != nil {
				return nil, err
			}
		}
		return s.loadExam(pending.ID, userID, template)
	}

	now := time.Now()
	record := &model.ExamRecord{
		UserID:     userID,
		TemplateID: templateID,
		TotalScore: examPaperSize,
		Status:     "pending",
		StartedAt:  now,
	}
	if err := s.recordRepo.Create(record); err != nil {
		return nil, err
	}

	// 论文卷每次随机抽 1 题，其余试卷随机抽 examPaperSize 道
	limit := examPaperSize
	if template.QuestionType == model.TypeEssay {
		limit = 1
	}
	questions, err := s.questionRepo.FindRandomFiltered(template.SubjectID, "", template.QuestionType, limit)
	if err != nil {
		return nil, err
	}
	if len(questions) > 0 {
		record.TotalScore = len(questions)
	}
	if err := s.recordRepo.Save(record); err != nil {
		return nil, err
	}

	snapshot := make([]model.ExamRecordAnswer, 0, len(questions))
	for _, q := range questions {
		snapshot = append(snapshot, model.ExamRecordAnswer{
			RecordID:   record.ID,
			QuestionID: q.ID,
			Answer:     "",
			IsCorrect:  0,
			Score:      0,
		})
	}
	if err := s.recordRepo.CreateBatchAnswers(snapshot); err != nil {
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

	questionList := make([]dto.ExamQuesResp, 0, len(answers))
	for i, a := range answers {
		q, ok := qMap[a.QuestionID]
		if !ok {
			continue
		}
		questionList = append(questionList, dto.ExamQuesResp{
			ID:         q.ID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    q.Options,
			Difficulty: q.Difficulty,
			Score:      1,
			SortOrder:  i + 1,
		})
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
		TotalScore:   record.TotalScore,
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

	snapshot, err := s.recordRepo.FindAnswersByRecord(recordID)
	if err != nil {
		return err
	}
	inPaper := false
	for _, a := range snapshot {
		if a.QuestionID == req.QuestionID {
			inPaper = true
			break
		}
	}
	if !inPaper {
		return errors.New("题目不属于本次考试")
	}

	isCorrect := 0
	score := 0
	if question.Type != model.TypeEssay {
		if question.Answer == req.Answer {
			isCorrect = 1
		}
		if isCorrect == 1 {
			score = 1
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
	return s.finishExam(userID, recordID)
}

func (s *ExamService) finishExam(userID, recordID uint) (*dto.ExamResultResp, error) {
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
	totalQuestions := 0
	for _, a := range answers {
		totalScore += a.Score
		if a.IsCorrect == 1 {
			correctCount++
		}
		if a.Score > 0 {
			totalQuestions++
		}
	}

	accuracy := float64(0)
	if totalQuestions > 0 {
		accuracy = math.Round(float64(correctCount)/float64(totalQuestions)*10000) / 100
	}

	record.Score = totalScore
	record.CorrectCount = correctCount
	record.Accuracy = accuracy
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

	answers, err := s.recordRepo.FindAnswersByRecord(recordID)
	if err != nil {
		return nil, err
	}
	totalQuestions := 0
	for _, a := range answers {
		if a.Score > 0 {
			totalQuestions++
		}
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

	// 构建模板相关字段：支持 AI 考试 (TemplateID=0) 和普通模板考试
	templateName := "AI 智能组卷"
	templateSubjectID := uint(0)
	templateTotalScore := record.TotalScore

	if record.TemplateID != 0 {
		if template, err := s.templateRepo.FindByID(record.TemplateID); err == nil {
			templateName = template.Name
			templateSubjectID = template.SubjectID
			templateTotalScore = template.TotalScore
		}
	} else if len(questions) > 0 {
		templateSubjectID = questions[0].SubjectID
	}

	correctCount := 0
	sectionMap := make(map[string]*dto.SectionAccuracyResp)
	details := make([]dto.AnswerDetailResp, 0, len(answers))
	for _, a := range answers {
		if a.Score > 0 {
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
		}

		q := qMap[a.QuestionID]
		details = append(details, dto.AnswerDetailResp{
			QuestionID:   a.QuestionID,
			ExamAnswerID: a.ID,
			Type:         q.Type,
			Content:      q.Content,
			Options:      q.Options,
			YourAnswer:   a.Answer,
			CorrectAns:   q.Answer,
			IsCorrect:    a.IsCorrect,
			Score:        a.Score,
			Analysis:     q.Analysis,
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

	accuracy := float64(0)
	if totalQuestions > 0 {
		accuracy = math.Round(float64(correctCount)/float64(totalQuestions)*10000) / 100
	}

	return &dto.ExamResultResp{
		RecordID:        recordID,
		TemplateName:    templateName,
		SubjectID:       templateSubjectID,
		Score:           record.Score,
		TotalScore:      templateTotalScore,
		Duration:        record.Duration,
		CorrectCount:    correctCount,
		TotalCount:      totalQuestions,
		Accuracy:        accuracy,
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
		if r.TemplateID == 0 {
			// AI 考试：从答卷关联的题目中获取科目
			if sub, err := s.getAISubjectID(r); err == nil {
				subjectID = sub
			}
			templateName = "AI 智能组卷"
		} else if t, err := s.templateRepo.FindByID(r.TemplateID); err == nil {
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
			CorrectCount: r.CorrectCount,
			Accuracy:     r.Accuracy,
			Status:       r.Status,
			StartedAt:    formatTime(r.StartedAt),
			FinishedAt:   formatTime(r.FinishedAt),
			CreatedAt:    formatTime(r.CreatedAt),
		})
	}
	return resp, nil
}

// getAISubjectID 通过 AI 考试成绩记录的题目，反查科目 ID
func (s *ExamService) getAISubjectID(record model.ExamRecord) (uint, error) {
	answers, err := s.recordRepo.FindAnswersByRecord(record.ID)
	if err != nil || len(answers) == 0 {
		return 0, ErrNotFound
	}
	questionIDs := make([]uint, len(answers))
	for i, a := range answers {
		questionIDs[i] = a.QuestionID
	}
	questions, err := s.questionRepo.FindByIDs(questionIDs)
	if err != nil || len(questions) == 0 {
		return 0, ErrNotFound
	}
	return questions[0].SubjectID, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
