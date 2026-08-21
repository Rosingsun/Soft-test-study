package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
)

// AITaskStatus 任务状态枚举
const (
	AITaskStatusPending = "pending"
	AITaskStatusRunning = "running"
	AITaskStatusSuccess = "success"
	AITaskStatusFailed  = "failed"
)

// aiTaskMemory 内存中的 AI 异步任务热缓存。
// DB 为权威存储（ai_generated_tasks 表），内存 map 仅作热缓存提升查询性能。
// 状态变更必须双写 DB；DB 失败让调用方感知，内存失败 best-effort。
type aiTaskMemory struct {
	mu    sync.RWMutex
	tasks map[string]*dto.AsyncGenerateTask
}

// TimeoutNotifier 任务超时失败时的通知回调。
// 由调用方在 manager 上注册，janitor 标记任务失败后调用。
type TimeoutNotifier func(userID uint, taskID string, errMsg string)

// AiTaskManager 集中管理 AI 异步任务（OPT-19 依赖注入重构）：
//   - 持有内存热缓存 + DB 持久化仓库 + 超时通知回调
//   - 替代旧实现的包级 var aiTaskStore / aiTaskRepo / aiTaskTimeoutNotifier
//   - 通过构造函数注入依赖，可单元测试
type AiTaskManager struct {
	store    *aiTaskMemory
	repo     *repository.AiGeneratedTaskRepo
	notifier TimeoutNotifier
}

func NewAiTaskManager(repo *repository.AiGeneratedTaskRepo) *AiTaskManager {
	return &AiTaskManager{
		store: &aiTaskMemory{tasks: make(map[string]*dto.AsyncGenerateTask)},
		repo:  repo,
	}
}

// SetNotifier 注册超时失败通知回调。调用一次即可，未注册时静默跳过。
func (m *AiTaskManager) SetNotifier(fn TimeoutNotifier) {
	m.notifier = fn
}

// NewTask 创建并持久化一个 pending 任务，返回 task_id。
// userID 必须 > 0；非法值直接 panic，由调用方保证（router 已鉴权）。
// subjectID 用于历史列表补全 subject_name，0 表示未知。
// 持久化失败时打印 error 日志但仍返回内存对象（保持向后兼容）。
func (m *AiTaskManager) NewTask(userID uint, subjectID uint, questionType string, chapterID uint, chapterName, difficulty string, count int) *dto.AsyncGenerateTask {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	task := &dto.AsyncGenerateTask{
		ID:           generateTaskID(),
		UserID:       userID,
		SubjectID:    subjectID,
		Status:       AITaskStatusPending,
		QuestionType: questionType,
		ChapterID:    chapterID,
		ChapterName:  chapterName,
		Difficulty:   difficulty,
		Count:        count,
		CreatedAt:    nowStr,
		UpdatedAt:    nowStr,
	}

	if m.repo != nil {
		dbTask := &model.AiGeneratedTask{
			TaskID:       task.ID,
			UserID:       userID,
			SubjectID:    subjectID,
			Status:       AITaskStatusPending,
			QuestionType: questionType,
			ChapterID:    chapterID,
			ChapterName:  chapterName,
			Difficulty:   difficulty,
			Count:        count,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := m.repo.Create(dbTask); err != nil {
			log.Printf("[ai task] persist failed task_id=%s user_id=%d: %v", task.ID, userID, err)
		}
	}

	m.store.mu.Lock()
	m.store.tasks[task.ID] = task
	m.store.mu.Unlock()
	return task
}

// generateTaskID 生成 16 字节随机 ID（32 hex 字符），无需第三方依赖
func generateTaskID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Get 获取任务。优先查内存热缓存，未命中或已被回收时回查 DB。
// 严格按 userID 过滤（防止越权）；userID=0 / 不属于该用户 / 任务不存在 一律返回 nil。
func (m *AiTaskManager) Get(id string, userID uint) *dto.AsyncGenerateTask {
	if userID == 0 {
		return nil
	}

	m.store.mu.RLock()
	t, ok := m.store.tasks[id]
	m.store.mu.RUnlock()
	if ok {
		if t.UserID != userID {
			return nil
		}
		return t
	}

	if m.repo == nil {
		return nil
	}
	row, err := m.repo.FindByTaskID(id)
	if err != nil {
		return nil
	}
	if row.UserID != userID {
		return nil
	}
	task := taskRowToDTO(row)

	m.store.mu.Lock()
	m.store.tasks[id] = task
	m.store.mu.Unlock()
	return task
}

// SetRunning 把任务状态切到 running
func (m *AiTaskManager) SetRunning(id string) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if m.repo != nil {
		if err := m.repo.UpdateStatus(id, AITaskStatusPending, AITaskStatusRunning, "", ""); err != nil {
			log.Printf("[ai task] set running db failed task_id=%s: %v", id, err)
		}
	}

	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	if t, ok := m.store.tasks[id]; ok {
		t.Status = AITaskStatusRunning
		t.UpdatedAt = nowStr
	}
}

