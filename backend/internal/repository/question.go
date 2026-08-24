package repository

import (
	"fmt"
	"math/rand"
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

// FindCaseStudies 按子科目查询案例分析题大题目（parent_id=0），支持按年份筛选
func (r *QuestionRepo) FindCaseStudies(subSubjectID uint, year int) ([]model.Question, error) {
	var list []model.Question
	query := r.db.Where("sub_subject_id = ? AND type = ? AND parent_id = 0 AND status = 1", subSubjectID, model.TypeCaseStudy)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	err := query.Order("year desc, id asc").Find(&list).Error
	return list, err
}

// FindByParentID 查询指定大题目下的所有小题目（按 id 升序）
func (r *QuestionRepo) FindByParentID(parentID uint) ([]model.Question, error) {
	var list []model.Question
	err := r.db.Where("parent_id = ? AND status = 1", parentID).
		Order("id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByID(id uint) (*model.Question, error) {
	var q model.Question
	err := r.db.Where("id = ? AND status = 1", id).First(&q).Error
	return &q, err
}

// FindRandom 按条件随机抽 N 道题。
//
// 实现：用"主键洗牌"替代 ORDER BY RAND()：
//   1. 仅 Pluck 命中条件的主键（只走索引，单次查询极快）
//   2. 应用层 Fisher-Yates 洗牌后取前 N 个
//   3. 用 IN 一次回表取完整行
//
// 相对原方案（每行生成随机数 + 全结果集 filesort）复杂度从 O(N log N) 降到 O(N)。
func (r *QuestionRepo) FindRandom(subjectID uint, difficulty string, limit int) ([]model.Question, error) {
	ids, err := r.randomIDs(r.db.Model(&model.Question{}).
		Where("subject_id = ? AND status = 1", subjectID).
		Where("type NOT IN ?", []string{model.TypeEssay, model.TypeCaseStudy}),
		"difficulty", difficulty, limit)
	if err != nil {
		return nil, err
	}
	return r.findByIDsOrdered(ids)
}

// FindRandomFiltered 按条件随机抽 N 道题
// excludeSource 非空时排除 source = excludeSource 的题（如排除 AI 生成的题，仅抽真题）
// 当 qtype = 'case_study' 时，同时返回父题及其子题（parent_id 关联）
func (r *QuestionRepo) FindRandomFiltered(subjectID uint, difficulty, qtype, excludeSource string, limit int) ([]model.Question, error) {
	q := r.db.Model(&model.Question{}).
		Where("subject_id = ? AND status = 1", subjectID)
	if qtype == "" {
		q = q.Where("type NOT IN ?", []string{model.TypeEssay, model.TypeCaseStudy})
	} else if qtype == model.TypeCaseStudy {
		// 案例分析：选择 parent_id=0 的大题目，同时带回子题
		ids, err := r.randomIDs(q.Where("type = ? AND parent_id = 0", model.TypeCaseStudy), "difficulty", difficulty, limit)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, nil
		}
		// 查询父题 + 所有子题
		var list []model.Question
		err = r.db.Where("(id IN ? AND status = 1) OR parent_id IN ?", ids, ids).
			Order("parent_id asc, id asc").Find(&list).Error
		return list, err
	} else {
		q = q.Where("type = ?", qtype)
	}
	if excludeSource != "" {
		q = q.Where("source <> ?", excludeSource)
	}
	ids, err := r.randomIDs(q, "difficulty", difficulty, limit)
	if err != nil {
		return nil, err
	}
	return r.findByIDsOrdered(ids)
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

// CountBySubjectTypes 批量按 (subject_id, type) 统计题量，返回 map。
// 用于 exam.ListTemplates 消除 N+1。
// excludeSource 非空时排除 source = excludeSource 的题（如排除 AI 生成的题，仅统计真题题量）
func (r *QuestionRepo) CountBySubjectTypes(pairs []SubjectTypePair, excludeSource string) (map[SubjectTypePair]int64, error) {
	out := make(map[SubjectTypePair]int64, len(pairs))
	if len(pairs) == 0 {
		return out, nil
	}
	// 一次 IN 查询，过滤所有命中 (subject_id, type) 的题
	subjectIDs := make([]uint, 0, len(pairs))
	typeSeen := make(map[string]bool)
	for _, p := range pairs {
		if p.SubjectID == 0 {
			continue
		}
		subjectIDs = append(subjectIDs, p.SubjectID)
		if p.Type != "" && !typeSeen[p.Type] {
			typeSeen[p.Type] = true
		}
	}
	if len(subjectIDs) == 0 {
		return out, nil
	}
	type rows struct {
		SubjectID uint
		Type      string
		Count     int64
	}
	var rs []rows
	q := r.db.Model(&model.Question{}).
		Select("subject_id, type, COUNT(*) AS count").
		Where("status = 1 AND subject_id IN ?", subjectIDs)
	if excludeSource != "" {
		q = q.Where("source <> ?", excludeSource)
	}
	if err := q.Group("subject_id, type").Scan(&rs).Error; err != nil {
		return nil, err
	}
	for _, row := range rs {
		out[SubjectTypePair{SubjectID: row.SubjectID, Type: row.Type}] = row.Count
	}
	return out, nil
}

// SubjectTypePair 复合 key
type SubjectTypePair struct {
	SubjectID uint
	Type      string
}

// FindSpecial 专项练习：按题型/难度随机抽题（数量按大题计）
// 题源 = 科目下 status=1 且 type 匹配的全部题目（含真题 + AI 生成题）
// 统一只抽 parent_id=0 的题目，小题永不单独计入数量；
// 案例分析等大题命中后自动带回其全部子题（parent_id 关联）
func (r *QuestionRepo) FindSpecial(subjectID uint, qtype, difficulty string, limit int) ([]model.Question, error) {
	q := r.db.Model(&model.Question{}).
		Where("subject_id = ? AND status = 1 AND parent_id = 0", subjectID)
	if qtype != "" {
		q = q.Where("type = ?", qtype)
	}
	ids, err := r.randomIDs(q, "difficulty", difficulty, limit)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 回表父题 + 子题：子题 parent_id 指向抽中的大题，按父题分组、组内按 id 升序
	var list []model.Question
	err = r.db.Where("(id IN ? AND status = 1) OR (parent_id IN ? AND status = 1)", ids, ids).
		Order("parent_id asc, id asc").Find(&list).Error
	return list, err
}

// FindEssayQuestions 论文题：按年份范围返回全部（不随机）
func (r *QuestionRepo) FindEssayQuestions(subjectID uint, years int) ([]model.Question, error) {
	if years <= 0 {
		years = 5
	}
	thresholdYear := time.Now().Year() - years + 1
	var list []model.Question
	err := r.db.Where("subject_id = ? AND type = ? AND status = 1 AND year >= ?", subjectID, model.TypeEssay, thresholdYear).
		Order("year desc, id asc").Find(&list).Error
	return list, err
}

func (r *QuestionRepo) FindByIDs(ids []uint) ([]model.Question, error) {
	var list []model.Question
	if len(ids) == 0 {
		return list, nil
	}
	err := r.db.Where("id IN ? AND status = 1", ids).
		Order("id asc").Find(&list).Error
	return list, err
}

// FindRandomReal 随机取真题客观题（year>0，单选/多选/判断）
func (r *QuestionRepo) FindRandomReal(limit int) ([]model.Question, error) {
	ids, err := r.randomIDs(r.db.Model(&model.Question{}).
		Where("year > 0 AND status = 1").
		Where("type IN ?", []string{model.TypeSingle, model.TypeMulti, model.TypeJudge}),
		"", "", limit)
	if err != nil {
		return nil, err
	}
	return r.findByIDsOrdered(ids)
}

// FindRandomObjective 随机取客观题（不限年份），用于真题不足时兜底
func (r *QuestionRepo) FindRandomObjective(limit int) ([]model.Question, error) {
	ids, err := r.randomIDs(r.db.Model(&model.Question{}).
		Where("status = 1").
		Where("type IN ?", []string{model.TypeSingle, model.TypeMulti, model.TypeJudge}),
		"", "", limit)
	if err != nil {
		return nil, err
	}
	return r.findByIDsOrdered(ids)
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

// randomIDs 主键洗牌：在给定 query（已拼好 WHERE 条件）上随机抽取 N 个主键。
// filterField / filterValue 可选：用于加一个 WHERE 条件。
func (r *QuestionRepo) randomIDs(q *gorm.DB, filterField, filterValue string, limit int) ([]uint, error) {
	if limit <= 0 {
		return nil, nil
	}
	if filterField != "" && filterValue != "" {
		q = q.Where(fmt.Sprintf("%s = ?", filterField), filterValue)
	}
	var ids []uint
	if err := q.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids, nil
}

// findByIDsOrdered 保持入参顺序回表查询（Pluck 后顺序已被打乱，按 id 升序重新排列）
func (r *QuestionRepo) findByIDsOrdered(ids []uint) ([]model.Question, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.Question
	err := r.db.Where("id IN ? AND status = 1", ids).
		Order("id asc").Find(&list).Error
	return list, err
}
