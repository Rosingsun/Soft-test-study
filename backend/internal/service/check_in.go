package service

import (
	"errors"
	"math"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

const checkInCount = 10

// CheckInService 每日打卡
type CheckInService struct {
	checkInRepo  *repository.CheckInRepo
	questionRepo *repository.QuestionRepo
	practiceRepo *repository.PracticeRecordRepo
	reviewSvc    *ReviewService
}

func NewCheckInService(
	checkInRepo *repository.CheckInRepo,
	questionRepo *repository.QuestionRepo,
	practiceRepo *repository.PracticeRecordRepo,
	reviewSvc *ReviewService,
) *CheckInService {
	return &CheckInService{
		checkInRepo:  checkInRepo,
		questionRepo: questionRepo,
		practiceRepo: practiceRepo,
		reviewSvc:    reviewSvc,
	}
}

// Today 今日打卡：未开始生成 10 道真题，已开始返回断点进度；已完成返回题目（支持重新打卡）与最新结果
func (s *CheckInService) Today(userID uint) (*dto.CheckInTodayResp, error) {
	today := time.Now().Format("2006-01-02")

	existing, _ := s.checkInRepo.FindByUserAndDate(userID, today)
	isCompleted := existing != nil && existing.ID > 0

	rows, err := s.checkInRepo.FindQuestionsByUserAndDate(userID, today)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 && !isCompleted {
		questions, err := s.pickQuestions()
		if err != nil {
			return nil, err
		}
		date, _ := time.Parse("2006-01-02", today)
		items := make([]model.CheckInQuestion, 0, len(questions))
		for _, q := range questions {
			items = append(items, model.CheckInQuestion{
				UserID:     userID,
				CheckDate:  date,
				QuestionID: q.ID,
			})
		}
		if err := s.checkInRepo.CreateQuestions(items); err != nil {
			return nil, err
		}
		rows = items
	}

	resp := s.buildToday(userID, today, rows)
	resp.Completed = isCompleted
	resp.RetakeAvailable = isCompleted
	if isCompleted {
		resp.Result = &dto.CheckInResultResp{
			CheckDate:    existing.CheckDate.Format("2006-01-02"),
			TotalCount:   existing.TotalCount,
			CorrectCount: existing.CorrectCount,
			Accuracy:     existing.Accuracy,
			Duration:     existing.Duration,
		}
	}
	return resp, nil
}

func (s *CheckInService) pickQuestions() ([]model.Question, error) {
	real, err := s.questionRepo.FindRandomReal(checkInCount)
	if err != nil {
		return nil, err
	}
	if len(real) >= checkInCount {
		return real[:checkInCount], nil
	}
	picked := make(map[uint]bool, len(real))
	for _, q := range real {
		picked[q.ID] = true
	}
	need := checkInCount - len(real)
	if need > 0 {
		obj, _ := s.questionRepo.FindRandomObjective(checkInCount * 3)
		for _, q := range obj {
			if need <= 0 {
				break
			}
			if picked[q.ID] {
				continue
			}
			real = append(real, q)
			picked[q.ID] = true
			need--
		}
	}
	return real, nil
}

func (s *CheckInService) buildToday(userID uint, date string, rows []model.CheckInQuestion) *dto.CheckInTodayResp {
	resp := &dto.CheckInTodayResp{
		CheckDate:  date,
		TotalCount: len(rows),
		Questions:  make([]dto.CheckInQuestionResp, 0, len(rows)),
		AnswerMap:  map[uint]string{},
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.QuestionID)
	}
	qMap := map[uint]model.Question{}
	if len(ids) > 0 {
		if qs, err := s.questionRepo.FindByIDs(ids); err == nil {
			for _, q := range qs {
				qMap[q.ID] = q
			}
		}
	}
	answeredCount := 0
	for _, row := range rows {
		answered := row.Answer != ""
		if answered {
			answeredCount++
			resp.AnswerMap[row.QuestionID] = row.Answer
		}
		q, ok := qMap[row.QuestionID]
		if !ok {
			continue
		}
		resp.Questions = append(resp.Questions, dto.CheckInQuestionResp{
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
			Answered:     answered,
		})
	}
	resp.AnsweredCount = answeredCount
	return resp
}

