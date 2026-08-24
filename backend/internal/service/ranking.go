package service

import (
	"math"
	"slices"
	"sort"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

const (
	rankPracticePriorCorrect = 25.0
	rankPracticePriorTotal   = 50.0
	rankExamScoreHalfLifeDay = 30.0
	rankExamTotalScore       = 75.0
)

// 分项预估（选择题 / 案例分析 / 论文）
const (
	sectionMaxScore     = 75.0
	sectionPriorCorrect = 10.0 // 贝叶斯平滑先验正确数
	sectionPriorTotal   = 20.0 // 贝叶斯平滑先验总题数
	sectionMinSample    = 10   // 低于此值标记 sufficient=false
	essayHalfLifeDay    = 30.0 // 论文评分时间加权半衰期
	sectionTotalMax     = 225.0
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

// findMyRank 寻找 my 的竞赛名次。
// items 必须按 (SortVal DESC, SubVal DESC) 排序。
// 返回 (rank, myValue, found)；未找到 userID 时 found=false。
// OPT-24：原实现先 O(n) 找 my、再 O(n) 算 rank，两次线性扫描；
// 本实现用 slices.IndexFunc 一次定位 my 在已排序数组中的下标，下标 +1 即为名次。
// （同一 SortVal/SubVal 组合按数组顺序区分名次，与旧行为一致。）
func findMyRank(items []rankValue, userID uint) (rank int, myValue float64, found bool) {
	idx := slices.IndexFunc(items, func(it rankValue) bool {
		return it.UserID == userID
	})
	if idx < 0 {
		return 0, 0, false
	}
	return idx + 1, items[idx].Value, true
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

	// 我的名次（OPT-24：slices.IndexFunc 一次定位，下标 +1 即名次）
	rank, myValue, found := findMyRank(items, userID)
	if found {
		percentile := 0
		if participants > 0 {
			percentile = int(math.Round(float64(rank) / float64(participants) * 100))
		}
		resp.My = &dto.RankingMyResp{
			Rank:              rank,
			Value:             myValue,
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

// ========================================================================
// 分项预估（选择题 / 案例分析 / 论文）
// ========================================================================

// sectionChoiceTypes 选择题对应的客观题型（不含 case_study / essay）
var sectionChoiceTypes = []string{
	model.TypeSingle, model.TypeMulti, model.TypeJudge, model.TypeFill,
	model.TypeShort, model.TypeComprehensive, model.TypeMultiBlank,
}

// sectionConfig 各 section 的配置
type sectionConfig struct {
	Code     string   // choice / case_study / essay
	Name     string   // 展示名
	Types    []string // 关联题型（仅客观题使用）
	FromEssay bool     // 是否走 essay_scores 表
}

// GetEstimatedSectionScores 计算用户在指定科目下的三项预估分（满分 225）
func (s *RankingService) GetEstimatedSectionScores(userID, subjectID uint) (*dto.EstimatedScoreResp, error) {
	sections := []sectionConfig{
		{Code: "choice", Name: "选择题", Types: sectionChoiceTypes},
		{Code: "case_study", Name: "案例分析", Types: []string{model.TypeCaseStudy}},
		{Code: "essay", Name: "论文", FromEssay: true},
	}

	// 拉取该科目下所有章节（含 weight）作为权重基准
	chapterMap, err := s.loadChaptersForSubject(subjectID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SectionBreakdown, 0, len(sections))
	for _, cfg := range sections {
		if cfg.FromEssay {
			result = append(result, s.computeEssaySection(userID, subjectID, cfg))
			continue
		}
		result = append(result, s.computeObjectiveSection(userID, subjectID, cfg, chapterMap))
	}

	total := 0.0
	for _, sec := range result {
		total += sec.EstScore
	}
	total = math.Round(total*100) / 100

	return &dto.EstimatedScoreResp{
		SubjectID:   subjectID,
		TotalEst:    total,
		TotalMax:    sectionTotalMax,
		Sections:    result,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// loadChaptersForSubject 拉取该科目的全部章节权重（subjectID=0 时返回空 map）
func (s *RankingService) loadChaptersForSubject(subjectID uint) (map[uint]float64, error) {
	if subjectID <= 0 {
		return map[uint]float64{}, nil
	}
	rows, err := s.repo.SectionChapters(subjectID)
	if err != nil {
		return nil, err
	}
	out := make(map[uint]float64, len(rows))
	for _, r := range rows {
		w := r.Weight
		if w <= 0 {
			w = 1.0
		}
		out[r.ChapterID] = w
	}
	return out, nil
}

// computeObjectiveSection 客观题分项（选择题/案例分析）
func (s *RankingService) computeObjectiveSection(userID, subjectID uint, cfg sectionConfig, chapterWeights map[uint]float64) dto.SectionBreakdown {
	// 合并 练习 + 模考 的章节答题数
	practiceRows, _ := s.repo.SectionPractice(userID, subjectID, cfg.Types)
	examRows, _ := s.repo.SectionExam(userID, subjectID, cfg.Types)

	merged := make(map[uint]*struct {
		Total   int64
		Correct int64
	}, len(practiceRows)+len(examRows))
	for _, r := range practiceRows {
		v, ok := merged[r.ChapterID]
		if !ok {
			v = &struct {
				Total   int64
				Correct int64
			}{}
			merged[r.ChapterID] = v
		}
		v.Total += r.Total
		v.Correct += r.Correct
	}
	for _, r := range examRows {
		v, ok := merged[r.ChapterID]
		if !ok {
			v = &struct {
				Total   int64
				Correct int64
			}{}
			merged[r.ChapterID] = v
		}
		v.Total += r.Total
		v.Correct += r.Correct
	}

	// 拉取章节名映射
	chapterNames, _ := s.repo.SectionChapterNames(subjectID)

	// 收集所有出现过的章节（含未答的以便完整展示权重）
	chapterSet := make(map[uint]struct{}, len(merged))
	for cid := range merged {
		chapterSet[cid] = struct{}{}
	}
	for cid := range chapterWeights {
		chapterSet[cid] = struct{}{}
	}

	// 计算加权预估分
	var weightSum, weightedSum float64
	totalAttempted, totalCorrect := int64(0), int64(0)
	chapterItems := make([]dto.SectionChapterItem, 0, len(chapterSet))
	for cid := range chapterSet {
		w := chapterWeights[cid]
		if w <= 0 {
			// 未配置权重的章节：仅在有答题数据时纳入，权重按 1.0
			if _, hasData := merged[cid]; hasData {
				w = 1.0
			} else {
				continue
			}
		}
		m := merged[cid]
		var attempted, correct int64
		var rawAccuracy float64
		if m != nil {
			attempted = m.Total
			correct = m.Correct
			if attempted > 0 {
				rawAccuracy = float64(correct) / float64(attempted) * 100
			}
		}

		// 贝叶斯平滑正确率
		smoothed := (float64(correct) + sectionPriorCorrect) /
			(float64(attempted) + sectionPriorTotal)

		// 该章节对该 section 的贡献分（加权后）
		est := smoothed * w * sectionMaxScore
		weightSum += w
		weightedSum += smoothed * w

		chapterItems = append(chapterItems, dto.SectionChapterItem{
			ChapterID:   cid,
			ChapterName: chapterNames[cid],
			Weight:      w,
			Attempted:   int(attempted),
			Correct:     int(correct),
			Accuracy:    math.Round(rawAccuracy*100) / 100,
			EstScore:    math.Round(est*100) / 100,
		})

		totalAttempted += attempted
		totalCorrect += correct
	}

	// 排序：按 est_score DESC
	sort.Slice(chapterItems, func(i, j int) bool {
		if chapterItems[i].EstScore != chapterItems[j].EstScore {
			return chapterItems[i].EstScore > chapterItems[j].EstScore
		}
		return chapterItems[i].ChapterID < chapterItems[j].ChapterID
	})

	// section 总分：Σ(章节贝叶斯正确率 × weight) / Σ(weight) × 75
	var sectionScore float64
	if weightSum > 0 {
		sectionScore = weightedSum / weightSum * sectionMaxScore
	}
	sectionScore = math.Round(sectionScore*100) / 100

	accuracy := 0.0
	if totalAttempted > 0 {
		accuracy = math.Round(float64(totalCorrect)/float64(totalAttempted)*10000) / 100
	}

	// 若没有任何权重配置（subjectID=0），回退为整体累计
	if weightSum == 0 && totalAttempted > 0 {
		rawRate := float64(totalCorrect) / float64(totalAttempted)
		sectionScore = math.Round(rawRate*sectionMaxScore*100) / 100
	}

	return dto.SectionBreakdown{
		Section:     cfg.Code,
		SectionName: cfg.Name,
		MaxScore:    sectionMaxScore,
		Attempted:   int(totalAttempted),
		EstScore:    sectionScore,
		Accuracy:    accuracy,
		SampleSize:  int(totalAttempted),
		Sufficient:  int(totalAttempted) >= sectionMinSample,
		Chapters:    chapterItems,
	}
}

// computeEssaySection 论文分项（基于 essay_scores 表时间加权）
func (s *RankingService) computeEssaySection(userID, subjectID uint, cfg sectionConfig) dto.SectionBreakdown {
	rows, err := s.repo.UserEssayScores(userID, subjectID)
	if err != nil {
		rows = nil
	}
	now := time.Now()
	var weighted, wSum float64
	for _, r := range rows {
		if r.TotalScore < 0 {
			continue
		}
		score := float64(r.TotalScore)
		if score > sectionMaxScore {
			score = sectionMaxScore
		}
		days := now.Sub(r.CreatedAt).Hours() / 24
		if days < 0 {
			days = 0
		}
		w := math.Pow(0.5, days/essayHalfLifeDay)
		weighted += score * w
		wSum += w
	}
	est := 0.0
	if wSum > 0 {
		est = weighted / wSum
	}
	est = math.Round(est*100) / 100
	accuracy := 0.0
	if wSum > 0 && est > 0 {
		accuracy = math.Round(est/sectionMaxScore*10000) / 100
	}
	return dto.SectionBreakdown{
		Section:     cfg.Code,
		SectionName: cfg.Name,
		MaxScore:    sectionMaxScore,
		Attempted:   len(rows),
		EstScore:    est,
		Accuracy:    accuracy,
		SampleSize:  len(rows),
		Sufficient:  len(rows) >= 1,
		Chapters:    []dto.SectionChapterItem{},
	}
}
