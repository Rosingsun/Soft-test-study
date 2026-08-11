package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WrongQuestionRepo struct {
	db *gorm.DB
}

func NewWrongQuestionRepo(db *gorm.DB) *WrongQuestionRepo {
	return &WrongQuestionRepo{db: db}
}

// Upsert 收录错题：错一次 wrong_count+1 并刷新 last_wrong_at
//
// 原实现：SELECT → if not found INSERT else UPDATE（3 次往返）
// 改造：ON DUPLICATE KEY UPDATE 一次完成（利用 uk_user_question 唯一索引）
func (r *WrongQuestionRepo) Upsert(userID, questionID uint) error {
	now := time.Now()
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"wrong_count":   gorm.Expr("wrong_count + 1"),
			"last_wrong_at": now,
		}),
	}).Create(&model.WrongQuestion{
		UserID:      userID,
		QuestionID:  questionID,
		WrongCount:  1,
		LastWrongAt: now,
	}).Error
}

// UpsertBatch 批量收录错题：一次 SQL 完成多条 upsert。
// 用于模考交卷后大量错题入库。
func (r *WrongQuestionRepo) UpsertBatch(userID uint, questionIDs []uint) error {
	if len(questionIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]model.WrongQuestion, 0, len(questionIDs))
	for _, qid := range questionIDs {
		rows = append(rows, model.WrongQuestion{
			UserID:      userID,
			QuestionID:  qid,
			WrongCount:  1,
			LastWrongAt: now,
		})
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"wrong_count":   gorm.Expr("wrong_count + 1"),
			"last_wrong_at": now,
		}),
	}).CreateInBatches(rows, 100).Error
}

func (r *WrongQuestionRepo) IncrementCorrect(userID, questionID uint) error {
	return r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Update("correct_count", gorm.Expr("correct_count + 1")).Error
}

func (r *WrongQuestionRepo) Delete(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.WrongQuestion{}).Error
}

func (r *WrongQuestionRepo) FindByUser(userID uint) ([]model.WrongQuestion, error) {
	var list []model.WrongQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("last_wrong_at desc").Find(&list).Error
	return list, err
}

// FindByUserAndSubject 错题本按科目过滤。
// 原实现：question_id IN (SELECT id FROM questions WHERE subject_id=?) 嵌套子查询，慢。
// 改造：JOIN questions 一次取齐（命中 idx_user_lastwrong + questions.id 主键）
func (r *WrongQuestionRepo) FindByUserAndSubject(userID, subjectID uint) ([]model.WrongQuestion, error) {
	var list []model.WrongQuestion
	err := r.db.Table("wrong_questions wq").
		Select("wq.*").
		Joins("JOIN questions q ON q.id = wq.question_id").
		Where("wq.user_id = ? AND q.subject_id = ?", userID, subjectID).
		Order("wq.last_wrong_at desc").
		Find(&list).Error
	return list, err
}

// Query 错题本列表：支持按科目/来源筛选与排序。
//   - subjectID 为 nil 时不限制科目
//   - source 支持 all / real（非 AI）/ ai；为空等价于 all
//   - sort 支持 wrong_count（答错次数降序，默认）/ last_wrong_at（最近错误时间降序）
func (r *WrongQuestionRepo) Query(userID uint, subjectID *uint, source, sort string) ([]model.WrongQuestion, error) {
	q := r.db.Table("wrong_questions wq").
		Select("wq.*").
		Joins("JOIN questions q ON q.id = wq.question_id").
		Where("wq.user_id = ?", userID)

	if subjectID != nil {
		q = q.Where("q.subject_id = ?", *subjectID)
	}
	switch source {
	case "ai":
		q = q.Where("q.source = 'ai'")
	case "real":
		q = q.Where("q.source <> 'ai' OR q.source IS NULL OR q.source = ''")
	}

	if sort == "wrong_count" {
		q = q.Order("wq.wrong_count desc, wq.last_wrong_at desc")
	} else {
		q = q.Order("wq.last_wrong_at desc")
	}

	var list []model.WrongQuestion
	err := q.Find(&list).Error
	return list, err
}

func (r *WrongQuestionRepo) FindByUserAndQuestion(userID, questionID uint) (*model.WrongQuestion, error) {
	var wq model.WrongQuestion
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).First(&wq).Error
	return &wq, err
}

func (r *WrongQuestionRepo) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
