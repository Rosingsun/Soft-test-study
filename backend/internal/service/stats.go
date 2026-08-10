package service

import (
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/repository"
)

type StatsService struct {
	repo        *repository.StatsRepo
	subjectRepo *repository.SubjectRepo
}

func NewStatsService(
	repo *repository.StatsRepo,
	subjectRepo *repository.SubjectRepo,
) *StatsService {
	return &StatsService{repo: repo, subjectRepo: subjectRepo}
}

func (s *StatsService) GetOverview(userID uint) (*dto.StatsOverviewResp, error) {
	ov, err := s.repo.Overview(userID)
	if err != nil {
		return nil, err
	}

	resp := &dto.StatsOverviewResp{
		TotalPracticed: ov.TotalPracticed,
		TotalCorrect:   ov.TotalCorrect,
		TotalExams:     ov.TotalExams,
		AvgExamScore:   ov.AvgExamScore,
		WrongCount:     ov.WrongCount,
		StudyDays:      ov.StudyDays,
	}
	if ov.TotalPracticed > 0 {
		resp.Accuracy = float64(ov.TotalCorrect) / float64(ov.TotalPracticed) * 100
	}
	return resp, nil
}

func (s *StatsService) GetDailyStats(userID uint, days int) ([]dto.DailyStatsResp, error) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	list, err := s.repo.Daily(userID, startDate)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.DailyStatsResp, 0, len(list))
	for _, r := range list {
		resp = append(resp, dto.DailyStatsResp{
			Date:           r.Date,
			TotalCount:     r.TotalCount,
			CorrectCount:   r.CorrectCount,
			IncorrectCount: r.TotalCount - r.CorrectCount,
		})
	}
	return resp, nil
}

// GetSubjectProgress 学科进度。
//
// 原实现：N 个学科 → 循环 N 次 subjectRepo.FindByID（典型 N+1）
// 改造：先 subjectRepo.FindAll 一次拿全量，再 map 拼装，固定 2 次 DB 往返
func (s *StatsService) GetSubjectProgress(userID uint) ([]dto.SubjectProgressResp, error) {
	list, err := s.repo.SubjectProgress(userID)
	if err != nil {
		return nil, err
	}

	subjects, err := s.subjectRepo.FindAll()
	if err != nil {
		return nil, err
	}
	sMap := make(map[uint]string, len(subjects))
	for _, sub := range subjects {
		sMap[sub.ID] = sub.Name
	}

	resp := make([]dto.SubjectProgressResp, 0, len(list))
	for _, r := range list {
		item := dto.SubjectProgressResp{
			SubjectID:    r.SubjectID,
			TotalCount:   r.TotalCount,
			CorrectCount: r.CorrectCount,
			SubjectName:  sMap[r.SubjectID],
		}
		if r.TotalCount > 0 {
			item.Accuracy = float64(r.CorrectCount) / float64(r.TotalCount) * 100
		}
		resp = append(resp, item)
	}
	return resp, nil
}

func (s *StatsService) GetCalendarStats(userID uint, month string) ([]dto.CalendarStatsResp, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}
	startDate := start.Format("2006-01-02")
	endDate := start.AddDate(0, 1, 0).Format("2006-01-02")

	list, err := s.repo.Calendar(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.CalendarStatsResp, 0, len(list))
	for _, r := range list {
		resp = append(resp, dto.CalendarStatsResp{
			Date:           r.Date,
			TotalCount:     r.TotalCount,
			CorrectCount:   r.CorrectCount,
			IncorrectCount: r.TotalCount - r.CorrectCount,
			Duration:       r.Duration,
		})
	}
	return resp, nil
}

func (s *StatsService) GetChapterProgress(userID uint) ([]dto.ChapterProgressResp, error) {
	list, err := s.repo.ChapterProgress(userID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ChapterProgressResp, 0, len(list))
	for _, r := range list {
		item := dto.ChapterProgressResp{
			ChapterID:    r.ChapterID,
			ChapterName:  r.ChapterName,
			SubSubjectID: r.SubSubjectID,
			TotalCount:   r.TotalCount,
			CorrectCount: r.CorrectCount,
		}
		if r.TotalCount > 0 {
			item.Accuracy = float64(r.CorrectCount) / float64(r.TotalCount) * 100
		}
		resp = append(resp, item)
	}
	return resp, nil
}
