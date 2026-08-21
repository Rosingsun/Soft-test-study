package service

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"gorm.io/gorm"
)

// ErrExamInProgress 已有进行中的考试（OPT-07）
var ErrExamInProgress = errors.New("已有进行中的考试")

type ExamService struct {
	db           *gorm.DB
	templateRepo *repository.ExamTemplateRepo
	recordRepo   *repository.ExamRecordRepo
	questionRepo *repository.QuestionRepo
	reviewSvc    *ReviewService
}

const examPaperSize = 75

func NewExamService(
	db *gorm.DB,
	templateRepo *repository.ExamTemplateRepo,
	recordRepo *repository.ExamRecordRepo,
	questionRepo *repository.QuestionRepo,
	reviewSvc *ReviewService,
) *ExamService {
	return &ExamService{db: db, templateRepo: templateRepo, recordRepo: recordRepo, questionRepo: questionRepo, reviewSvc: reviewSvc}
}

// ListTemplates 公开考试模板列表。
//
// 原实现：N 个模板 → 循环 N 次 questionRepo.CountBySubjectAndType（N+1）
// 改造：先用 templateRepo.FindPublic 一次拉全；再调用 questionRepo.CountBySubjectTypes
//       按 (subject_id, type) 一次性 GROUP BY 聚合
func (s *ExamService) ListTemplates() ([]dto.ExamTemplateResp, error) {
	templates, err := s.templateRepo.FindPublic()
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return []dto.ExamTemplateResp{}, nil
	}

	// 收集需要按 type 统计题量的模板
	pairs := make([]repository.SubjectTypePair, 0, len(templates))
	for _, t := range templates {
		if t.QuestionType != "" && t.QuestionType != model.TypeEssay {
			pairs = append(pairs, repository.SubjectTypePair{
				SubjectID: t.SubjectID,
				Type:      t.QuestionType,
			})
		}
	}
	countMap, err := s.questionRepo.CountBySubjectTypes(pairs, "ai")
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ExamTemplateResp, len(templates))
	for i, t := range templates {
		questionCount := examPaperSize
		if t.QuestionType == model.TypeEssay {
			questionCount = 1
		} else if t.QuestionType != "" {
			questionCount = int(countMap[repository.SubjectTypePair{
				SubjectID: t.SubjectID,
				Type:      t.QuestionType,
			}])
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
		// pending 已超时：自动交卷（与 janitor 同路径，结果一致）
		if _, err := s.settleIfExpired(pending, template, userID); err != nil {
			return nil, err
		}
		return s.loadExam(pending.ID, userID, template)
	}

	// OPT-07: 用事务包裹「查询 pending → 创建 record」，
	// 配合 uk_exam_pending 唯一索引，并发 StartExam 仅一条成功。
	now := time.Now()
	var recordID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		record := &model.ExamRecord{
			UserID:     userID,
			TemplateID: templateID,
			TotalScore: examPaperSize,
			Status:     "pending",
			StartedAt:  now,
		}
		if err := tx.Create(record).Error; err != nil {
			// MySQL 1062 唯一约束冲突 = 已有 pending 记录
			var me *mysqldriver.MySQLError
			if errors.As(err, &me) && me.Number == 1062 {
				return ErrExamInProgress
			}
			return err
		}
		recordID = record.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrExamInProgress) {
			return nil, err
		}
		return nil, err
	}

	// 论文卷每次随机抽 1 题，其余试卷随机抽 examPaperSize 道
	limit := examPaperSize
	if template.QuestionType == model.TypeEssay {
		limit = 1
	}
	// 普通模拟考试：排除 AI 生成的题目（AI 智能组卷走 StartAiExam 单独通道）
	questions, err := s.questionRepo.FindRandomFiltered(template.SubjectID, "", template.QuestionType, "ai", limit)
	if err != nil {
		return nil, err
	}
	record := &model.ExamRecord{ID: recordID}
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

