# 任务 #20：AI 异步任务持久化

> 软考学系平台业务流程修复系列 · 第 8 项
>
> 前 7 项（#4、#5、#1、#13、#14、#11、#12）已修复，本任务承接 #20。

---

## 0. 给新会话 AI 的使用说明

请按以下顺序执行：

1. 读 `D:\ProgramData\Soft-test-study\AGENTS.md`（项目根的 AGENTS.md），遵守全部规范
2. 读本文件全文
3. 按 `第 6 节 · 实施步骤` 顺序编码
4. 按 `第 7 节 · 验收标准` 自检
5. 按 `第 8 节 · 输出格式` 汇报

不要修改本文件描述范围之外的代码（避免 PR 范围蔓延）。

---

## 1. 项目背景

**软考学系平台**——软考刷题与考试系统。

### 1.1 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.x + Gin + GORM + MySQL (utf8mb4) |
| 前端 | Vue 3 + TS + Tailwind + Pinia + Vite |
| 鉴权 | JWT (Bearer) + bcrypt |
| 异步 | 当前：内存 map + goroutine；目标：DB 持久化 |

### 1.2 目录结构

```
backend/
├── cmd/server/main.go         # 入口
├── internal/
│   ├── handler/               # 参数绑定 + 调用 service + 返回响应
│   ├── service/               # 业务编排
│   ├── repository/            # 数据库查询
│   ├── model/                 # GORM 模型
│   ├── dto/                   # 请求/响应结构体
│   ├── middleware/            # JWT、rate-limit
│   ├── router/                # 路由注册
│   ├── config/                # 错误码、常量
│   └── migrations/            # 数据库迁移
├── pkg/                       # jwt、response、llm、validator
└── data/                      # 上传文件

frontend/
├── src/
│   ├── api/                   # axios 封装
│   ├── stores/                # pinia
│   ├── views/                 # 页面
│   ├── components/            # 通用组件
│   ├── types/                 # TS 类型
│   └── router/                # vue-router
```

### 1.3 业务模块

认证、科目/子科目/章节、题目、练习（4 模式）、考试、错题本、复习（SM-2 遗忘曲线）、AI 异步出题/AI 评分/AI 考试、每日打卡、学习计划、排行榜、收藏、标记、学习资料、通知。

---

## 2. 此前已修复的 7 项业务流程问题

| # | 问题 | 关键文件 | 修复要点 |
|---|---|---|---|
| 4 | AI 异步任务无用户隔离 | `service/ai_task.go` `dto/ai.go` `handler/ai.go` | `AsyncGenerateTask` 加 `UserID uint \`json:"-"\``；`GetAITask(id, userID)` / `ListAITasksByUser(userID, limit)` 严格按 userID 过滤；handler 从 JWT 取 userID 透传；越权与"不存在"统一返回 10005 |
| 5 | 收藏文件夹越权写入 | `repository/bookmark.go` `service/bookmark.go` `handler/bookmark.go` | 新增 `FindFolderByID(folderID, userID)`；`AddFavorite` 在 `folderID>0` 时校验归属；不存在/越权统一返回 `ErrForbidden` |
| 1 | 错题练习主观题判分 | `service/wrong_question.go` | `PracticeSubmit` 检测 `IsSubjectiveType` 跳过判分与 OnWrong；`List` 静默过滤主观题 |
| 13 | 排行榜/统计混算主观题 | `repository/ranking.go` `repository/stats.go` | 新增 `practiceSubjectiveFilter` 常量 `q.type NOT IN ('essay', 'case_study')`；替换 6 处硬编码 |
| 14 | AI 评分 record_id 丢失 | `frontend/src/components/practice/PracticeRunner.vue` | 加 `submitting` / `submitError` 状态；`handleSubmit` 不再静默吞错；`handleAiScore` 缺 recordId 时显式提示；模板增加"重新提交"按钮 |
| 11 | StudyPlan 全局 1 个 active | `repository/study_plan.go` `service/study_plan.go` | `FindActiveByUserAndSubject(userID, subjectID)` 替代 `FindActiveByUser`；Create/Update 切换 subject 时校验冲突；新增 `subjectLabelOf` 辅助 |
| 12 | 考试超时结算时机 | `repository/exam_record.go` `service/exam.go` `router/router.go` | 新增 `FindExpiredPending` repo + `settleIfExpired` 辅助 + `settleExpiredRecords` janitor + `StartExamJanitor` 启动入口；`SubmitAnswer` 加超时防御；`StartExam` 重构调用 settleIfExpired |