// Answer 提交单题答案（支持断点续答与已完成后的重新打卡覆盖作答）
// 首次/断点作答：同时写入练习记录并联动错题本。
// 重打模式（今日已完成）：仅更新题目快照答案，练习记录与错题联动在 Complete 统一处理，避免中间态污染。
func (s *CheckInService) Answer(userID uint, req dto.CheckInAnswerReq) (*dto.PracticeRecordResp, error) {
	today := time.Now().Format("2006-01-02")

	rows, err := s.checkInRepo.FindQuestionsByUserAndDate(userID, today)
	if err != nil || len(rows) == 0 {
		return nil, errors.New("今日打卡尚未开始，请先获取题目")
	}

	existing, _ := s.checkInRepo.FindByUserAndDate(userID, today)
	isRetake := existing != nil && existing.ID > 0

	var row *model.CheckInQuestion
	for i := range rows {
		if rows[i].QuestionID == req.QuestionID {
			row = &rows[i]
			break
		}
	}
	if row == nil {
		return nil, errors.New("题目不属于今日打卡")
	}

	q, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, err
	}
	isCorrect := 0
	if IsAnswerCorrect(q.Type, q.Answer, req.Answer) {
		isCorrect = 1
	}
	if err := s.checkInRepo.UpdateQuestionAnswer(row.ID, req.Answer, isCorrect, req.Duration); err != nil {
		return nil, err
	}

	resp := &dto.PracticeRecordResp{
		QuestionID: req.QuestionID,
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Duration:   req.Duration,
	}

	// 重打模式：不逐题写练习记录/错题，交由 Complete 统一处理
	if isRetake {
		return resp, nil
	}

	record := &model.PracticeRecord{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Mode:       "checkin",
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Duration:   req.Duration,
	}
	if err := s.practiceRepo.Create(record); err != nil {
		return nil, err
	}
	resp.ID = record.ID
	resp.Mode = record.Mode

	if isCorrect == 0 {
		if err := s.reviewSvc.OnWrong(userID, req.QuestionID); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// Complete 完成今日打卡
// 首次完成：创建汇总行并写入首次快照（rank_*，排名/个人平均口径）。
// 重新打卡：清理当日旧练习记录并统一写入本次作答，同步错题/复习卡，仅更新最新数据（快照保持不变）。
func (s *CheckInService) Complete(userID uint) (*dto.CheckInResultResp, error) {
	today := time.Now().Format("2006-01-02")

	existing, _ := s.checkInRepo.FindByUserAndDate(userID, today)
	isRetake := existing != nil && existing.ID > 0

	rows, err := s.checkInRepo.FindQuestionsByUserAndDate(userID, today)
	if err != nil || len(rows) == 0 {
		return nil, errors.New("今日打卡尚未开始")
	}

	total := len(rows)
	correct := 0
	duration := 0
	for _, row := range rows {
		if row.Answer == "" {
			return nil, errors.New("还有题目未作答，无法完成打卡")
		}
		if row.IsCorrect == 1 {
			correct++
		}
		duration += row.Duration
	}
	accuracy := 0.0
	if total > 0 {
		accuracy = math.Round(float64(correct)/float64(total)*10000) / 100
	}

	// 重新打卡：清理旧练习记录，统一写入本次作答并同步错题/复习卡
	if isRetake {
		if err := s.practiceRepo.DeleteCheckInRecordsByUserAndDate(userID, today); err != nil {
			return nil, err
		}
		for _, row := range rows {
			record := &model.PracticeRecord{
				UserID:     userID,
				QuestionID: row.QuestionID,
				Mode:       "checkin",
				Answer:     row.Answer,
				IsCorrect:  row.IsCorrect,
				Duration:   row.Duration,
			}
			if err := s.practiceRepo.Create(record); err != nil {
				return nil, err
			}
			if row.IsCorrect == 1 {
				if err := s.reviewSvc.OnCorrect(userID, row.QuestionID); err != nil {
					return nil, err
				}
			} else {
				if err := s.reviewSvc.OnWrong(userID, row.QuestionID); err != nil {
					return nil, err
				}
			}
		}
		if err := s.checkInRepo.UpdateSummary(existing.ID, correct, accuracy, duration); err != nil {
			return nil, err
		}
		return &dto.CheckInResultResp{
			CheckDate:    existing.CheckDate.Format("2006-01-02"),
			TotalCount:   total,
			CorrectCount: correct,
			Accuracy:     accuracy,
			Duration:     duration,
		}, nil
	}

	// 首次完成：创建汇总行（含首次快照）
	date, _ := time.Parse("2006-01-02", today)
	ci := &model.CheckIn{
		UserID:          userID,
		CheckDate:       date,
		TotalCount:      total,
		CorrectCount:    correct,
		Accuracy:        accuracy,
		Duration:        duration,
		Status:          1,
		RankCorrectCount: correct,
		RankAccuracy:     accuracy,
		RankDuration:     duration,
	}
	if err := s.checkInRepo.Create(ci); err != nil {
		// 并发唯一冲突：已有行则转为重打更新最新数据
		if ex, e2 := s.checkInRepo.FindByUserAndDate(userID, today); e2 == nil && ex.ID > 0 {
			if err := s.checkInRepo.UpdateSummary(ex.ID, correct, accuracy, duration); err != nil {
				return nil, err
			}
			return &dto.CheckInResultResp{
				CheckDate:    ex.CheckDate.Format("2006-01-02"),
				TotalCount:   total,
				CorrectCount: correct,
				Accuracy:     accuracy,
				Duration:     duration,
			}, nil
		}
		return nil, err
	}
	return &dto.CheckInResultResp{
		CheckDate:    ci.CheckDate.Format("2006-01-02"),
		TotalCount:   ci.TotalCount,
		CorrectCount: ci.CorrectCount,
		Accuracy:     ci.Accuracy,
		Duration:     ci.Duration,
	}, nil
}

// Reset 重置今日打卡的作答（清空已答答案，保留题目快照），用于未完成态下的"重新打卡"
func (s *CheckInService) Reset(userID uint) error {
	today := time.Now().Format("2006-01-02")
	return s.checkInRepo.ClearAnswersByUserAndDate(userID, today)
}

// History 指定月份的打卡历史
func (s *CheckInService) History(userID uint, month string) ([]dto.CheckInHistoryResp, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}
	startDate := start.Format("2006-01-02")
	endDate := start.AddDate(0, 1, -1).Format("2006-01-02")

	list, err := s.checkInRepo.FindHistory(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.CheckInHistoryResp, 0, len(list))
	for _, c := range list {
		resp = append(resp, dto.CheckInHistoryResp{
			Date:         c.CheckDate.Format("2006-01-02"),
			TotalCount:   c.TotalCount,
			CorrectCount: c.CorrectCount,
			Accuracy:     c.Accuracy,
		})
	}
	return resp, nil
}

