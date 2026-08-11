package repository

import (
	"gorm.io/gorm"
)

// practiceSubjectiveFilter 主观题过滤 SQL 片段
// 主观题（essay / case_study）is_correct 被强制 0，不应计入训练/统计指标
// 集中维护，避免散点 hardcode 与 service.IsSubjectiveType 失同步
const practiceSubjectiveFilter = "q.type NOT IN ('essay', 'case_study')"

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

// Overview 一次往返拿全 6 个聚合指标。
//
// 原实现：4 次独立 SQL（practice/exam/wrong/study_days 各 1 次）+ 4 次 RTT
// 改造：1 条 SQL 用子查询拼成 6 列，单次 RTT 拿全
func (r *StatsRepo) Overview(userID uint) (*StatsOverview, error) {
	var ov StatsOverview
	err := r.db.Raw(`
		SELECT
			COALESCE(p.total_practiced, 0) AS total_practiced,
			COALESCE(p.total_correct,   0) AS total_correct,
			COALESCE(e.total_exams,      0) AS total_exams,
			COALESCE(e.avg_exam_score,   0) AS avg_exam_score,
			COALESCE(w.wrong_count,      0) AS wrong_count,
			COALESCE(s.study_days,       0) AS study_days
		FROM (SELECT 1) x
		LEFT JOIN (
			SELECT
				COUNT(*) AS total_practiced,
				COALESCE(SUM(CASE WHEN pr.is_correct = 1 THEN 1 ELSE 0 END), 0) AS total_correct
			FROM practice_records pr
			JOIN questions q ON q.id = pr.question_id AND ` + practiceSubjectiveFilter + `
			WHERE pr.user_id = ?
		) p ON 1=1
		LEFT JOIN (
			SELECT
				COUNT(*) AS total_exams,
				COALESCE(AVG(er.score), 0) AS avg_exam_score
			FROM exam_records er
			JOIN exam_templates t ON t.id = er.template_id AND (t.question_type IS NULL OR t.question_type = '')
			WHERE er.user_id = ? AND er.status = 'finished'
		) e ON 1=1
		LEFT JOIN (
			SELECT COUNT(*) AS wrong_count
			FROM wrong_questions WHERE user_id = ?
		) w ON 1=1
		LEFT JOIN (
			SELECT COUNT(DISTINCT DATE(pr.created_at)) AS study_days
			FROM practice_records pr
			JOIN questions q ON q.id = pr.question_id AND ` + practiceSubjectiveFilter + `
			WHERE pr.user_id = ?
		) s ON 1=1
	`, userID, userID, userID, userID).Scan(&ov).Error
	if err != nil {
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
		JOIN questions q ON q.id = pr.question_id AND `+practiceSubjectiveFilter+`
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
		WHERE pr.user_id = ? AND `+practiceSubjectiveFilter+`
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
		JOIN questions q ON q.id = pr.question_id AND `+practiceSubjectiveFilter+`
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
		JOIN questions q ON q.id = pr.question_id AND `+practiceSubjectiveFilter+`
		JOIN chapters c ON c.id = q.chapter_id
		WHERE pr.user_id = ? AND q.chapter_id > 0
		GROUP BY c.id, c.name, c.sub_subject_id
		ORDER BY total_count DESC
	`, userID).Scan(&list).Error
	return list, err
}