完整 diff 留在 git 历史；新任务无需重读这些文件，仅作为「既有约束」参考。

---

## 3. 本任务目标（#20）

### 3.1 问题描述

AI 异步出题任务（`POST /api/v1/ai/generate/async`）当前**完全依赖内存存储**：

- 任务状态保存在 `aiTaskStore.tasks`（`service/ai_task.go`），进程内 `map[string]*dto.AsyncGenerateTask`
- 后台 goroutine（`service/ai.go:361` `go s.runGenerateTask(...)`）执行 LLM 调用
- 完成后通过 `notifySvc.Push` 推送通知（`service/ai.go:405, 421`）

### 3.2 4 个具体问题

| # | 问题 | 影响 |
|---|---|---|
| 1 | 服务重启 → 内存 map 清空 → 所有 in-flight 任务消失，用户拿不到结果也没有失败通知 | 数据丢失、用户体验断裂 |
| 2 | `notifySvc.Push` 错误被吞（`_ = s.notifySvc.Push(...)`） | 通知失败无人察觉 |
| 3 | janitor 只清理 >1h 未更新任务，**不区分 running 状态** → 永久 hang 的 LLM 任务不会被超时处理 | 资源泄漏 |
| 4 | 用户无法查询历史任务（内存只保留 1h 内热数据） | 历史任务不可追溯 |

### 3.3 既有约束（来自 #4 修复）

- `dto.AsyncGenerateTask.UserID uint \`json:"-"\`` 已存在，必须沿用
- `GetAITask(id, userID)` 必须按用户过滤（防止越权）
- `ListAITasksByUser(userID, limit)` 同上
- handler 已从 JWT context 取 userID 透传

新表 schema 必须带 `user_id` 列。

---

## 4. 涉及文件清单

### 4.1 需修改

| 文件 | 改动 |
|---|---|
| `backend/internal/service/ai_task.go` | 核心改造：内存 store 降级为热缓存；状态变更双写 DB；启动时从 DB 加载 in-flight 任务 |
| `backend/internal/service/ai.go` | 去掉 `_ = ...Push` 吞错；`runGenerateTask` 状态变更双写；`GetGenerateTask` / `ListGenerateTasks` 改为查 DB（可保留内存为缓存层） |
| `backend/internal/router/router.go` | 注入 `*gorm.DB` 到 janitor 路径；调用新的 `StartAITaskJanitor` |
| `backend/internal/repository/ai.go` | 新增 `AiGeneratedTaskRepo` 与相关方法 |
| `backend/internal/model/ai_generated_task.go` | 新文件：定义 `AiGeneratedTask` model |
| `backend/migrations/000007_ai_generated_tasks.up.sql` | 新文件：建表 SQL |
| `backend/migrations/000007_ai_generated_tasks.down.sql` | 新文件：回滚 SQL |

### 4.2 不需修改

- 前端：API URL 与 DTO 字段不变（`user_id` 已用 `json:"-"` 隐藏）
- 路由层（除 router.go 的 janitor 启动）
- 其他 service

---

## 5. 数据库 Schema 设计

### 5.1 建议表结构

```sql
CREATE TABLE `ai_generated_tasks` (
  `id`              BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `task_id`         VARCHAR(64)     NOT NULL COMMENT '前端可见的 task_id (16 字节 hex)',
  `user_id`         BIGINT UNSIGNED  NOT NULL                COMMENT '归属用户',
  `status`          VARCHAR(20)     NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed',
  `question_type`   VARCHAR(20)     NOT NULL                COMMENT 'single/multi/judge/case_study/essay',
  `chapter_id`      BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `chapter_name`    VARCHAR(200)    NOT NULL DEFAULT '',
  `difficulty`      VARCHAR(20)     NOT NULL DEFAULT '',
  `count`           INT             NOT NULL DEFAULT 0,
  `error`           TEXT            DEFAULT NULL             COMMENT '失败原因',
  `result_questions` JSON           DEFAULT NULL             COMMENT '成功后保存题目 JSON 数组（与 ai_generated_questions 解耦）',
  `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `finished_at`     DATETIME        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_id` (`task_id`),
  KEY `idx_user_status_updated` (`user_id`, `status`, `updated_at`),
  KEY `idx_status_updated` (`status`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='AI 异步出题任务状态';
```