// settleIfExpired 若 record 仍 pending 且已超过 template 时限，则自动 finishExam。
// 返回 settled=true 表示本次调用完成了一次超时结算。
// 行为复用：StartExam 入口防御 + 后台 janitor 周期扫描。
func (s *ExamService) settleIfExpired(record *model.ExamRecord, template *model.ExamTemplate, userID uint) (settled bool, err error) {
	if record.Status != "pending" {
		return record.Status == "finished", nil
	}
	if int(time.Since(record.StartedAt).Seconds()) < template.Duration*60 {
		return false, nil
	}
	if _, err := s.finishExam(userID, record.ID); err != nil {
		return false, err
	}
	return true, nil
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
			ID:           q.ID,
			Type:         q.Type,
			Content:      q.Content,
			CaseMaterial: q.CaseMaterial,
			Options:      q.Options,
			BlankOptions: q.BlankOptions,
			Difficulty:   q.Difficulty,
			Score:        1,
			SortOrder:    i + 1,
			Source:       q.Source,
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

	// 超时防御：即使 janitor 尚未跑、客户端定时器还显示正数，服务器侧也拒绝继续保存答案。
	// 引导前端"考试已超时"提示并刷新查看结果页。
	// AI 考试（template_id=0）依赖客户端 / StartExam 入口防御，这里跳过。
	if record.TemplateID > 0 {
		template, terr := s.templateRepo.FindByID(record.TemplateID)
		if terr == nil && int(time.Since(record.StartedAt).Seconds()) >= template.Duration*60 {
			// 顺手触发结算，让下次刷新直接看到结果
			_, _ = s.finishExam(userID, record.ID)
			return errors.New("考试已超时，请刷新页面查看结果")
		}
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
		if IsAnswerCorrect(question.Type, question.Answer, req.Answer) {
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

	// 模考答错的客观题自动收录错题本 + 遗忘曲线复习卡片
	s.recordWrongAnswers(userID, answers)

	return s.GetResult(userID, recordID)
}

// recordWrongAnswers 模考中答错的客观题写入错题本与复习卡片。
//
// 原实现：N 道错题 → 循环 N 次 reviewSvc.OnWrong，每次内部 3+ 次 DB 往返（75 题考试 = 300+ 次）
// 改造：先按题型过滤出客观题错题，2 次批量 SQL 完成（一次错题 upsert + 一次复习卡 upsert）
func (s *ExamService) recordWrongAnswers(userID uint, answers []model.ExamRecordAnswer) {
	questionIDs := make([]uint, 0, len(answers))
	for _, a := range answers {
		if a.IsCorrect == 0 {
			questionIDs = append(questionIDs, a.QuestionID)
		}
	}
	if len(questionIDs) == 0 {
		return
	}
	questions, err := s.questionRepo.FindByIDs(questionIDs)
	if err != nil {
		return
	}
	objIDs := make([]uint, 0, len(questions))
	for _, q := range questions {
		if q.Type == model.TypeEssay {
			continue
		}
		objIDs = append(objIDs, q.ID)
	}
	if len(objIDs) == 0 {
		return
	}
	if err := s.reviewSvc.OnWrongBatch(userID, objIDs); err != nil {
		return
	}
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
			BlankOptions: q.BlankOptions,
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

// ListRecords 考试记录列表。
//
// 原实现：N 条记录 → 循环 N 次 templateRepo.FindByID 或 getAISubjectID（N+1）
// 改造：先 FindByUser 一次拿全量；用 templateRepo.FindByIDs 批量补齐模板信息；
//       AI 考试用 recordRepo.GetSubjectIDsByRecords 一次性 JOIN 反查
func (s *ExamService) ListRecords(userID uint) ([]dto.ExamRecordResp, error) {
	records, err := s.recordRepo.FindByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []dto.ExamRecordResp{}, nil
	}

	// 收集需要查的 templateID
	templateIDs := make([]uint, 0, len(records))
	aiRecordIDs := make([]uint, 0)
	for _, r := range records {
		if r.TemplateID == 0 {
			aiRecordIDs = append(aiRecordIDs, r.ID)
		} else {
			templateIDs = append(templateIDs, r.TemplateID)
		}
	}
	templates, err := s.templateRepo.FindByIDs(templateIDs)
	if err != nil {
		return nil, err
	}
	aiSubjects, err := s.recordRepo.GetSubjectIDsByRecords(aiRecordIDs)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ExamRecordResp, 0, len(records))
	for _, r := range records {
		templateName := ""
		subjectID := r.TemplateID
		if r.TemplateID == 0 {
			subjectID = aiSubjects[r.ID]
			templateName = "AI 智能组卷"
		} else if t, ok := templates[r.TemplateID]; ok {
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

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// examJanitor 持有后台 janitor 引用的 ExamService 实例
// 由 router.go 通过 SetExamJanitor 注册；janitor 启动后周期性扫描并自动交卷过期 pending 记录
var examJanitor *ExamService

// SetExamJanitor 注册 janitor 使用的 ExamService。
// 必须在 StartExamJanitor 之前调用。
func SetExamJanitor(svc *ExamService) {
	examJanitor = svc
}

// StartExamJanitor 启动后台 goroutine，每 30s 扫描一次过期 pending 考试并自动交卷。
// 仅处理 template_id > 0 的常规考试（AI 考试依赖客户端定时器 / StartExam 入口防御）。
// 首次启动延迟 5s 等待服务就绪。ctx 取消后 janitor 立刻退出（OPT-20）。
func StartExamJanitor(ctx context.Context) {
	if examJanitor == nil {
		log.Println("[exam janitor] ExamService 未注册，janitor 启动跳过")
		return
	}
	go func() {
		select {
		case <-ctx.Done():
			log.Println("[exam janitor] 启动前已收到取消信号，janitor 不再启动")
			return
		case <-time.After(5 * time.Second):
		}
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[exam janitor] 收到取消信号，janitor 退出")
				return
			case <-ticker.C:
				examJanitor.settleExpiredRecords()
			}
		}
	}()
}

// settleExpiredRecords 扫描并自动交卷所有已过期的 pending 考试。
// 分批处理：每批 limit=50 条，直至本轮无过期记录。
func (s *ExamService) settleExpiredRecords() {
	const batch = 50
	for {
		rows, err := s.recordRepo.FindExpiredPending(batch)
		if err != nil {
			log.Printf("[exam janitor] 扫描过期考试失败: %v", err)
			return
		}
		if len(rows) == 0 {
			return
		}
		for _, r := range rows {
			if _, err := s.finishExam(r.UserID, r.RecordID); err != nil {
				log.Printf("[exam janitor] 自动交卷失败 record_id=%d user_id=%d: %v",
					r.RecordID, r.UserID, err)
			} else {
				log.Printf("[exam janitor] 已自动交卷 record_id=%d user_id=%d", r.RecordID, r.UserID)
			}
		}
		// 本批已处理完；如 batch 满则继续下一批
		if len(rows) < batch {
			return
		}
	}
}
