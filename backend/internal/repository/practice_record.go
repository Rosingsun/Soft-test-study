package repository

import (
	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type PracticeRecordRepo struct {
	db *gorm.DB
}

func NewPracticeRecordRepo(db *gorm.DB) *PracticeRecordRepo {
	return &PracticeRecordRepo{db: db}
}

func (r *PracticeRecordRepo) Create(record *model.PracticeRecord) error {
	return r.db.Create(record).Error
}

// CreateBatch 批量写入练习记录（重打/AI 收题场景）
func (r *PracticeRecordRepo) CreateBatch(records []model.PracticeRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.CreateInBatches(records, 100).Error
}

func (r *PracticeRecordRepo) FindByUserAndQuestion(userID, questionID uint) (*model.PracticeRecord, error) {
	var record model.PracticeRecord
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Order("created_at desc").First(&record).Error
	return &record, err
}

// FindByUserAndChapter 错题重做：按章节拉取所有练习记录。
// 原实现：question_id IN (SELECT id FROM questions WHERE chapter_id=?) 嵌套子查询
// 改造：JOIN questions 一次取齐
func (r *PracticeRecordRepo) FindByUserAndChapter(userID, chapterID uint) ([]model.PracticeRecord, error) {
	var list []model.PracticeRecord
	err := r.db.Table("practice_records pr").
		Select("pr.*").
		Joins("JOIN questions q ON q.id = pr.question_id").
		Where("pr.user_id = ? AND q.chapter_id = ?", userID, chapterID).
		Find(&list).Error
	return list, err
}

// CountInRange 统计区间内答题数（可过滤科目）。
//
// 修复：subjectID > 0 时会 JOIN questions 表，若 WHERE 子句仍使用裸列名
// （user_id / created_at），MySQL 会报 "Column 'xxx' in where clause is ambiguous"
// 并被 service 层的 `_, _ :=` 静默吞掉，导致学习计划进度永远是 0。
// 这里显式限定为 practice_records.<col>，消除歧义。
func (r *PracticeRecordRepo) CountInRange(userID uint, startDate, endDate string, subjectID uint) (int64, error) {
	var n int64
	q := r.db.Model(&model.PracticeRecord{}).
		Where("practice_records.user_id = ? AND practice_records.created_at >= ? AND practice_records.created_at < DATE_ADD(?, INTERVAL 1 DAY)",
			userID, startDate+" 00:00:00", endDate)
	if subjectID > 0 {
		q = q.Joins("JOIN questions q ON q.id = practice_records.question_id AND q.subject_id = ?", subjectID)
	}
	err := q.Count(&n).Error
	return n, err
}

// CountOnDate 统计指定日期答题数（可过滤科目）
func (r *PracticeRecordRepo) CountOnDate(userID uint, date string, subjectID uint) (int64, error) {
	return r.CountInRange(userID, date, date, subjectID)
}

// CountByMode 统计指定模式的答题数
func (r *PracticeRecordRepo) CountByMode(userID uint, mode string) (int64, error) {
	var n int64
	err := r.db.Model(&model.PracticeRecord{}).
		Where("user_id = ? AND mode = ?", userID, mode).Count(&n).Error
	return n, err
}

// DeleteCheckInRecordsByUserAndDate 删除指定日期内打卡模式（mode='checkin'）的练习记录（重新打卡前重置）
func (r *PracticeRecordRepo) DeleteCheckInRecordsByUserAndDate(userID uint, date string) error {
	return r.db.Where("user_id = ? AND mode = ? AND created_at >= ? AND created_at < DATE_ADD(?, INTERVAL 1 DAY)",
		userID, "checkin", date+" 00:00:00", date).
		Delete(&model.PracticeRecord{}).Error
}