### 5.2 关键决策

- **`task_id` 唯一索引**：16 字节 hex（32 字符）保留与现内存 store 一致
- **复合索引 `(user_id, status, updated_at)`**：支撑 `FindByUserAndStatus` + 用户维度的超时扫描
- **`status, updated_at` 索引**：支撑 janitor 全局扫描 expired pending/running 任务
- **`result_questions` JSON**：保存成功生成的题目快照，避免 janitor 重启时丢失；与 `ai_generated_questions` 表解耦
- **`finished_at` 可空**：仅在 status='success'/'failed' 时填充

### 5.3 迁移文件

参考已有 `000006_ai_generated_batch_id.up.sql` 风格：

```
backend/migrations/
├── 000007_ai_generated_tasks.up.sql
└── 000007_ai_generated_tasks.down.sql
```

---

## 6. 实施步骤

### Step 1: 新增 model

`backend/internal/model/ai_generated_task.go`：

```go
package model

import "time"

type AiGeneratedTask struct {
    ID              uint      `gorm:"primarykey"`
    TaskID          string    `gorm:"column:task_id;size:64;not null;uniqueIndex:uk_task_id"`
    UserID          uint      `gorm:"column:user_id;not null;index:idx_user_status_updated"`
    Status          string    `gorm:"column:status;type:varchar(20);not null;default:'pending'"`
    QuestionType    string    `gorm:"column:question_type;type:varchar(20);not null"`
    ChapterID       uint      `gorm:"column:chapter_id;default:0"`
    ChapterName     string    `gorm:"column:chapter_name;type:varchar(200);default:''"`
    Difficulty      string    `gorm:"column:difficulty;type:varchar(20);default:''"`
    Count           int       `gorm:"column:count;default:0"`
    Error           string    `gorm:"column:error;type:text"`
    ResultQuestions string    `gorm:"column:result_questions;type:json"`
    CreatedAt       time.Time `gorm:"column:created_at"`
    UpdatedAt       time.Time `gorm:"column:updated_at"`
    FinishedAt      *time.Time `gorm:"column:finished_at"`
}

func (AiGeneratedTask) TableName() string {
    return "ai_generated_tasks"
}
```

### Step 2: 新增迁移文件

参考 `000006_ai_generated_batch_id.up.sql` 风格创建 000007 两个文件（up + down）。

### Step 3: 新增 repo

`backend/internal/repository/ai.go` 末尾追加 `AiGeneratedTaskRepo`：

```go
type AiGeneratedTaskRepo struct {
    db *gorm.DB
}

func NewAiGeneratedTaskRepo(db *gorm.DB) *AiGeneratedTaskRepo {
    return &AiGeneratedTaskRepo{db: db}
}

// Create 插入新任务
func (r *AiGeneratedTaskRepo) Create(t *model.AiGeneratedTask) error

// FindByTaskID 按 task_id 查
func (r *AiGeneratedTaskRepo) FindByTaskID(taskID string) (*model.AiGeneratedTask, error)

// FindPendingAndRunningByUser 查某用户进行中的任务（启动时恢复用）
func (r *AiGeneratedTaskRepo) FindPendingAndRunningByUser(userID uint) ([]model.AiGeneratedTask, error)

// ListByUser 列某用户最近 limit 条（按 updated_at 倒序）
func (r *AiGeneratedTaskRepo) ListByUser(userID uint, limit int) ([]model.AiGeneratedTask, error)

// UpdateStatus 原子更新状态（仅当 status 仍在旧值时，避免覆盖完成态）
// finishedAt 非空时填充；error 非空时写入；resultQuestions 非空时写入
func (r *AiGeneratedTaskRepo) UpdateStatus(taskID, fromStatus, toStatus string, errorMsg, resultJSON string) error

// ListExpiredCandidates 全局扫描超时候选：pending > 5min 或 running > 30min
func (r *AiGeneratedTaskRepo) ListExpiredCandidates(limit int) ([]model.AiGeneratedTask, error)
```

### Step 4: 改造 `service/ai_task.go`

**目标**：内存 store 降级为热缓存，DB 为权威。

