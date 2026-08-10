package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
)

// aiTaskStatus 任务状态枚举
const (
	AITaskStatusPending = "pending"
	AITaskStatusRunning = "running"
	AITaskStatusSuccess = "success"
	AITaskStatusFailed  = "failed"
)

// aiTaskMemory 内存中的 AI 异步任务管理器
// 设计要点：
//   - 内存存储：任务量大、生命周期短（5min 即可读完），重启丢失可接受
//   - 线程安全：sync.RWMutex 保护 map
//   - 自动清理：超过 1 小时的已完成任务定期回收（避免内存膨胀）
type aiTaskMemory struct {
	mu     sync.RWMutex
	tasks  map[string]*dto.AsyncGenerateTask
	maxAge time.Duration
}

var aiTaskStore = &aiTaskMemory{
	tasks:  make(map[string]*dto.AsyncGenerateTask),
	maxAge: 1 * time.Hour,
}

// NewAITask 创建并存储一个 pending 任务，返回 task_id
func NewAITask(questionType string, chapterID uint, chapterName, difficulty string, count int) *dto.AsyncGenerateTask {
	now := time.Now()
	task := &dto.AsyncGenerateTask{
		ID:           generateTaskID(),
		Status:       AITaskStatusPending,
		QuestionType: questionType,
		ChapterID:    chapterID,
		ChapterName:  chapterName,
		Difficulty:   difficulty,
		Count:        count,
		CreatedAt:    now.Format(time.RFC3339),
		UpdatedAt:    now.Format(time.RFC3339),
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

// GetAITask 获取任务；若不存在或已过期返回 nil
func GetAITask(id string) *dto.AsyncGenerateTask {
	aiTaskStore.mu.RLock()
	t, ok := aiTaskStore.tasks[id]
	aiTaskStore.mu.RUnlock()
	if !ok {
		return nil
	}
	// 超过 1 小时的视为过期
	if isExpired(t, aiTaskStore.maxAge) {
		DeleteAITask(id)
		return nil
	}
	return t
}

// SetAITaskRunning 把任务状态切到 running
func SetAITaskRunning(id string) {
	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		t.Status = AITaskStatusRunning
		t.UpdatedAt = time.Now().Format(time.RFC3339)
	}
}

// SetAITaskSuccess 写入成功结果
func SetAITaskSuccess(id string, result *dto.GenerateQuestionsResp) {
	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		now := time.Now()
		t.Status = AITaskStatusSuccess
		t.Result = result
		t.UpdatedAt = now.Format(time.RFC3339)
		t.FinishedAt = now.Format(time.RFC3339)
	}
}

// SetAITaskFailed 写入错误信息
func SetAITaskFailed(id string, errMsg string) {
	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	if t, ok := aiTaskStore.tasks[id]; ok {
		now := time.Now()
		t.Status = AITaskStatusFailed
		t.Error = errMsg
		t.UpdatedAt = now.Format(time.RFC3339)
		t.FinishedAt = now.Format(time.RFC3339)
	}
}

// DeleteAITask 主动删除任务
func DeleteAITask(id string) {
	aiTaskStore.mu.Lock()
	delete(aiTaskStore.tasks, id)
	aiTaskStore.mu.Unlock()
}

// ListAITasksByUser 列某用户最近 20 条任务（用于前端"任务中心"）
// 当前未做用户维度隔离索引（in-memory 简单实现），可加 if needed
func ListAITasksByUser(userID uint, limit int) []*dto.AsyncGenerateTask {
	aiTaskStore.mu.RLock()
	defer aiTaskStore.mu.RUnlock()
	out := make([]*dto.AsyncGenerateTask, 0, limit)
	for _, t := range aiTaskStore.tasks {
		out = append(out, t)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// CleanupExpiredTasks 清理过期任务；建议由定时器调用
func CleanupExpiredTasks() {
	aiTaskStore.mu.Lock()
	defer aiTaskStore.mu.Unlock()
	now := time.Now()
	for id, t := range aiTaskStore.tasks {
		if isExpired(t, aiTaskStore.maxAge) {
			delete(aiTaskStore.tasks, id)
		}
	}
	_ = now
}

// StartTaskJanitor 启动后台 goroutine 定期清理过期任务（每小时一次）
func StartTaskJanitor() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			CleanupExpiredTasks()
		}
	}()
}

// isExpired 判断任务是否超过 maxAge 未更新
func isExpired(t *dto.AsyncGenerateTask, maxAge time.Duration) bool {
	updated, err := time.Parse(time.RFC3339, t.UpdatedAt)
	if err != nil {
		return false
	}
	return time.Since(updated) > maxAge
}
