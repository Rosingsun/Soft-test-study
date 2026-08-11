package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type RankingRepo struct {
	db *gorm.DB
}

func NewRankingRepo(db *gorm.DB) *RankingRepo {
	return &RankingRepo{db: db}
}

type RankItem struct {
	UserID uint
	Value  float64
}

type PracticeRankRow struct {
	UserID      uint
	Correct     int64
	Total       int64
}

type ExamScoreRow struct {
	UserID     uint
	Score      int
	FinishedAt time.Time
}

type CheckInDateRow struct {
	UserID    uint
	CheckDate string
}

// TotalUsersByLevel 统计范围内用户数（status=1）
func (r *RankingRepo) TotalUsersByLevel(levelID uint) (int64, error) {
	var n int64
	q := r.db.Model(&struct {
		ID uint `gorm:"primarykey"`
	}{}).Table("users").Where("status = 1")
	if levelID > 0 {
		q = q.Where("level_id = ?", levelID)
	}
	err := q.Count(&n).Error
	return n, err
}

// CheckinAccuracy 各用户打卡平均正确率（采用首次打卡快照 rank_accuracy，重打不计入排名）
func (r *RankingRepo) CheckinAccuracy(levelID uint) ([]RankItem, error) {
	var list []RankItem
	q := r.db.Table("check_ins ci").
		Select("ci.user_id AS user_id, COALESCE(AVG(ci.rank_accuracy),0) AS value").
		Where("ci.status = 1")
	if levelID > 0 {
		q = q.Joins("JOIN users u ON u.id = ci.user_id AND u.level_id = ?", levelID)
	}
	err := q.Group("ci.user_id").Scan(&list).Error
	return list, err
}

// CheckinDates 各用户打卡日期（用于计算连续天数）
func (r *RankingRepo) CheckinDates(levelID uint) ([]CheckInDateRow, error) {
	var list []CheckInDateRow
	q := r.db.Table("check_ins ci").
		Select("ci.user_id AS user_id, DATE_FORMAT(ci.check_date, '%Y-%m-%d') AS check_date").
		Where("ci.status = 1")
	if levelID > 0 {
		q = q.Joins("JOIN users u ON u.id = ci.user_id AND u.level_id = ?", levelID)
	}
	err := q.Order("ci.user_id asc, ci.check_date asc").Scan(&list).Error
	return list, err
}

// ExamAccuracy 各用户模考平均正确率
func (r *RankingRepo) ExamAccuracy(levelID, subjectID uint) ([]RankItem, error) {
	var list []RankItem
	q := r.db.Table("exam_records er").
		Select("er.user_id AS user_id, COALESCE(AVG(er.accuracy),0) AS value").
		Where("er.status = ?", "finished")
	if levelID > 0 {
		q = q.Joins("JOIN users u ON u.id = er.user_id AND u.level_id = ?", levelID)
	}
	if subjectID > 0 {
		q = q.Where("er.template_id IN (SELECT id FROM exam_templates WHERE subject_id = ?)", subjectID)
	}
	err := q.Group("er.user_id").Scan(&list).Error
	return list, err
}

// ExamScores 各用户已交卷记录（分数 + 交卷时间），用于时间加权预估分
func (r *RankingRepo) ExamScores(levelID, subjectID uint) ([]ExamScoreRow, error) {
	var list []ExamScoreRow
	q := r.db.Table("exam_records er").
		Select("er.user_id AS user_id, er.score AS score, er.finished_at AS finished_at").
		Where("er.status = ?", "finished")
	if levelID > 0 {
		q = q.Joins("JOIN users u ON u.id = er.user_id AND u.level_id = ?", levelID)
	}
	if subjectID > 0 {
		q = q.Where("er.template_id IN (SELECT id FROM exam_templates WHERE subject_id = ?)", subjectID)
	}
	err := q.Order("er.user_id asc, er.finished_at asc").Scan(&list).Error
	return list, err
}

// PracticeAccuracy 各用户训练正确率（仅统计练习类模式，排除打卡/复习/主观题）
// 排除主观题（essay/case_study）：主观题 is_correct 被强制 0，会拉低正确率且不具可比性
func (r *RankingRepo) PracticeAccuracy(levelID, subjectID uint) ([]PracticeRankRow, error) {
	var list []PracticeRankRow
	q := r.db.Table("practice_records pr").
		Select("pr.user_id AS user_id, COALESCE(SUM(pr.is_correct),0) AS correct, COUNT(*) AS total").
		Joins("JOIN questions q ON q.id = pr.question_id AND q.type NOT IN ?",
			[]string{model.TypeEssay, model.TypeCaseStudy}).
		Where("pr.mode IN ?", []string{"chapter", "random", "special", "wrong", "ai"})
	if levelID > 0 {
		q = q.Joins("JOIN users u ON u.id = pr.user_id AND u.level_id = ?", levelID)
	}
	if subjectID > 0 {
		q = q.Where("q.subject_id = ?", subjectID)
	}
	err := q.Group("pr.user_id").Scan(&list).Error
	return list, err
}