```go
// 内存 cache 仅用于热点查询，结构保持现状
// 所有 mutator 必须同时写 DB（best-effort，先 DB 后内存，DB 失败时让调用方感知）

func NewAITask(...) *dto.AsyncGenerateTask {
    task := ... // 同 #4 修复
    // 1. 持久化到 DB
    dbTask := &model.AiGeneratedTask{
        TaskID: task.ID, UserID: task.UserID, Status: task.Status,
        QuestionType: task.QuestionType, ...
    }
    if err := aiTaskRepo.Create(dbTask); err != nil {
        log.Printf("[ai task] persist failed task_id=%s: %v", task.ID, err)
    }
    // 2. 写内存（best-effort）
    aiTaskStore.mu.Lock()
    aiTaskStore.tasks[task.ID] = task
    aiTaskStore.mu.Unlock()
    return task
}

func SetAITaskSuccess(id string, result *dto.GenerateQuestionsResp) {
    resultJSON, _ := json.Marshal(result.Questions)  // 序列化题目快照
    if err := aiTaskRepo.UpdateStatus(id,
        AITaskStatusPending, AITaskStatusSuccess, "", string(resultJSON)); err != nil {
        // pending -> success 或 running -> success 都尝试
        _ = aiTaskRepo.UpdateStatus(id,
            AITaskStatusRunning, AITaskStatusSuccess, "", string(resultJSON))
    }
    // 内存 cache 更新
    aiTaskStore.mu.Lock()
    if t, ok := aiTaskStore.tasks[id]; ok {
        t.Status = AITaskStatusSuccess
        t.Result = result
        t.UpdatedAt = time.Now().Format(time.RFC3339)
        t.FinishedAt = time.Now().Format(time.RFC3339)
    }
    aiTaskStore.mu.Unlock()
}
// SetAITaskFailed / SetAITaskRunning 类似改造
```

**保留**：
- `GetAITask(id, userID)` 先查内存，未命中查 DB
- `ListAITasksByUser(userID, limit)` 直接查 DB（最新数据）
- `CleanupExpiredTasks` 改为 `service.expireStaleTasks`，扫描 DB + 更新 DB

### Step 5: 改造 `service/ai.go`

#### 5a. 去掉 `Push` 错误吞没

```go
// 原：
_ = s.notifySvc.Push(...)
// 改为：
if err := s.notifySvc.Push(...); err != nil {
    log.Printf("[ai task] push notification failed user_id=%d task_id=%s: %v", userID, taskID, err)
    // 不返回 error：通知失败不应回滚任务成功状态
}
```

涉及位置：
- `ai.go:405` `pushGenerateSuccessNotification`
- `ai.go:421` `pushGenerateFailedNotification`

#### 5b. 启动时恢复 in-flight 任务

新增方法：

```go
// recoverInflightTasks 服务启动时调用：
// 把 DB 中 status='running' 的任务重新加入内存，
// 让 janitor 接管超时检查（避免重启导致永久 hang）。
func (s *AiService) recoverInflightTasks() error {
    // 不区分用户：所有 running 都加载到内存
    // （使用单独的 recover query，跨用户）
    var rows []model.AiGeneratedTask
    if err := s.taskDB.Where("status = ?", "running").
        Find(&rows).Error; err != nil {
        return err
    }
    for _, row := range rows {
        // 内存中插入占位，状态用 DB 的真实状态
        aiTaskStore.mu.Lock()
        aiTaskStore.tasks[row.TaskID] = &dto.AsyncGenerateTask{
            ID: row.TaskID, UserID: row.UserID, Status: row.Status,
            ...
        }
        aiTaskStore.mu.Unlock()
    }
    return nil
}
```

#### 5c. 重写 `runGenerateTask` 的状态变更

```go
func (s *AiService) runGenerateTask(userID uint, taskID, subjectName string, req dto.GenerateQuestionsReq) {
    defer func() {
        if r := recover(); r != nil {
            errMsg := fmt.Sprintf("AI 出题后台任务异常: %v", r)
            SetAITaskFailed(taskID, errMsg)
            s.pushGenerateFailedNotification(userID, taskID, errMsg)
        }
    }()

    SetAITaskRunning(taskID)

    items, err := s.callAIAndParse(...)
    if err != nil {
        SetAITaskFailed(taskID, err.Error())
        s.pushGenerateFailedNotification(userID, taskID, err.Error())
        return
    }

    resp, err := s.saveGenerated(...)
    if err != nil {
        SetAITaskFailed(taskID, err.Error())
        s.pushGenerateFailedNotification(userID, taskID, "题目入库失败: "+err.Error())
        return
    }

    SetAITaskSuccess(taskID, resp)
    s.pushGenerateSuccessNotification(userID, taskID, subjectName, len(items))
}
```

