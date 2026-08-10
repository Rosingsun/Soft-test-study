package service

import (
	"errors"
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

const (
	reviewMaxInterval = 180
	reviewEfMin       = 1.3
	reviewEfMax       = 3.0
)

// ReviewService 遗忘曲线复习（SM-2 算法）
type ReviewService struct {
	cardRepo     *repository.ReviewCardRepo
	questionRepo *repository.QuestionRepo
	practiceRepo *repository.PracticeRecordRepo
	wrongRepo    *repository.WrongQuestionRepo
}

func NewReviewService(
	cardRepo *repository.ReviewCardRepo,
	questionRepo *repository.QuestionRepo,
	practiceRepo *repository.PracticeRecordRepo,
	wrongRepo *repository.WrongQuestionRepo,
) *ReviewService {
	return &ReviewService{
		cardRepo:     cardRepo,
		questionRepo: questionRepo,
		practiceRepo: practiceRepo,
		wrongRepo:    wrongRepo,
	}
}

// OnWrong 错题联动统一入口：收录错题本 + 确保/重置 SM-2 复习卡片
func (s *ReviewService) OnWrong(userID, questionID uint) error {
	if err := s.wrongRepo.Upsert(userID, questionID); err != nil {
		return err
	}
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	card, err := s.cardRepo.FindByUserAndQuestion(userID, questionID)
	if err == nil && card.ID > 0 {
		card.Repetition = 0
		card.IntervalDays = 1
		card.DueDate = today
		card.Status = 1
		return s.cardRepo.Save(card)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.cardRepo.Create(&model.ReviewCard{
		UserID:       userID,
		QuestionID:   questionID,
		Repetition:   0,
		IntervalDays: 1,
		EaseFactor:   2.5,
		DueDate:      today,
		Status:       1,
	})
}

// OnWrongBatch 批量收录错题。
//
// 原实现：N 道错题 → 循环 N 次 OnWrong，每次内部 2 次 DB 往返
//        75 题全错 = 150+ 次 RTT
// 改造：2 次批量 upsert（错题本 + 复习卡），固定 2 次 RTT
func (s *ReviewService) OnWrongBatch(userID uint, questionIDs []uint) error {
	if len(questionIDs) == 0 {
		return nil
	}
	if err := s.wrongRepo.UpsertBatch(userID, questionIDs); err != nil {
		return err
	}
	return s.cardRepo.UpsertBatch(userID, questionIDs)
}

// OnCorrect 重新打卡答对时解除错题联动：从错题本移除该题，并将复习卡片置为已掌握
func (s *ReviewService) OnCorrect(userID, questionID uint) error {
	if err := s.wrongRepo.Delete(userID, questionID); err != nil {
		return err
	}
	card, err := s.cardRepo.FindByUserAndQuestion(userID, questionID)
	if err == nil && card.ID > 0 {
		card.Status = 0
		card.Repetition = 0
		card.IntervalDays = 1
		return s.cardRepo.Save(card)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

// Today 今日待复习列表
func (s *ReviewService) Today(userID uint) (*dto.ReviewTodayResp, error) {
	today := time.Now().Format("2006-01-02")
	cards, err := s.cardRepo.FindDueByUser(userID, today, 200)
	if err != nil {
		return nil, err
	}

	ids := make([]uint, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.QuestionID)
	}
	qMap := map[uint]model.Question{}
	if len(ids) > 0 {
		qs, err := s.questionRepo.FindByIDs(ids)
		if err != nil {
			return nil, err
		}
		for _, q := range qs {
			qMap[q.ID] = q
		}
	}

	questions := make([]dto.ReviewCardResp, 0, len(cards))
	for _, c := range cards {
		q, ok := qMap[c.QuestionID]
		if !ok {
			continue
		}
		questions = append(questions, dto.ReviewCardResp{
			CardID:       c.ID,
			QuestionID:   q.ID,
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
			Repetition:   c.Repetition,
			IntervalDays: c.IntervalDays,
			DueDate:      c.DueDate.Format("2006-01-02"),
		})
	}
	return &dto.ReviewTodayResp{Total: len(questions), Questions: questions}, nil
}

// Answer 提交复习答题：判分 + 更新 SM-2 + 联动错题本
func (s *ReviewService) Answer(userID uint, req dto.ReviewAnswerReq) (*dto.ReviewAnswerResp, error) {
	q, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, err
	}
	card, err := s.cardRepo.FindByUserAndQuestion(userID, req.QuestionID)
	if err != nil || card.ID == 0 {
		return nil, ErrNotFound
	}
	if card.Status != 1 {
		return nil, errors.New("该题已掌握，无需复习")
	}

	isCorrect := 0
	if IsAnswerCorrect(q.Type, q.Answer, req.Answer) {
		isCorrect = 1
	}

	record := &model.PracticeRecord{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Mode:       "review",
		Answer:     req.Answer,
		IsCorrect:  isCorrect,
		Duration:   req.Duration,
	}
	if err := s.practiceRepo.Create(record); err != nil {
		return nil, err
	}

	mastered := false
	if isCorrect == 1 {
		applySM2(card, 5)
		if err := s.wrongRepo.IncrementCorrect(userID, req.QuestionID); err != nil {
			return nil, err
		}
		if wq, e := s.wrongRepo.FindByUserAndQuestion(userID, req.QuestionID); e == nil && wq.CorrectCount >= 2 {
			s.wrongRepo.Delete(userID, req.QuestionID)
			card.Status = 0
			mastered = true
		}
	} else {
		applySM2(card, 2)
		if err := s.wrongRepo.Upsert(userID, req.QuestionID); err != nil {
			return nil, err
		}
		card.Status = 1
	}

	now := time.Now()
	card.LastReviewedAt = &now
	if err := s.cardRepo.Save(card); err != nil {
		return nil, err
	}

	return &dto.ReviewAnswerResp{
		QuestionID:  req.QuestionID,
		IsCorrect:   isCorrect,
		CorrectAns:  q.Answer,
		Analysis:    q.Analysis,
		Content:     q.Content,
		Options:     q.Options,
		BlankOptions: q.BlankOptions,
		Type:        q.Type,
		NextDueDate: card.DueDate.Format("2006-01-02"),
		Mastered:    mastered,
	}, nil
}

// Overview 复习概览。
//
// 原实现：4 次 review_card count + 1 次 practice count = 5 次 RTT
// 改造：reviewCardRepo.Stats 一次聚合拿 4 项，practice count 1 次 = 2 次 RTT
func (s *ReviewService) Overview(userID uint) (*dto.ReviewOverviewResp, error) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	tomorrowStr := today.AddDate(0, 0, 1).Format("2006-01-02")

	stats, err := s.cardRepo.Stats(userID, todayStr, tomorrowStr)
	if err != nil {
		return nil, err
	}

	totalReviews, err := s.practiceRepo.CountByMode(userID, "review")
	if err != nil {
		return nil, err
	}

	return &dto.ReviewOverviewResp{
		DueToday:     stats.DueToday,
		DueTomorrow:  stats.DueTomorrow,
		Mastered:     stats.Mastered,
		Reviewing:    stats.Reviewing,
		TotalReviews: totalReviews,
	}, nil
}

// applySM2 SM-2 算法：quality>=3 为答对（q=5），否则答错（q=2）
func applySM2(card *model.ReviewCard, quality int) {
	ef := card.EaseFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if ef < reviewEfMin {
		ef = reviewEfMin
	}
	if ef > reviewEfMax {
		ef = reviewEfMax
	}

	if quality >= 3 {
		if card.Repetition == 0 {
			card.IntervalDays = 1
		} else if card.Repetition == 1 {
			card.IntervalDays = 6
		} else {
			card.IntervalDays = int(math.Round(float64(card.IntervalDays) * ef))
		}
		card.Repetition++
	} else {
		card.Repetition = 0
		card.IntervalDays = 1
	}
	card.EaseFactor = ef

	if card.IntervalDays > reviewMaxInterval {
		card.IntervalDays = reviewMaxInterval
	}
	if card.IntervalDays < 1 {
		card.IntervalDays = 1
	}
	card.DueDate = time.Now().AddDate(0, 0, card.IntervalDays)
}