// SetSuccess 写入成功结果（题目快照序列化到 result_questions）
func (m *AiTaskManager) SetSuccess(id string, result *dto.GenerateQuestionsResp) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	var resultJSON string
	if result != nil {
		if b, err := json.Marshal(result.Questions); err == nil {
			resultJSON = string(b)
		} else {
			log.Printf("[ai task] marshal result failed task_id=%s: %v", id, err)
		}
	}

	if m.repo != nil {
		if err := m.repo.UpdateStatus(id, AITaskStatusRunning, AITaskStatusSuccess, "", resultJSON); err != nil {
			log.Printf("[ai task] set success db failed (running) task_id=%s: %v", id, err)
			if err2 := m.repo.UpdateStatus(id, AITaskStatusPending, AITaskStatusSuccess, "", resultJSON); err2 != nil {
				log.Printf("[ai task] set success db failed (pending fallback) task_id=%s: %v", id, err2)
			}
		}
	}

	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	if t, ok := m.store.tasks[id]; ok {
		t.Status = AITaskStatusSuccess
		t.Result = result
		t.UpdatedAt = nowStr
		t.FinishedAt = nowStr
	}
}

// SetFailed 写入错误信息
func (m *AiTaskManager) SetFailed(id string, errMsg string) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if m.repo != nil {
		if err := m.repo.UpdateStatus(id, AITaskStatusRunning, AITaskStatusFailed, errMsg, ""); err != nil {
			log.Printf("[ai task] set failed db failed (running) task_id=%s: %v", id, err)
			if err2 := m.repo.UpdateStatus(id, AITaskStatusPending, AITaskStatusFailed, errMsg, ""); err2 != nil {
				log.Printf("[ai task] set failed db failed (pending fallback) task_id=%s: %v", id, err2)
			}
		}
	}

	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	if t, ok := m.store.tasks[id]; ok {
		t.Status = AITaskStatusFailed
		t.Error = errMsg
		t.UpdatedAt = nowStr
		t.FinishedAt = nowStr
	}
}

// ListByUser 列某用户最近 limit 条任务，按 updated_at 倒序。
// 直接查 DB（用户需要看到历史任务，不能只靠 1h 热缓存）。
func (m *AiTaskManager) ListByUser(userID uint, limit int) []*dto.AsyncGenerateTask {
	if userID == 0 || limit <= 0 || m.repo == nil {
		return nil
	}
	rows, err := m.repo.ListByUser(userID, limit)
	if err != nil {
		log.Printf("[ai task] list by user failed user_id=%d: %v", userID, err)
		return nil
	}
	out := make([]*dto.AsyncGenerateTask, 0, len(rows))
	for i := range rows {
		out = append(out, taskRowToDTO(&rows[i]))
	}
	return out
}

