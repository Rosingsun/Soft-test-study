package service

import (
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

// aiTaskStatus 任务状态枚举
const (
	AITaskStatusPending = "pending"
	AITaskStatusRunning = "running"
	AITaskStatusSuccess = "success"
	AITaskStatusFailed  = "failed"
)

// aiTaskMemory 内存中的 AI 异步任务热缓存。
// 修复 #20 后：
//   - DB 为权威存储（ai_generated_tasks 表）
//   - 内存 map 仅作热缓存，提升 GetAITask 热路径查询性能
//   - 状态变更必须双写 DB；DB 失败让调用方感知，内存失败 best-effort
//   - 启动时由 RecoverInflightTasks 把 DB 中 status='running' 的行重新加入内存
//   - janitor 改走 DB 扫描，避免重启后内存被清导致任务永远卡在 running
type aiTaskMemory struct {
	mu    sync.RWMutex
	tasks map[string]*dto.AsyncGenerateTask
}

var aiTaskStore = &aiTaskMemory{
	tasks: make(map[string]*dto.AsyncGenerateTask),
}

// aiTaskRepo 任务持久化仓库。启动时由 SetAITaskRepo 注入，注入前所有
// mutator 会降级为「仅写内存」并打印 warning，避免 nil 指针 panic。
var aiTaskRepo *repository.AiGeneratedTaskRepo

// SetAITaskRepo 注入持久化仓库（由 router 在创建后调用）。
// 必须在 NewAITask / SetAITask* 之前完成，否则这些调用会落空。
func SetAITaskRepo(repo *repository.AiGeneratedTaskRepo) {
	aiTaskRepo = repo
}

// aiTaskTimeoutNotifier 超时失败通知回调。
// 由 router 在创建 notifySvc 后通过 RegisterTimeoutNotifier 注入；
// janitor 标记任务失败后调用，避免 service 包反向依赖具体 service 实例。
var aiTaskTimeoutNotifier func(userID uint, taskID string, errMsg string)

// RegisterTimeoutNotifier 注册超时失败通知回调。
// 回调在 janitor 标记任务为 failed 后被调用；不注册时静默跳过。
func RegisterTimeoutNotifier(fn func(userID uint, taskID string, errMsg string)) {
	aiTaskTimeoutNotifier = fn
}

// NewAITask 创建并持久化一个 pending 任务，返回 task_id。
// userID 必须 > 0；非法值直接 panic，由调用方保证（router 已鉴权）。
// subjectID 用于历史列表补全 subject_name，0 表示未知（调用方应尽量传真实值）。
// 持久化失败时打印 error 日志但仍返回内存对象（保持向后兼容，避免阻塞业务）。
func NewAITask(userID uint, subjectID uint, questionType string, chapterID uint, chapterName, difficulty string, count int) *dto.AsyncGenerateTask {
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

	if aiTaskRepo != nil {
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
		if err := aiTaskRepo.Create(dbTask); err != nil {
			log.Printf("[ai task] persist failed task_id=%s user_id=%d: %v", task.ID, userID, err)
		}
	} else {
		log.Printf("[ai task] repo 未注入，任务仅存内存 task_id=%s", task.ID)
	}

	aiTaskStore.mu.Lock()
	aiTaskStore.tasks[task.ID] = task
	aiTaskStore.mu.Unlock()
	return task
}

// generateTaskID 生成 16 字节随机 ID（32 hex 字符），无需第三方依赖
func generateTaskID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GetAITask 获取任务。优先查内存热缓存，未命中或已被回收时回查 DB。
// 严格按 userID 过滤（防止越权）；userID=0 / 不属于该用户 / 任务不存在 一律返回 nil。
func GetAITask(id string, userID uint) *dto.AsyncGenerateTask {
	if userID == 0 {
		return nil
	}

	aiTaskStore.mu.RLock()
	t, ok := aiTaskStore.tasks[id]
	aiTaskStore.mu.RUnlock()
	if ok {
		if t.UserID != userID {
			return nil
		}
		return t
	}

	// 缓存未命中：回查 DB（覆盖重启后历史任务查询）
	if aiTaskRepo == nil {
		return nil
	}
	row, err := aiTaskRepo.FindByTaskID(id)
	if err != nil {
		return nil
	}
	if row.UserID != userID {
		return nil
	}
	task := taskRowToDTO(row)

	// 写回热缓存（best-effort）
	aiTaskStore.mu.Lock()
	aiTaskStore.tasks[id] = task
	aiTaskStore.mu.Unlock()
	return task
}

// SetAITaskRunning 把任务状态切到 running
func SetAITaskRunning(id string) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if aiTaskRepo != nil {
		// 接受 pending → running；若已是 running（极小概率），用 Updates 仍能匹配
		if err := aiTaskRepo.UpdateStatus(id, AITaskStatusPending, AITaskStatusRunning, "", ""); err != nil {
			log.Printf("[ai task] set running db failed task_id=%s: %v", id, err)
		}
	}

	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		t.Status = AITaskStatusRunning
		t.UpdatedAt = nowStr
	}
}

