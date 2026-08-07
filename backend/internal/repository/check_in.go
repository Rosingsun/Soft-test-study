package repository

import (
	"gorm.io/gorm"

	"github.com/soft-test-study/backend/internal/model"
)

type CheckInRepo struct {
	db *gorm.DB
}

func NewCheckInRepo(db *gorm.DB) *CheckInRepo {
	return &CheckInRepo{db: db}
}

func (r *CheckInRepo) FindByUserAndDate(userID uint, date string) (*model.CheckIn, error) {
	var c model.CheckIn
	err := r.db.Where("user_id = ? AND check_date = ?", userID, date).First(&c).Error
	return &c, err
}

func (r *CheckInRepo) Create(c *model.CheckIn) error {
	return r.db.Create(c).Error
}

// UpdateSummary 更新当日打卡汇总的「最新一次」数据（重打使用），快照字段保持不变
func (r *CheckInRepo) UpdateSummary(id uint, correctCount int, accuracy float64, duration int) error {
	return r.db.Model(&model.CheckIn{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"correct_count": correctCount,
			"accuracy":      accuracy,
			"duration":      duration,
		}).Error
}

// DeleteQuestionsByUserAndDate 删除指定日期的打卡题目快照（重新打卡时重置）
func (r *CheckInRepo) DeleteQuestionsByUserAndDate(userID uint, date string) error {
	return r.db.Where("user_id = ? AND check_date = ?", userID, date).
		Delete(&model.CheckInQuestion{}).Error
}

// ClearAnswersByUserAndDate 清空指定日期打卡题目的作答（保留题目快照，重新作答）
func (r *CheckInRepo) ClearAnswersByUserAndDate(userID uint, date string) error {
	return r.db.Model(&model.CheckInQuestion{}).
		Where("user_id = ? AND check_date = ?", userID, date).
		Updates(map[string]interface{}{
			"answer":     "",
			"is_correct": 0,
			"duration":   0,
		}).Error
}

func (r *CheckInRepo) FindQuestionsByUserAndDate(userID uint, date string) ([]model.CheckInQuestion, error) {
	var list []model.CheckInQuestion
	err := r.db.Where("user_id = ? AND check_date = ?", userID, date).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *CheckInRepo) CreateQuestions(list []model.CheckInQuestion) error {
	if len(list) == 0 {
		return nil
	}
	return r.db.CreateInBatches(list, 20).Error
}

func (r *CheckInRepo) UpdateQuestionAnswer(id uint, answer string, isCorrect, duration int) error {
	return r.db.Model(&model.CheckInQuestion{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"answer":     answer,
			"is_correct": isCorrect,
			"duration":   duration,
		}).Error
}

func (r *CheckInRepo) FindDatesByUser(userID uint) ([]string, error) {
	var dates []string
	err := r.db.Model(&model.CheckIn{}).
		Where("user_id = ? AND status = 1", userID).
		Order("check_date asc").
		Select("DATE_FORMAT(check_date, '%Y-%m-%d')").
		Pluck("DATE_FORMAT(check_date, '%Y-%m-%d')", &dates).Error
	return dates, err
}

func (r *CheckInRepo) FindHistory(userID uint, startDate, endDate string) ([]model.CheckIn, error) {
	var list []model.CheckIn
	err := r.db.Where("user_id = ? AND check_date >= ? AND check_date <= ?",
		userID, startDate, endDate).Order("check_date asc").Find(&list).Error
	return list, err
}

func (r *CheckInRepo) CountByUser(userID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.CheckIn{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *CheckInRepo) AvgAccuracyByUser(userID uint) (float64, error) {
	var v float64
	err := r.db.Model(&model.CheckIn{}).
		Where("user_id = ? AND status = 1", userID).
		Select("COALESCE(AVG(rank_accuracy),0)").Scan(&v).Error
	return v, err
}