// Stats 打卡统计：连续/最高连续/累计/平均正确率
func (s *CheckInService) Stats(userID uint) (*dto.CheckInStatsResp, error) {
	dates, err := s.checkInRepo.FindDatesByUser(userID)
	if err != nil {
		return nil, err
	}
	current, max := calcStreaks(dates)

	total, err := s.checkInRepo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	avg, err := s.checkInRepo.AvgAccuracyByUser(userID)
	if err != nil {
		return nil, err
	}

	todayDone := false
	if len(dates) > 0 {
		todayStr := time.Now().Format("2006-01-02")
		if dates[len(dates)-1] == todayStr {
			todayDone = true
		}
	}

	return &dto.CheckInStatsResp{
		CurrentStreak: current,
		MaxStreak:     max,
		TotalDays:     int(total),
		AvgAccuracy:   avg,
		TodayDone:     todayDone,
	}, nil
}

// calcStreaks 计算当前连续与最高连续打卡天数（dates 升序）
func calcStreaks(dates []string) (current, max int) {
	set := make(map[string]bool, len(dates))
	for _, d := range dates {
		set[d] = true
	}

	run := 0
	var prev time.Time
	for _, ds := range dates {
		d, err := time.Parse("2006-01-02", ds)
		if err != nil {
			continue
		}
		if !prev.IsZero() && d.Sub(prev).Hours() == 24 {
			run++
		} else {
			run = 1
		}
		if run > max {
			max = run
		}
		prev = d
	}

	cur := time.Now()
	if !set[cur.Format("2006-01-02")] {
		cur = cur.AddDate(0, 0, -1)
	}
	for set[cur.Format("2006-01-02")] {
		current++
		cur = cur.AddDate(0, 0, -1)
	}
	return current, max
}
