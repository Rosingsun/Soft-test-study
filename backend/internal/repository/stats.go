package repository

import (
	"gorm.io/gorm"
)

type StatsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

type StatsOverview struct {
	TotalPracticed int64
	TotalCorrect   int64
	TotalExams     int64
	AvgExamScore   float64
	WrongCount     int64
	StudyDays      int64
}

func (r *StatsRepo) Overview(userID uint) (*StatsOverview, error) {
	var ov StatsOverview

	if err := r.db.Raw(`
		SELECT
			COUNT(*) AS total_practiced,
			COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) AS total_correct
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id AND q.type <> 'essay'
		WHERE pr.user_id = ?
	`, userID).Scan(&ov).Error; err != nil {
		return nil, err
	}

	if err := r.db.Raw(`
		SELECT
			COUNT(*) AS total_exams,
			COALESCE(AVG(er.score), 0) AS avg_exam_score
		FROM exam_records er
		JOIN exam_templates t ON t.id = er.template_id AND (t.question_type IS NULL OR t.question_type = '')
		WHERE er.user_id = ? AND er.status = 'finished'
	`, userID).Scan(&ov).Error; err != nil {
		return nil, err
	}

	if err := r.db.Raw(`
		SELECT COUNT(*) AS wrong_count
		FROM wrong_questions WHERE user_id = ?
	`, userID).Scan(&ov).Error; err != nil {
		return nil, err
	}

	if err := r.db.Raw(`
		SELECT COUNT(DISTINCT DATE(pr.created_at)) AS study_days
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id AND q.type <> 'essay'
		WHERE pr.user_id = ?
	`, userID).Scan(&ov).Error; err != nil {
		return nil, err
	}

	return &ov, nil
}

type DailyStat struct {
	Date         string
	TotalCount   int64
	CorrectCount int64
}

func (r *StatsRepo) Daily(userID uint, startDate string) ([]DailyStat, error) {
	var list []DailyStat
	err := r.db.Raw(`
		SELECT DATE_FORMAT(pr.created_at, '%Y-%m-%d') AS date,
		       COUNT(*) AS total_count,
		       COALESCE(SUM(CASE WHEN pr.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_count
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id AND q.type <> 'essay'
		WHERE pr.user_id = ? AND pr.created_at >= ?
		GROUP BY DATE(pr.created_at)
		ORDER BY date ASC
	`, userID, startDate).Scan(&list).Error
	return list, err
}

type SubjectProgressStat struct {
	SubjectID    uint
	TotalCount   int64
	CorrectCount int64
}

func (r *StatsRepo) SubjectProgress(userID uint) ([]SubjectProgressStat, error) {
	var list []SubjectProgressStat
	err := r.db.Raw(`
		SELECT q.subject_id,
		       COUNT(*) AS total_count,
		       COALESCE(SUM(CASE WHEN pr.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_count
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id
		WHERE pr.user_id = ? AND q.type <> 'essay'
		GROUP BY q.subject_id
		ORDER BY total_count DESC
	`, userID).Scan(&list).Error
	return list, err
}

type CalendarStat struct {
	Date         string
	TotalCount   int64
	CorrectCount int64
	Duration     int64
}

func (r *StatsRepo) Calendar(userID uint, startDate, endDate string) ([]CalendarStat, error) {
	var list []CalendarStat
	err := r.db.Raw(`
		SELECT DATE_FORMAT(pr.created_at, '%Y-%m-%d') AS date,
		       COUNT(*) AS total_count,
		       COALESCE(SUM(CASE WHEN pr.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_count,
		       COALESCE(SUM(pr.duration), 0) AS duration
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id AND q.type <> 'essay'
		WHERE pr.user_id = ? AND pr.created_at >= ? AND pr.created_at < ?
		GROUP BY DATE(pr.created_at)
		ORDER BY date ASC
	`, userID, startDate, endDate).Scan(&list).Error
	return list, err
}

type ChapterProgressStat struct {
	ChapterID    uint
	ChapterName  string
	SubSubjectID uint
	TotalCount   int64
	CorrectCount int64
}

func (r *StatsRepo) ChapterProgress(userID uint) ([]ChapterProgressStat, error) {
	var list []ChapterProgressStat
	err := r.db.Raw(`
		SELECT c.id AS chapter_id,
		       c.name AS chapter_name,
		       c.sub_subject_id,
		       COUNT(*) AS total_count,
		       COALESCE(SUM(CASE WHEN pr.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_count
		FROM practice_records pr
		JOIN questions q ON q.id = pr.question_id AND q.type <> 'essay'
		JOIN chapters c ON c.id = q.chapter_id
		WHERE pr.user_id = ? AND q.chapter_id > 0
		GROUP BY c.id, c.name, c.sub_subject_id
		ORDER BY total_count DESC
	`, userID).Scan(&list).Error
	return list, err
}
