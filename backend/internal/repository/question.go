package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type QuestionRepo struct {
	db *gorm.DB
}

func NewQuestionRepo(db *gorm.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) FindByChapterID(chapterID uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("chapter_id = ? AND status = 1", chapterID).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByChapterIDFiltered(chapterID uint, difficulty string) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("chapter_id = ? AND status = 1", chapterID)
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindBySubSubjectID(subSubjectID uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("sub_subject_id = ? AND status = 1", subSubjectID).
		Order("id asc").Find(&list).Error
	return list, err
}

// FindCaseStudies 按子科目查询案例分析题，支持按年份筛选
func (r *QuestionRepo) FindCaseStudies(subSubjectID uint, year int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("sub_subject_id = ? AND type = ? AND status = 1", subSubjectID, model.TypeCaseStudy)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	err := query.Order("year desc, id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByID(id uint) (*model.Question, error) {
	var q model.Question
	err := r.db.Where("id = ? AND status = 1", id).First(&q).Error
	return &q, err
}

func (r *QuestionRepo) FindRandom(subjectID uint, difficulty string, limit int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("subject_id = ? AND status = 1", subjectID).
		Where("type NOT IN ?", []string{model.TypeEssay, model.TypeCaseStudy})
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("RAND()").Limit(limit).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindRandomFiltered(subjectID uint, difficulty, qtype string, limit int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("subject_id = ? AND status = 1", subjectID)
	if qtype == "" {
		query = query.Where("type NOT IN ?", []string{model.TypeEssay, model.TypeCaseStudy})
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if qtype != "" {
		query = query.Where("type = ?", qtype)
	}
	err := query.Order("RAND()").Limit(limit).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) CountBySubjectAndType(subjectID uint, qtype string) (int64, error) {
	var count int64
	query := r.db.Model(&model.Question{}).
		Where("subject_id = ? AND status = 1", subjectID)
	if qtype != "" {
		query = query.Where("type = ?", qtype)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *QuestionRepo) FindSpecial(subjectID uint, qtype, difficulty string, limit int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("subject_id = ? AND status = 1", subjectID)
	if qtype != "" {
		query = query.Where("type = ?", qtype)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("RAND()").Limit(limit).Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindEssayQuestions(subjectID uint, years int) ([]model.Question, error) {
	var list []model.Question
	if years <= 0 {
		years = 5
	}
	thresholdYear := time.Now().Year() - years + 1
	err := r.db.Where("subject_id = ? AND type = ? AND status = 1 AND year >= ?", subjectID, model.TypeEssay, thresholdYear).
		Order("year desc, id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByIDs(ids []uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("id IN ? AND status = 1", ids).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) CountByChapterID(chapterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Question{}).
		Where("chapter_id = ? AND status = 1", chapterID).
		Count(&count).Error
	return count, err
}

func (r *QuestionRepo) BatchCreate(questions []model.Question) error {
	if len(questions) == 0 {
		return nil
	}
	return r.db.CreateInBatches(questions, 50).Error
}