// RecoverInflight 服务启动时调用：把 DB 中所有 status='running' 的任务
// 重新加入内存（跨用户），让 janitor 接管超时检查。
// 这样即便服务重启导致原 goroutine 丢失，DB 记录仍在，janitor 会在下一轮
// 扫描时（30min running 超时阈值）自动标记为 failed 并推送通知。
func (m *AiTaskManager) RecoverInflight() error {
	if m.repo == nil {
		return nil
	}
	rows, err := m.repo.FindAllRunning()
	if err != nil {
		return err
	}
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	for i := range rows {
		row := &rows[i]
		task := taskRowToDTO(row)
		m.store.tasks[row.TaskID] = task
	}
	if len(rows) > 0 {
		log.Printf("[ai task] 启动恢复 in-flight 任务 %d 条", len(rows))
	}
	return nil
}

// StartJanitor 启动后台 janitor 定期扫描超时的 pending/running 任务并标记 failed。
//   - 首次启动延迟 5s 等待服务就绪与 RecoverInflight 完成
//   - 每 1min 扫描一次
//   - ctx 取消后 janitor 在下一次 ticker 或立刻退出（OPT-20）
func (m *AiTaskManager) StartJanitor(ctx context.Context) {
	if m.repo == nil {
		log.Println("[ai task janitor] repo is nil, janitor disabled")
		return
	}
	go func() {
		select {
		case <-ctx.Done():
			log.Println("[ai task janitor] 启动前已收到取消信号，janitor 不再启动")
			return
		case <-time.After(5 * time.Second):
		}
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[ai task janitor] 收到取消信号，janitor 退出")
				return
			case <-ticker.C:
				m.expireStaleTasks()
			}
		}
	}()
}

// expireStaleTasks 扫描并标记超时任务为 failed；同时同步内存缓存并触发超时通知。
// 分批处理：每批 limit=50 条，单批处理完即结束（本轮留给下次 ticker）。
func (m *AiTaskManager) expireStaleTasks() {
	const batch = 50
	rows, err := m.repo.ListExpiredCandidates(batch)
	if err != nil {
		log.Printf("[ai task janitor] list expired failed: %v", err)
		return
	}
	for i := range rows {
		t := &rows[i]
		reason := "任务执行超时自动失败"
		if t.Status == AITaskStatusPending {
			reason = "任务排队超时自动失败"
		}
		if err := m.repo.UpdateStatus(t.TaskID, t.Status, AITaskStatusFailed, reason, ""); err != nil {
			log.Printf("[ai task janitor] mark failed task_id=%s: %v", t.TaskID, err)
			continue
		}
		m.store.mu.Lock()
		if mem, ok := m.store.tasks[t.TaskID]; ok {
			mem.Status = AITaskStatusFailed
			mem.Error = reason
			mem.FinishedAt = time.Now().Format(time.RFC3339)
			mem.UpdatedAt = mem.FinishedAt
		}
		m.store.mu.Unlock()

		if m.notifier != nil {
			m.notifier(t.UserID, t.TaskID, reason)
		}
		log.Printf("[ai task janitor] 自动失败 task_id=%s user_id=%d reason=%s",
			t.TaskID, t.UserID, reason)
	}
}

// taskRowToDTO model.AiGeneratedTask → dto.AsyncGenerateTask。
// 处理 result_questions JSON 反序列化与时间格式化。
func taskRowToDTO(row *model.AiGeneratedTask) *dto.AsyncGenerateTask {
	task := &dto.AsyncGenerateTask{
		ID:           row.TaskID,
		UserID:       row.UserID,
		SubjectID:    row.SubjectID,
		Status:       row.Status,
		QuestionType: row.QuestionType,
		ChapterID:    row.ChapterID,
		ChapterName:  row.ChapterName,
		Difficulty:   row.Difficulty,
		Count:        row.Count,
		CreatedAt:    row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    row.UpdatedAt.Format(time.RFC3339),
		Error:        row.Error,
	}
	if row.FinishedAt != nil {
		task.FinishedAt = row.FinishedAt.Format(time.RFC3339)
	}
	if row.ResultQuestions != "" {
		var questions []dto.AiGeneratedQuestionResp
		if err := json.Unmarshal([]byte(row.ResultQuestions), &questions); err == nil {
			task.Result = &dto.GenerateQuestionsResp{Questions: questions}
		}
	}
	return task
}