无需大改，状态变更方法内部已升级。

### Step 6: 改造 janitor

`router.go` 当前：
```go
service.StartTaskJanitor()  // 启动 AI 任务清理
```

改为：
```go
// 注入 db
aiTaskRepo := repository.NewAiGeneratedTaskRepo(db)
aiSvc := service.NewAiService(..., aiTaskRepo, ...)
// 启动时恢复 in-flight
_ = aiSvc.RecoverInflightTasks()
// 启动 janitor（含超时清理）
service.StartAITaskJanitor(aiTaskRepo)
```

`service/ai_task.go` 改造 janitor：

```go
// StartAITaskJanitor 每 1min 扫描一次超时任务并标记 failed
//   - pending > 5min 未启动 → failed
//   - running > 30min 未完成 → failed
func StartAITaskJanitor(repo *repository.AiGeneratedTaskRepo) {
    if repo == nil {
        log.Println("[ai task janitor] repo is nil, janitor disabled")
        return
    }
    go func() {
        time.Sleep(5 * time.Second)
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        for range ticker.C {
            expireStaleTasks(repo)
        }
    }()
}

func expireStaleTasks(repo *repository.AiGeneratedTaskRepo) {
    rows, err := repo.ListExpiredCandidates(50)
    if err != nil {
        log.Printf("[ai task janitor] list expired failed: %v", err)
        return
    }
    for _, t := range rows {
        reason := "任务超时自动失败"
        if t.Status == AITaskStatusPending {
            reason = "任务排队超时自动失败"
        }
        if err := repo.UpdateStatus(t.TaskID, t.Status, AITaskStatusFailed, reason, ""); err != nil {
            log.Printf("[ai task janitor] mark failed task_id=%s: %v", t.TaskID, err)
        }
        // 内存 cache 同步
        aiTaskStore.mu.Lock()
        if m, ok := aiTaskStore.tasks[t.TaskID]; ok {
            m.Status = AITaskStatusFailed
            m.Error = reason
            m.FinishedAt = time.Now().Format(time.RFC3339)
        }
        aiTaskStore.mu.Unlock()
        // 推送失败通知
        // （需持有 notifySvc 与 aiSvc 实例；可在 router 层注入回调）
    }
}
```

**提示**：janitor 需要 `notifySvc` 推送超时通知。建议在 `service` 包加一个**包级 notify 回调**：

```go
// service/ai_task.go
var aiTaskTimeoutNotifier func(userID uint, taskID string, errMsg string)

func RegisterTimeoutNotifier(fn func(userID uint, taskID string, errMsg string)) {
    aiTaskTimeoutNotifier = fn
}
```

router.go 在创建 aiSvc 后调用：
```go
service.RegisterTimeoutNotifier(func(uid uint, tid, msg string) {
    notifySvc.Push(uid, "ai_generate_failed", "AI 出题失败", "任务超时："+msg, "/ai/practice?task_id="+tid)
})
```

### Step 7: 启动恢复调用

router.go `Setup` 函数末尾（所有 service 创建完后）：

```go
// 启动时恢复 in-flight 任务
if err := aiSvc.RecoverInflightTasks(); err != nil {
    log.Printf("[ai task] recover inflight failed: %v", err)
}
```

### Step 8: 向后兼容

- `NewAiService` 构造函数需要新增 `aiTaskRepo` 参数
- `router.go:57` 当前 `NewAiService(aiRepo, questionRepo, subjectRepo, chapterRepo, examRepo, essayScoreRepo, practiceRecordRepo, notifySvc)` 要加 `aiTaskRepo`

---

## 7. 验收标准

执行以下清单逐项打勾，全部通过才算完成。

### 7.1 代码完整性

- [ ] `model/ai_generated_task.go` 新建，含 `TableName()` 与全部 GORM tag
- [ ] `migrations/000007_ai_generated_tasks.up.sql` 含建表 SQL（5.1 节参考）
- [ ] `migrations/000007_ai_generated_tasks.down.sql` 含 `DROP TABLE` 回滚
- [ ] `repository/ai.go` 末尾追加 `AiGeneratedTaskRepo` + 6 个方法
- [ ] `service/ai_task.go` 改造：所有 mutator 双写 DB；janitor 走 repo
- [ ] `service/ai.go` 改造：去掉 `_ = ...Push`；新增 `RecoverInflightTasks` 方法
- [ ] `router/router.go` 注入 `aiTaskRepo` 到 aiSvc；启动时调用 `RecoverInflightTasks` 与 `StartAITaskJanitor`

