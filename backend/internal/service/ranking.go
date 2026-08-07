package service

import (
	"math"
	"sort"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

const (
	rankPracticePriorCorrect = 25.0
	rankPracticePriorTotal   = 50.0
	rankExamScoreHalfLifeDay = 30.0
	rankExamTotalScore       = 75.0
)

// RankingService 成绩排行（三类榜单 × 双维度）
type RankingService struct {
	repo     *repository.RankingRepo
	userRepo *repository.UserRepo
}

func NewRankingService(repo *repository.RankingRepo, userRepo *repository.UserRepo) *RankingService {
	return &RankingService{repo: repo, userRepo: userRepo}
}

type rankValue struct {
	UserID  uint
	Value   float64
	SortVal float64 // 参与排名的排序值（可与展示值不同，如平滑正确率）
	SubVal  float64 // 次要排序值（如打卡总天数）
}

// Get 获取排行：仅返回本人名次 + 匿名统计线
func (s *RankingService) Get(userID uint, category, metric string, subjectID uint) (*dto.RankingResp, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	levelID := user.LevelID

	if !validRankingMetric(category, metric) {
		return nil, ErrConflict
	}

	items, err := s.buildItems(category, metric, levelID, subjectID)
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].SortVal != items[j].SortVal {
			return items[i].SortVal > items[j].SortVal
		}
		return items[i].SubVal > items[j].SubVal
	})

	totalUsers, err := s.repo.TotalUsersByLevel(levelID)
	if err != nil {
		return nil, err
	}

	resp := &dto.RankingResp{
		Category: category,
		Metric:   metric,
		Scope: dto.RankingScopeResp{
			LevelID:   levelID,
			SubjectID: subjectID,
			TotalUsers: totalUsers,
		},
		Reference: dto.RankingReferenceResp{},
	}

	participants := len(items)
	if participants == 0 {
		return resp, nil
	}

	// 我的名次（竞赛排名：并列同 rank）
	var my *rankValue
	for _, it := range items {
		if it.UserID == userID {
			my = &it
			break
		}
	}
	if my != nil {
		rank := 1
		for _, it := range items {
			if it.SortVal > my.SortVal {
				rank++
			} else if it.SortVal == my.SortVal && it.SubVal > my.SubVal {
				rank++
			} else {
				break
			}
		}
		percentile := 0
		if participants > 0 {
			percentile = int(math.Round(float64(rank) / float64(participants) * 100))
		}
		resp.My = &dto.RankingMyResp{
			Rank:              rank,
			Value:             my.Value,
			TotalParticipants: participants,
			Percentile:        percentile,
		}
	}

	// 匿名统计线
	sum := 0.0
	for _, it := range items {
		sum += it.Value
	}
	resp.Reference.Avg = math.Round(sum/float64(participants)*100) / 100

	topCount := int(math.Ceil(float64(participants) * 0.10))
	if topCount < 1 {
		topCount = 1
	}
	resp.Reference.Top10Threshold = items[topCount-1].Value

	// 各阶段人员分布（可视化表格）
	myValue := -1.0
	if my != nil {
		myValue = my.Value
	}
	resp.Distribution = buildDistribution(category, metric, items, myValue)

	return resp, nil
}

// distributionBucket 分段定义
type distributionBucket struct {
	Label  string
	In     func(v float64) bool // 是否属于该分段
}

// buildDistribution 将参与者按 metric 取值划分为若干分数段，统计各段人数占比
func buildDistribution(category, metric string, items []rankValue, myValue float64) []dto.DistributionItem {
	var buckets []distributionBucket
	switch metric {
	case "streak":
		buckets = []distributionBucket{
			{Label: "0-2 天", In: func(v float64) bool { return v >= 0 && v < 3 }},
			{Label: "3-7 天", In: func(v float64) bool { return v >= 3 && v < 8 }},
			{Label: "8-14 天", In: func(v float64) bool { return v >= 8 && v < 15 }},
			{Label: "15-30 天", In: func(v float64) bool { return v >= 15 && v < 31 }},
			{Label: "30+ 天", In: func(v float64) bool { return v >= 31 }},
		}
	case "score":
		if category == "exam" || category == "practice" {
			buckets = []distributionBucket{
				{Label: "0-30 分", In: func(v float64) bool { return v >= 0 && v < 30 }},
				{Label: "30-45 分", In: func(v float64) bool { return v >= 30 && v < 45 }},
				{Label: "45-60 分", In: func(v float64) bool { return v >= 45 && v < 60 }},
				{Label: "60-70 分", In: func(v float64) bool { return v >= 60 && v < 70 }},
				{Label: "70-75 分", In: func(v float64) bool { return v >= 70 && v <= 75 }},
			}
		} else {
			// 兜底：默认按 0-100 分五等分
			buckets = []distributionBucket{
				{Label: "0-20 分", In: func(v float64) bool { return v >= 0 && v < 20 }},
				{Label: "20-40 分", In: func(v float64) bool { return v >= 20 && v < 40 }},
				{Label: "40-60 分", In: func(v float64) bool { return v >= 40 && v < 60 }},
				{Label: "60-80 分", In: func(v float64) bool { return v >= 60 && v < 80 }},
				{Label: "80-100 分", In: func(v float64) bool { return v >= 80 && v <= 100 }},
			}
		}
	default: // accuracy 正确率 0-100%
		buckets = []distributionBucket{
			{Label: "0-20%", In: func(v float64) bool { return v >= 0 && v < 20 }},
			{Label: "20-40%", In: func(v float64) bool { return v >= 20 && v < 40 }},
			{Label: "40-60%", In: func(v float64) bool { return v >= 40 && v < 60 }},
			{Label: "60-80%", In: func(v float64) bool { return v >= 60 && v < 80 }},
			{Label: "80-100%", In: func(v float64) bool { return v >= 80 && v <= 100 }},
		}
	}

	total := len(items)
	result := make([]dto.DistributionItem, 0, len(buckets))
	for _, b := range buckets {
		count := 0
		for _, it := range items {
			if b.In(it.Value) {
				count++
			}
		}
		// 单独判断当前用户是否落在该分段
		myIn := myValue >= 0 && b.In(myValue)
		ratio := 0.0
		if total > 0 {
			ratio = math.Round(float64(count)/float64(total)*10000) / 10000
		}
		result = append(result, dto.DistributionItem{
			Label:  b.Label,
			Count:  count,
			Ratio:  ratio,
			IsMine: myIn,
		})
	}
	return result
}