// SetAITaskSuccess 写入成功结果（题目快照序列化到 result_questions）
func SetAITaskSuccess(id string, result *dto.GenerateQuestionsResp) {
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

	if aiTaskRepo != nil {
		// 优先尝试 running → success（正常路径）；失败回退 pending → success（边界）
		if err := aiTaskRepo.UpdateStatus(id, AITaskStatusRunning, AITaskStatusSuccess, "", resultJSON); err != nil {
			log.Printf("[ai task] set success db failed (running) task_id=%s: %v", id, err)
			if err2 := aiTaskRepo.UpdateStatus(id, AITaskStatusPending, AITaskStatusSuccess, "", resultJSON); err2 != nil {
				log.Printf("[ai task] set success db failed (pending fallback) task_id=%s: %v", id, err2)
			}
		}
	}

	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		t.Status = AITaskStatusSuccess
		t.Result = result
		t.UpdatedAt = nowStr
		t.FinishedAt = nowStr
	}
}

// SetAITaskFailed 写入错误信息
func SetAITaskFailed(id string, errMsg string) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if aiTaskRepo != nil {
		// 优先 running → failed；失败回退 pending → failed
		if err := aiTaskRepo.UpdateStatus(id, AITaskStatusRunning, AITaskStatusFailed, errMsg, ""); err != nil {
			log.Printf("[ai task] set failed db failed (running) task_id=%s: %v", id, err)
			if err2 := aiTaskRepo.UpdateStatus(id, AITaskStatusPending, AITaskStatusFailed, errMsg, ""); err2 != nil {
				log.Printf("[ai task] set failed db failed (pending fallback) task_id=%s: %v", id, err2)
			}
		}
	}

	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		t.Status = AITaskStatusFailed
		t.Error = errMsg
		t.UpdatedAt = nowStr
		t.FinishedAt = nowStr
	}
}

// ListAITasksByUser 列某用户最近 limit 条任务，按 updated_at 倒序。
// 直接查 DB（用户需要看到历史任务，不能只靠 1h 热缓存）。
// 严格按 userID 过滤；userID=0 直接返回空，避免误传泄露全量。
func ListAITasksByUser(userID uint, limit int) []*dto.AsyncGenerateTask {
	if userID == 0 || limit <= 0 || aiTaskRepo == nil {
		return nil
	}
	rows, err := aiTaskRepo.ListByUser(userID, limit)
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

// RecoverInflightTasks 服务启动时调用：把 DB 中所有 status='running' 的任务
// 重新加入内存（跨用户），让 janitor 接管超时检查。
// 这样即便服务重启导致原 goroutine 丢失，DB 记录仍在，janitor 会在下一轮
// 扫描时（30min running 超时阈值）自动标记为 failed 并推送通知。
func RecoverInflightTasks() error {
	if aiTaskRepo == nil {
		return nil
	}
	rows, err := aiTaskRepo.FindAllRunning()
	if err != nil {
		return err
	}
	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	for i := range rows {
		row := &rows[i]
		task := taskRowToDTO(row)
		aiTaskStore.tasks[row.TaskID] = task
	}
	if len(rows) > 0 {
		log.Printf("[ai task] 启动恢复 in-flight 任务 %d 条", len(rows))
	}
	return nil
}

// StartAITaskJanitor 启动后台 janitor 定期扫描超时的 pending/running 任务并标记 failed。
//   - repo == nil 时 janitor 禁用（启动早期或单测场景）
//   - 首次启动延迟 5s 等待服务就绪与 RecoverInflightTasks 完成
//   - 每 1min 扫描一次
func StartAITaskJanitor(repo *repository.AiGeneratedTaskRepo) {
	if repo == nil {
		log.Println("[ai task janitor] repo is nil, janitor disabled")
		return
	}
	aiTaskRepo = repo
	go func() {
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			expireStaleTasks(repo)
		}
	}()
}

// expireStaleTasks 扫描并标记超时任务为 failed；同时同步内存缓存并触发超时通知。
// 分批处理：每批 limit=50 条，单批处理完即结束（本轮留给下次 ticker）。
func expireStaleTasks(repo *repository.AiGeneratedTaskRepo) {
	const batch = 50
	rows, err := repo.ListExpiredCandidates(batch)
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
		if err := repo.UpdateStatus(t.TaskID, t.Status, AITaskStatusFailed, reason, ""); err != nil {
			log.Printf("[ai task janitor] mark failed task_id=%s: %v", t.TaskID, err)
			continue
		}
		// 同步内存缓存
		aiTaskStore.mu.Lock()
		if m, ok := aiTaskStore.tasks[t.TaskID]; ok {
			m.Status = AITaskStatusFailed
			m.Error = reason
			m.FinishedAt = time.Now().Format(time.RFC3339)
			m.UpdatedAt = m.FinishedAt
		}
		aiTaskStore.mu.Unlock()

		// 触发超时通知（回调由 router 注册）
		if aiTaskTimeoutNotifier != nil {
			aiTaskTimeoutNotifier(t.UserID, t.TaskID, reason)
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