### 7.2 编译 / 类型检查

- [ ] `cd backend && go build ./...` 零报错
- [ ] `cd backend && go vet ./...` 零告警
- [ ] 前端无修改，`cd frontend && npx vue-tsc --noEmit` 仍然通过

### 7.3 行为验证（伪代码测试 / 手动测试）

- [ ] **重启不丢任务**：DB 中写入一条 status='running' 的任务 → 重启服务 → janitor 第 1 轮扫描后该任务被标记 failed 并推送通知
- [ ] **超时清理**：DB 中写入 status='running' + updated_at < NOW() - 30min → janitor 标记 failed
- [ ] **排队超时**：DB 中写入 status='pending' + updated_at < NOW() - 5min → janitor 标记 failed
- [ ] **通知失败不静默**：把 `notifications` 表改为只读 / DROP，触发 push → 服务日志有 error（不再 `_ =` 吞）
- [ ] **user 过滤仍然生效**：用户 A 创建任务，用户 B 调 `GET /ai/tasks/{task_id}` → 返回 10005（与 #4 一致）
- [ ] **用户列表**：用户 A 调 `GET /ai/tasks` → 只看到自己创建的任务（不包含其他用户的）
- [ ] **正常完成路径**：发起一次真实 AI 出题 → DB 出现 status='success' + result_questions JSON + finished_at；用户收到通知
- [ ] **失败路径**：构造一个会失败的 req（错误的 API Key）→ DB 出现 status='failed' + error + finished_at

### 7.4 错误码一致性

- [ ] 新增的所有错误使用 `config.CodeXxx` 常量（不要硬编码 10001 等数字）
- [ ] 错误消息中文，与现有风格一致

---

## 8. 输出格式

完成任务后请按以下格式汇报：

```markdown
## 完成总结

### 改动文件清单
- `path/to/file.go`：[改了什么，1-2 行]
- ...

### 新增文件清单
- `path/to/new_file.go`：[用途]
- ...

### 编译 / 类型检查
- `go build ./...`：✅ / ❌（如有错误请贴出）
- `go vet ./...`：✅ / ❌
- `npx vue-tsc --noEmit`：N/A（前端无改动）

### 关键决策
- [选择 1：为什么选 A 不选 B]
- [选择 2：...]
- ...

### 遗留 TODO
- [问题 1：影响 + 建议方案]
- ...

### 验收清单自检
- [x] 代码完整性
- [x] 编译 / 类型检查
- [x] 行为验证（[简要描述如何验证]）
- [x] 错误码一致性
```

---

## 9. 不要做的事

- ❌ 不要修改 #4 已修代码的 user_id 过滤逻辑
- ❌ 不要修改前端代码（任务范围限定为后端）
- ❌ 不要改 `dto.AsyncGenerateTask` 的 JSON 字段（前端会受影响）
- ❌ 不要引入 Redis 等额外依赖（保持 MySQL-only）
- ❌ 不要把内存 store 完全去掉（保留为热缓存提升性能）
- ❌ 不要在 janitor 里跑长事务（分批 + 短事务）
- ❌ 不要修改其他 service 的业务逻辑
- ❌ 不要修改 issue 清单中其他未解决问题（聚焦 #20）

---

## 10. 参考资源

- `AGENTS.md`（项目根）—— 必须遵守的规范
- `backend/migrations/000001_init_schema.up.sql` —— 现有表风格
- `backend/migrations/000006_ai_generated_batch_id.up.sql` —— 与本任务最相关的迁移参考
- `backend/internal/service/ai_task.go` —— 当前实现
- `backend/internal/service/ai.go:341-395` —— `runGenerateTask` 流程
- `backend/internal/service/exam.go`（最近一次 janitor 参考）—— `StartExamJanitor` 的启动模式
- `backend/internal/repository/exam_record.go:78-104`（`FindExpiredPending`）—— 超时扫描 SQL 风格参考

---

> 文档版本：v1.0 · 创建于 2026-08-11
> 上游会话：「软考学系平台业务流程问题总结 + 逐项修复 #4/#5/#1/#13/#14/#11/#12」