func validRankingMetric(category, metric string) bool {
	switch category {
	case "checkin":
		return metric == "streak" || metric == "accuracy"
	case "exam":
		return metric == "accuracy" || metric == "score"
	case "practice":
		return metric == "accuracy" || metric == "score"
	}
	return false
}

func (s *RankingService) buildItems(category, metric string, levelID, subjectID uint) ([]rankValue, error) {
	switch category {
	case "checkin":
		if metric == "streak" {
			return s.checkinStreaks(levelID)
		}
		return s.checkinAccuracy(levelID)
	case "exam":
		if metric == "score" {
			return s.examScores(levelID, subjectID)
		}
		return s.examAccuracy(levelID, subjectID)
	case "practice":
		return s.practice(levelID, subjectID, metric)
	}
	return nil, nil
}

func (s *RankingService) checkinStreaks(levelID uint) ([]rankValue, error) {
	rows, err := s.repo.CheckinDates(levelID)
	if err != nil {
		return nil, err
	}
	byUser := map[uint][]string{}
	for _, row := range rows {
		byUser[row.UserID] = append(byUser[row.UserID], row.CheckDate)
	}
	items := make([]rankValue, 0, len(byUser))
	for uid, dates := range byUser {
		_, max := calcStreaks(dates)
		items = append(items, rankValue{UserID: uid, Value: float64(max), SortVal: float64(max), SubVal: float64(len(dates))})
	}
	return items, nil
}

func (s *RankingService) checkinAccuracy(levelID uint) ([]rankValue, error) {
	list, err := s.repo.CheckinAccuracy(levelID)
	if err != nil {
		return nil, err
	}
	items := make([]rankValue, 0, len(list))
	for _, it := range list {
		v := math.Round(it.Value*100) / 100
		items = append(items, rankValue{UserID: it.UserID, Value: v, SortVal: v})
	}
	return items, nil
}

func (s *RankingService) examAccuracy(levelID, subjectID uint) ([]rankValue, error) {
	list, err := s.repo.ExamAccuracy(levelID, subjectID)
	if err != nil {
		return nil, err
	}
	items := make([]rankValue, 0, len(list))
	for _, it := range list {
		v := math.Round(it.Value*100) / 100
		items = append(items, rankValue{UserID: it.UserID, Value: v, SortVal: v})
	}
	return items, nil
}

// examScores 时间加权预估分：w = 0.5^(距今天数/30)
func (s *RankingService) examScores(levelID, subjectID uint) ([]rankValue, error) {
	rows, err := s.repo.ExamScores(levelID, subjectID)
	if err != nil {
		return nil, err
	}
	byUser := map[uint][]repository.ExamScoreRow{}
	for _, row := range rows {
		byUser[row.UserID] = append(byUser[row.UserID], row)
	}
	now := time.Now()
	items := make([]rankValue, 0, len(byUser))
	for uid, records := range byUser {
		weighted := 0.0
		wSum := 0.0
		for _, r := range records {
			days := now.Sub(r.FinishedAt).Hours() / 24
			if days < 0 {
				days = 0
			}
			w := math.Pow(0.5, days/rankExamScoreHalfLifeDay)
			weighted += float64(r.Score) * w
			wSum += w
		}
		if wSum <= 0 {
			continue
		}
		v := math.Round(weighted/wSum*100) / 100
		items = append(items, rankValue{UserID: uid, Value: v, SortVal: v})
	}
	return items, nil
}

// practice 训练榜：正确率（平滑排序/原始展示）或预估分（75×平滑）
func (s *RankingService) practice(levelID, subjectID uint, metric string) ([]rankValue, error) {
	rows, err := s.repo.PracticeAccuracy(levelID, subjectID)
	if err != nil {
		return nil, err
	}
	items := make([]rankValue, 0, len(rows))
	for _, row := range rows {
		smoothed := (float64(row.Correct) + rankPracticePriorCorrect) /
			(float64(row.Total) + rankPracticePriorTotal) * 100
		if metric == "score" {
			if row.Total < 10 {
				continue // 答题量不足不参与预估分排名
			}
			v := math.Round(smoothed/100*rankExamTotalScore*100) / 100
			items = append(items, rankValue{UserID: row.UserID, Value: v, SortVal: v})
		} else {
			raw := 0.0
			if row.Total > 0 {
				raw = float64(row.Correct) / float64(row.Total) * 100
			}
			items = append(items, rankValue{
				UserID:  row.UserID,
				Value:   math.Round(raw*100) / 100,
				SortVal: math.Round(smoothed*100) / 100,
			})
		}
	}
	return items, nil
}
