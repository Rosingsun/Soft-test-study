package repository

import (
	"time"

	"github.com/soft-test-study/backend/internal/model"
	"gorm.io/gorm"
)

type AiRepo struct {
	db *gorm.DB
}

func NewAiRepo(db *gorm.DB) *AiRepo {
	return &AiRepo{db: db}
}

func (r *AiRepo) CreateQuestions(questions []model.AiGeneratedQuestion) error {
	return r.db.Create(&questions).Error
}

func (r *AiRepo) FindByUserID(userID uint, limit int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *AiRepo) FindRandomByUser(userID uint, count int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ?", userID).
		Order("RAND()").Limit(count).Find(&list).Error
	return list, err
}

// FindRandomByUserAndSubject 按用户+科目随机获取 AI 生成的题目，仅返回已关联 questions 表的记录
func (r *AiRepo) FindRandomByUserAndSubject(userID, subjectID uint, count int) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ? AND subject_id = ? AND question_id > 0", userID, subjectID).
		Order("RAND()").Limit(count).Find(&list).Error
	return list, err
}

// CountByUserAndSubject 统计用户在某科目下已关联题库的 AI 题目数量
func (r *AiRepo) CountByUserAndSubject(userID, subjectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.AiGeneratedQuestion{}).
		Where("user_id = ? AND subject_id = ? AND question_id > 0", userID, subjectID).Count(&count).Error
	return count, err
}

func (r *AiRepo) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.AiGeneratedQuestion{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// AiBatchSummary 一次 AI 生成批次的聚合信息
type AiBatchSummary struct {
	BatchID    string
	SubjectID  uint
	ChapterID  uint
	Type       string
	Difficulty string
	Count      int64
	CreatedAt  time.Time
	// QuestionIDs 该批次题目 ID 列表（逗号分隔，取自 GROUP_CONCAT）
	QuestionIDs string
	// KnowledgePoints 该批次去重知识点（| 分隔，取自 GROUP_CONCAT DISTINCT）
	KnowledgePoints string
}

// ListBatchSummaries 按 batch_id 分组列出某用户最近 limit 次生成批次
func (r *AiRepo) ListBatchSummaries(userID uint, limit int) ([]AiBatchSummary, error) {
	var summaries []AiBatchSummary
	err := r.db.Model(&model.AiGeneratedQuestion{}).
		Select("batch_id, subject_id, chapter_id, type, difficulty, COUNT(*) AS count, MAX(created_at) AS created_at, " +
			"GROUP_CONCAT(question_id) AS question_ids, " +
			"GROUP_CONCAT(DISTINCT knowledge_point SEPARATOR '|') AS knowledge_points").
		Where("user_id = ? AND batch_id != ''", userID).
		Group("batch_id, subject_id, chapter_id, type, difficulty").
		Order("created_at DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

// FindByBatch 查询某用户指定批次下的全部题目（按生成顺序）
func (r *AiRepo) FindByBatch(userID uint, batchID string) ([]model.AiGeneratedQuestion, error) {
	var list []model.AiGeneratedQuestion
	err := r.db.Where("user_id = ? AND batch_id = ?", userID, batchID).
		Order("id ASC").Find(&list).Error
	return list, err
}

// BatchAnswerMap 查询某用户对给定题目（mode='ai'）的最近一次答对情况。
// 返回 question_id → 是否答对；未作答的题目不出现在 map 中。
func (r *AiRepo) BatchAnswerMap(userID uint, questionIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(questionIDs))
	if len(questionIDs) == 0 {
		return result, nil
	}
	var records []model.PracticeRecord
	err := r.db.Where("user_id = ? AND mode = ? AND question_id IN ?",
		userID, "ai", questionIDs).
		Order("id DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[uint]struct{}, len(questionIDs))
	for _, rec := range records {
		if _, ok := seen[rec.QuestionID]; ok {
			continue
		}
		seen[rec.QuestionID] = struct{}{}
		result[rec.QuestionID] = rec.IsCorrect == 1
	}
	return result, nil
}

// AiGeneratedTaskRepo AI 异步出题任务持久化仓库。
// 修复 #20：原 service.aiTaskStore 内存 map 重启即丢。
// DB 为权威存储；service 内存 store 仅作热缓存，详见 service/ai_task.go。
type AiGeneratedTaskRepo struct {
	db *gorm.DB
}

func NewAiGeneratedTaskRepo(db *gorm.DB) *AiGeneratedTaskRepo {
	return &AiGeneratedTaskRepo{db: db}
}

// Create 插入新任务
func (r *AiGeneratedTaskRepo) Create(t *model.AiGeneratedTask) error {
	return r.db.Create(t).Error
}

// FindByTaskID 按 task_id 查（不区分用户，调用方负责 userID 过滤）
func (r *AiGeneratedTaskRepo) FindByTaskID(taskID string) (*model.AiGeneratedTask, error) {
	var row model.AiGeneratedTask
	err := r.db.Where("task_id = ?", taskID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// FindRunningByUser 查某用户当前 status='running' 的任务（启动恢复用，单用户维度）
func (r *AiGeneratedTaskRepo) FindRunningByUser(userID uint) ([]model.AiGeneratedTask, error) {
	var rows []model.AiGeneratedTask
	err := r.db.Where("user_id = ? AND status = ?", userID, "running").
		Find(&rows).Error
	return rows, err
}

// FindAllRunning 查所有 status='running' 的任务（启动恢复用，跨用户）
// 用于 RecoverInflightTasks：服务重启后把全站所有进行中任务加回内存让 janitor 接管。
func (r *AiGeneratedTaskRepo) FindAllRunning() ([]model.AiGeneratedTask, error) {
	var rows []model.AiGeneratedTask
	err := r.db.Where("status = ?", "running").Find(&rows).Error
	return rows, err
}

// ListByUser 列某用户最近 limit 条（按 updated_at 倒序）
func (r *AiGeneratedTaskRepo) ListByUser(userID uint, limit int) ([]model.AiGeneratedTask, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []model.AiGeneratedTask
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// UpdateStatus 原子更新状态（仅当 status 仍在旧值时，避免覆盖完成态）。
//   - errorMsg 非空时写入 error 列
//   - resultJSON 非空时写入 result_questions 列
//   - 目标 status 为 success/failed 时填充 finished_at = NOW()
func (r *AiGeneratedTaskRepo) UpdateStatus(taskID, fromStatus, toStatus string, errorMsg, resultJSON string) error {
	updates := map[string]interface{}{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	if errorMsg != "" {
		updates["error"] = errorMsg
	}
	if resultJSON != "" {
		updates["result_questions"] = resultJSON
	}
	if toStatus == "success" || toStatus == "failed" {
		updates["finished_at"] = time.Now()
	}
	return r.db.Model(&model.AiGeneratedTask{}).
		Where("task_id = ? AND status = ?", taskID, fromStatus).
		Updates(updates).Error
}

// ListExpiredCandidates 全局扫描超时候选：
//   - status='pending' 且 updated_at < NOW() - 5min  → 排队超时
//   - status='running' 且 updated_at < NOW() - 30min → 任务执行超时
// 限定 limit 防止 janitor 单次扫描过载。
func (r *AiGeneratedTaskRepo) ListExpiredCandidates(limit int) ([]model.AiGeneratedTask, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.AiGeneratedTask
	err := r.db.Where(
		"(status = ? AND updated_at < (NOW() - INTERVAL 5 MINUTE)) OR "+
			"(status = ? AND updated_at < (NOW() - INTERVAL 30 MINUTE))",
		"pending", "running",
	).Order("updated_at ASC").Limit(limit).Find(&rows).Error
	return rows, err
}
