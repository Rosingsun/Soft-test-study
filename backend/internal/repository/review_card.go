package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReviewCardRepo struct {
	db *gorm.DB
}

func NewReviewCardRepo(db *gorm.DB) *ReviewCardRepo {
	return &ReviewCardRepo{db: db}
}

func (r *ReviewCardRepo) FindByUserAndQuestion(userID, questionID uint) (*model.ReviewCard, error) {
	var card model.ReviewCard
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&card).Error
	return &card, err
}

func (r *ReviewCardRepo) Create(card *model.ReviewCard) error {
	return r.db.Create(card).Error
}

func (r *ReviewCardRepo) Save(card *model.ReviewCard) error {
	return r.db.Save(card).Error
}

// UpsertBatch 批量新增或重置复习卡片（SM-2 初始值）。
// 用于模考交卷后批量收录错题复习，1 次 SQL 完成 N 条 upsert。
func (r *ReviewCardRepo) UpsertBatch(userID uint, questionIDs []uint) error {
	if len(questionIDs) == 0 {
		return nil
	}
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	cards := make([]model.ReviewCard, 0, len(questionIDs))
	for _, qid := range questionIDs {
		cards = append(cards, model.ReviewCard{
			UserID:       userID,
			QuestionID:   qid,
			Repetition:   0,
			IntervalDays: 1,
			EaseFactor:   2.5,
			DueDate:      today,
			Status:       1,
		})
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"repetition":    0,
			"interval_days": 1,
			"due_date":      today,
			"status":        1,
		}),
	}).CreateInBatches(cards, 100).Error
}

func (r *ReviewCardRepo) FindDueByUser(userID uint, date string, limit int) ([]model.ReviewCard, error) {
	var list []model.ReviewCard
	err := r.db.Where("user_id = ? AND status = 1 AND due_date <= ?", userID, date).
		Order("due_date asc, id asc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *ReviewCardRepo) CountDue(userID uint, date string) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = 1 AND due_date <= ?", userID, date).Count(&n).Error
	return n, err
}

func (r *ReviewCardRepo) CountDueBetween(userID uint, startDate, endDate string) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = 1 AND due_date >= ? AND due_date <= ?", userID, startDate, endDate).
		Count(&n).Error
	return n, err
}

func (r *ReviewCardRepo) CountByStatus(userID uint, status int) (int64, error) {
	var n int64
	err := r.db.Model(&model.ReviewCard{}).
		Where("user_id = ? AND status = ?", userID, status).Count(&n).Error
	return n, err
}

// ReviewCardStats 复习概览聚合：一次 SELECT 拿全 4 项指标。
//   - due_today    = COUNT(status=1 AND due_date <= today)
//   - due_tomorrow = COUNT(status=1 AND due_date == tomorrow)
//   - mastered     = COUNT(status=0)
//   - reviewing    = COUNT(status=1)
type ReviewCardStats struct {
	DueToday    int64
	DueTomorrow int64
	Mastered    int64
	Reviewing   int64
}

// Stats 一次往返拿全概览数据。
//
// 原实现：4 次独立 Count（每次 1 次 RTT）
// 改造：1 条 SQL 用 SUM(CASE WHEN ...) 聚合
func (r *ReviewCardRepo) Stats(userID uint, today, tomorrow string) (*ReviewCardStats, error) {
	var s ReviewCardStats
	err := r.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN status = 1 AND due_date <= ? THEN 1 ELSE 0 END), 0) AS due_today,
			COALESCE(SUM(CASE WHEN status = 1 AND due_date =  ? THEN 1 ELSE 0 END), 0) AS due_tomorrow,
			COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END),                  0) AS mastered,
			COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END),                  0) AS reviewing
		FROM review_cards
		WHERE user_id = ?
	`, today, tomorrow, userID).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
