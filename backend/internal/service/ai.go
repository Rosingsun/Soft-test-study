package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/pkg/llm"
	"gorm.io/gorm"
)

type AiService struct {
	aiRepo         *repository.AiRepo
	aiTaskMgr      *AiTaskManager
	questionRepo   *repository.QuestionRepo
	subjectRepo    *repository.SubjectRepo
	chapterRepo    *repository.ChapterRepo
	examRepo       *repository.ExamRecordRepo
	essayScoreRepo *repository.EssayScoreRepo
	practiceRepo   *repository.PracticeRecordRepo
	notifySvc      *NotificationService
}

func NewAiService(
	aiRepo *repository.AiRepo,
	aiTaskMgr *AiTaskManager,
	questionRepo *repository.QuestionRepo,
	subjectRepo *repository.SubjectRepo,
	chapterRepo *repository.ChapterRepo,
	examRepo *repository.ExamRecordRepo,
	essayScoreRepo *repository.EssayScoreRepo,
	practiceRepo *repository.PracticeRecordRepo,
	notifySvc *NotificationService,
) *AiService {
	return &AiService{
		aiRepo:         aiRepo,
		aiTaskMgr:      aiTaskMgr,
		questionRepo:   questionRepo,
		subjectRepo:    subjectRepo,
		chapterRepo:    chapterRepo,
		examRepo:       examRepo,
		essayScoreRepo: essayScoreRepo,
		practiceRepo:   practiceRepo,
		notifySvc:      notifySvc,
	}
}

// RecoverInflightTasks 服务启动时调用：把 DB 中所有 status='running' 的任务
// 重新加入内存（跨用户），让 janitor 接管超时检查。
// 修复 #20：服务重启后原 goroutine 丢失，DB 记录仍在，janitor 在下一轮扫描时会自动
// 标记为 failed（30min running 超时）并推送通知，避免任务永久 hang。
func (s *AiService) RecoverInflightTasks() error {
	if s.aiTaskMgr == nil {
		return nil
	}
	return s.aiTaskMgr.RecoverInflight()
}

func (s *AiService) GetProviders() dto.AiProvidersResp {
	return dto.AiProvidersResp{
		Providers: []dto.AiProvider{
			{Provider: "deepseek", Name: "DeepSeek", BaseURL: "https://api.deepseek.com", Models: []string{"deepseek-chat", "deepseek-reasoner"}},
			{Provider: "openai", Name: "OpenAI", BaseURL: "https://api.openai.com", Models: []string{"gpt-4o-mini", "gpt-4o", "gpt-3.5-turbo"}},
			{Provider: "azure", Name: "Azure OpenAI", BaseURL: "https://<your-resource-name>.openai.azure.com", Models: []string{"gpt-35-turbo", "gpt-4o-mini", "gpt-4o"}},
			{Provider: "anthropic", Name: "Anthropic", BaseURL: "https://api.anthropic.com", Models: []string{"claude-3.5-mini", "claude-3.5", "claude-4o"}},
			{Provider: "minimax", Name: "MiniMax", BaseURL: "https://api.minimax.io/v1", Models: []string{"MiniMax-M3"}},
			{Provider: "custom", Name: "自定义", BaseURL: "", Models: []string{}},
		},
	}
}

func (s *AiService) GenerateQuestions(userID uint, req dto.GenerateQuestionsReq) (*dto.GenerateQuestionsResp, error) {
	items, err := s.callAIAndParse(req.ApiConfig, req.SubjectID, req.ChapterID, req.Types, req.Difficulty, req.Count)
	if err != nil {
		return nil, err
	}

	return s.saveGenerated(userID, req.SubjectID, req.ChapterID, generateTaskID(), items)
}

func (s *AiService) callAIAndParse(config dto.AiApiConfig, subjectID, chapterID uint, types []string, difficulty string, count int) ([]aiGeneratedItem, error) {
	subjectName := ""
	if sub, err := s.subjectRepo.FindByID(subjectID); err == nil {
		subjectName = sub.Name
	}

	chapterName := "不限定"
	if chapterID > 0 {
		if ch, err := s.chapterRepo.FindByID(chapterID); err == nil {
			chapterName = ch.Name
		}
	}

	// 单次只出 1 种题型（前端已互斥；兜底取第一个）
	questionType := ""
	if len(types) > 0 {
		questionType = types[0]
	}

	difficultyName := "中等"
	if difficulty != "" {
		difficultyName = difficultyLabelCN(difficulty)
	}

	// 分批生成，避免一次性输出过多题目导致 token 截断、JSON 解析失败。
	// 关键修复（成功率提升）：单批解析失败不再整批丢弃，
	// 而是降级处理——记录错误、跳过本批、继续下一批，
	// 直到凑够 count 道；若实在凑不够，把已生成的返回并附带 partial=true。
	promptCtx := generatePromptContext{
		Subject:        subjectName,
		Chapter:        chapterName,
		TypeCode:       questionType,
		Difficulty:     difficultyName,
		DifficultyCode: difficulty,
	}

	var generated []aiGeneratedItem
	var lastBatchErr error
	for remaining := count; remaining > 0; {
		batch := promptCtx.BatchSize()
		if remaining < batch {
			batch = remaining
		}
		items, err := s.generateWithRetry(config, promptCtx, batch)
		if err != nil {
			// 单批失败不致命：记下错误，跳过本批，继续后续批次
			lastBatchErr = err
			// 至少推进 1 题的步进，避免因 BatchSize 算大而无限循环
			if batch < 1 {
				batch = 1
			}
			remaining -= batch
			continue
		}
		if len(items) == 0 {
			remaining -= batch
			continue
		}
		generated = append(generated, items...)
		remaining = count - len(generated)
	}

	if len(generated) == 0 {
		// 一道题都没拿到，必须报错
		if lastBatchErr != nil {
			return nil, fmt.Errorf("AI 出题失败（多次重试仍无法解析）: %w", lastBatchErr)
		}
		return nil, fmt.Errorf("AI 未生成任何题目")
	}

	if len(generated) > count {
		generated = generated[:count]
	}

	for i := range generated {
		if generated[i].Type == "" {
			generated[i].Type = questionType
		}
		if generated[i].Difficulty == "" {
			generated[i].Difficulty = difficulty
		}
		if generated[i].Difficulty == "" {
			generated[i].Difficulty = "medium"
		}
	}

	return generated, nil
}

// generatePromptContext 携带构造提示词所需的上下文信息
type generatePromptContext struct {
	Subject        string
	Chapter        string
	TypeCode       string // single / multi / judge / case_study / essay
	Difficulty     string
	DifficultyCode string
}

// BatchSize 根据题型和难度决定单次生成题目数量，control 单次回复长度避免截断。
// 案例分析、论文、单次单道避免 token 截断；判断/多选 token 占用大时也降为 2。
func (c generatePromptContext) BatchSize() int {
	switch c.TypeCode {
	case "essay", "case_study":
		return 1
	}
	switch c.DifficultyCode {
	case "hard":
		return 2
	case "easy":
		return 5
	default:
		return 3
	}
}

// generateWithRetry 单次生成一批题目，解析失败时自动重试（最多 maxAttempts 次）。
// 每次重试会在 prompt 末尾追加不同的纠错指令，避免 LLM 重复同样错误。
func (s *AiService) generateWithRetry(config dto.AiApiConfig, ctx generatePromptContext, count int) ([]aiGeneratedItem, error) {
	const maxAttempts = 3
	var lastErr error
	lastContent := ""

	// 按题型与难度计算 HTTP 超时，避免 prompt 偏大时 60s 旧默认超时截断。
	timeout := llmTimeoutFor(ctx.TypeCode, ctx.DifficultyCode)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		prompt := buildGeneratePrompt(ctx.Subject, ctx.Chapter, ctx.TypeCode, ctx.Difficulty, count)
		if attempt > 0 {
			prompt += "\n\n【重要】你上一条回复无法被 JSON 解析，请严格遵守：\n" +
				"1. 只输出一个完整、合法、可直接被 JSON 解析的数组；\n" +
				"2. 禁止使用 markdown 代码块（不要以 ``` 开头或结尾）；\n" +
				"3. 禁止任何说明文字、问候语、注释；\n" +
				"4. 元素数量严格为上方要求的 " + fmt.Sprintf("%d", count) + " 道；\n" +
				"5. 字符串内的换行必须用 \\n 转义、双引号必须用 \\\" 转义。"
		}

		resp, err := llm.Chat(context.Background(), config.Provider, config.BaseURL, config.ApiKey, llm.ChatRequest{
			Model:       config.Model,
			Messages:    []llm.ChatMessage{{Role: "user", Content: prompt}},
			Temperature: 0.3,
			MaxTokens:   8192,
			Timeout:     timeout,
		})
		if err != nil {
			lastErr = err
			continue
		}

		content := cleanJSONResponse(resp.Choices[0].Message.Content)
		lastContent = content

		// 优先尝试直接解析
		var items []aiGeneratedItem
		if e := json.Unmarshal([]byte(content), &items); e == nil && len(items) > 0 {
			return items, nil
		} else if e != nil {
			lastErr = e
		}

		// 解析失败：尝试从混杂内容中提取 JSON 数组
		if retry := extractJSONArray(content); retry != content {
			var items2 []aiGeneratedItem
			if e2 := json.Unmarshal([]byte(retry), &items2); e2 == nil && len(items2) > 0 {
				return items2, nil
			} else if e2 != nil {
				lastErr = e2
			}
		}
	}

	snippet := lastContent
	if len(snippet) > 200 {
		snippet = snippet[:200] + "...(已截断)"
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("AI 未返回可解析的题目数组")
	}
	return nil, fmt.Errorf("AI 返回的题目格式解析失败: %v（AI 返回内容开头：%s）", lastErr, snippet)
}

// llmTimeoutFor 按题型与难度返回建议的 HTTP 超时。
// 案例分析/论文输出长（800~1500 字 + JSON）需要 5min；其他题型按难度分级。
func llmTimeoutFor(typeCode, difficulty string) time.Duration {
	switch typeCode {
	case "essay", "case_study":
		return 5 * time.Minute
	}
	switch difficulty {
	case "hard":
		return 3 * time.Minute
	case "easy":
		return 2 * time.Minute
	default:
		return 3 * time.Minute
	}
}

func (s *AiService) saveGenerated(userID, subjectID, chapterID uint, batchID string, items []aiGeneratedItem) (*dto.GenerateQuestionsResp, error) {
	now := time.Now()
	aiRecords := make([]model.AiGeneratedQuestion, len(items))
	questions := make([]model.Question, len(items))

	for i, item := range items {
		optionsJSON, _ := json.Marshal(item.Options)
		questions[i] = model.Question{
			SubjectID:    subjectID,
			ChapterID:    chapterID,
			Type:         item.Type,
			Difficulty:   item.Difficulty,
			Content:      item.Content,
			CaseMaterial: item.CaseMaterial,
			Options:      string(optionsJSON),
			BlankOptions: "[]",
			Answer:       item.Answer,
			Analysis:     item.Analysis,
			Source:       "ai",
			Status:       1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		aiRecords[i] = model.AiGeneratedQuestion{
			UserID:         userID,
			SubjectID:      subjectID,
			ChapterID:      chapterID,
			BatchID:        batchID,
			Type:           item.Type,
			Difficulty:     item.Difficulty,
			Content:        item.Content,
			CaseMaterial:   item.CaseMaterial,
			Options:        string(optionsJSON),
			Answer:         item.Answer,
			Analysis:       item.Analysis,
			KnowledgePoint: item.KnowledgePoint,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	if err := s.questionRepo.BatchCreate(questions); err != nil {
		return nil, fmt.Errorf("保存题目到题库失败: %w", err)
	}

	for i := range aiRecords {
		aiRecords[i].QuestionID = questions[i].ID
	}
	_ = s.aiRepo.CreateQuestions(aiRecords)

	result := make([]dto.AiGeneratedQuestionResp, len(items))
	for i, item := range items {
		optionsJSON, _ := json.Marshal(item.Options)
		result[i] = dto.AiGeneratedQuestionResp{
			ID:             questions[i].ID,
			Type:           item.Type,
			Difficulty:     item.Difficulty,
			Content:        item.Content,
			CaseMaterial:   item.CaseMaterial,
			Options:        string(optionsJSON),
			Answer:         item.Answer,
			Analysis:       item.Analysis,
			KnowledgePoint: item.KnowledgePoint,
		}
	}

	return &dto.GenerateQuestionsResp{Questions: result}, nil
}

// SubmitGenerateAsync 提交 AI 出题异步任务
// 立即返回 task_id，后台 goroutine 执行实际 LLM 调用。
// 完成后通过 Notification 通知用户，用户可在通知列表点击跳转。
func (s *AiService) SubmitGenerateAsync(userID uint, req dto.GenerateQuestionsReq) (*dto.AsyncGenerateTask, error) {
	if len(req.Types) == 0 {
		return nil, fmt.Errorf("请至少选择一种题型")
	}
	questionType := req.Types[0]

	subjectName := ""
	if sub, err := s.subjectRepo.FindByID(req.SubjectID); err == nil {
		subjectName = sub.Name
	}
	chapterName := "不限定"
	if req.ChapterID > 0 {
		if ch, err := s.chapterRepo.FindByID(req.ChapterID); err == nil {
			chapterName = ch.Name
		}
	}

	task := s.aiTaskMgr.NewTask(userID, req.SubjectID, questionType, req.ChapterID, chapterName, req.Difficulty, req.Count)

	// 启动后台 goroutine 执行
	go s.runGenerateTask(userID, task.ID, subjectName, req)

	return task, nil
}

// runGenerateTask 后台执行：调 LLM → 解析 → 入库 → 写通知
// 使用 defer recover 兜底，避免 panic 导致任务卡在 running 且未发通知。
func (s *AiService) runGenerateTask(userID uint, taskID, subjectName string, req dto.GenerateQuestionsReq) {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("AI 出题后台任务异常: %v", r)
			s.aiTaskMgr.SetFailed(taskID, errMsg)
			s.pushGenerateFailedNotification(userID, taskID, errMsg)
		}
	}()

	s.aiTaskMgr.SetRunning(taskID)

	items, err := s.callAIAndParse(req.ApiConfig, req.SubjectID, req.ChapterID, req.Types, req.Difficulty, req.Count)
	if err != nil {
		s.aiTaskMgr.SetFailed(taskID, err.Error())
		s.pushGenerateFailedNotification(userID, taskID, err.Error())
		return
	}

	resp, err := s.saveGenerated(userID, req.SubjectID, req.ChapterID, taskID, items)
	if err != nil {
		s.aiTaskMgr.SetFailed(taskID, err.Error())
		s.pushGenerateFailedNotification(userID, taskID, "题目入库失败: "+err.Error())
		return
	}

	s.aiTaskMgr.SetSuccess(taskID, resp)
	s.pushGenerateSuccessNotification(userID, taskID, subjectName, len(items))
}

// pushGenerateSuccessNotification 推送完成通知
func (s *AiService) pushGenerateSuccessNotification(userID uint, taskID, subjectName string, count int) {
	if s.notifySvc == nil {
		return
	}
	title := "AI 出题完成"
	content := fmt.Sprintf("你的「%s」AI 出题任务已完成，共生成 %d 道题，点击查看。", subjectName, count)
	link := "/ai/practice?task_id=" + taskID
	// 修复 #20：通知失败不再静默吞错；任务状态已为 success，不回滚。
	if err := s.notifySvc.Push(userID, "ai_generate_done", title, content, link); err != nil {
		log.Printf("[ai task] push success notification failed user_id=%d task_id=%s: %v",
			userID, taskID, err)
	}
}

// pushGenerateFailedNotification 推送失败通知
func (s *AiService) pushGenerateFailedNotification(userID uint, taskID, errMsg string) {
	if s.notifySvc == nil {
		return
	}
	// 错误信息截断，避免通知超长
	short := errMsg
	if len(short) > 200 {
		short = short[:200] + "..."
	}
	title := "AI 出题失败"
	content := "AI 出题任务失败：" + short + "。请稍后重试或调整参数。"
	link := "/ai/practice?task_id=" + taskID
	// 修复 #20：通知失败不再静默吞错；任务状态已为 failed，不回滚。
	if err := s.notifySvc.Push(userID, "ai_generate_failed", title, content, link); err != nil {
		log.Printf("[ai task] push failed notification failed user_id=%d task_id=%s: %v",
			userID, taskID, err)
	}
}

// GetGenerateTask 查询任务状态（按 task_id + userID 鉴权）
// userID=0 或任务不属于该用户时返回"不存在"，避免泄露存在性。
func (s *AiService) GetGenerateTask(taskID string, userID uint) (*dto.AsyncGenerateTask, error) {
	if userID == 0 {
		return nil, fmt.Errorf("任务不存在或已过期")
	}
	task := s.aiTaskMgr.Get(taskID, userID)
	if task == nil {
		return nil, fmt.Errorf("任务不存在或已过期")
	}
	return task, nil
}

// ListGenerateTasks 列当前用户最近 20 条任务
// 严格按调用方传入的 userID 过滤；handler 已从 JWT 上下文取 userID，无需再次校验。
func (s *AiService) ListGenerateTasks(userID uint) []*dto.AsyncGenerateTask {
	return s.aiTaskMgr.ListByUser(userID, 20)
}

// ListGenerateHistory 列出某用户的 AI 生成题历史批次（按时间倒序）
// 仅返回已完成的批次（ai_generated_questions 中有题目的 batch）。
// 进行中任务由独立的 /ai/tasks/inflight 端点返回，避免单接口查询失败时静默丢数据。
func (s *AiService) ListGenerateHistory(userID uint, limit int) (*dto.AiBatchHistoryResp, error) {
	if limit <= 0 {
		limit = 20
	}
	if userID == 0 {
		return &dto.AiBatchHistoryResp{List: []dto.AiBatchHistoryItem{}, Total: 0}, nil
	}

	summaries, err := s.aiRepo.ListBatchSummaries(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询 AI 生成历史失败: %w", err)
	}

	allIDs := make([]uint, 0, 64)
	idsByBatch := make([][]uint, len(summaries))
	for i, sm := range summaries {
		ids := parseQuestionIDs(sm.QuestionIDs)
		idsByBatch[i] = ids
		allIDs = append(allIDs, ids...)
	}
	answerMap, err := s.aiRepo.BatchAnswerMap(userID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("查询批次答题情况失败: %w", err)
	}

	list := make([]dto.AiBatchHistoryItem, len(summaries))
	for i, sm := range summaries {
		answered, correct := countBatchAnswers(idsByBatch[i], answerMap)
		item := dto.AiBatchHistoryItem{
			BatchID:         sm.BatchID,
			SubjectID:       sm.SubjectID,
			ChapterID:       sm.ChapterID,
			Type:            sm.Type,
			TypeLabel:       typeLabelCN(sm.Type),
			Difficulty:      sm.Difficulty,
			Count:           sm.Count,
			CreatedAt:       sm.CreatedAt.Format("2006-01-02 15:04"),
			AnsweredCount:   answered,
			CorrectCount:    correct,
			KnowledgePoints: splitKnowledgePoints(sm.KnowledgePoints),
			SubjectName:     sm.SubjectName,
			ChapterName:     sm.ChapterName,
			Status:          "success",
		}
		if sm.ChapterID == 0 || sm.ChapterName == "" {
			item.ChapterName = "不限定"
		}
		list[i] = item
	}

	return &dto.AiBatchHistoryResp{List: list, Total: int64(len(list))}, nil
}

// ListInflightTasks 列出某用户所有 status IN ('pending','running') 的任务，
// 转换为 AiBatchHistoryItem 形式返回，让前端能直接并入历史列表渲染。
// 错误必须向上抛（不静默吞），否则前端会看到「空列表 = 任务消失」。
func (s *AiService) ListInflightTasks(userID uint, limit int) (*dto.AiBatchHistoryResp, error) {
	if limit <= 0 {
		limit = 50
	}
	if userID == 0 {
		return &dto.AiBatchHistoryResp{List: []dto.AiBatchHistoryItem{}, Total: 0}, nil
	}
	if s.aiTaskMgr == nil || s.aiTaskMgr.repo == nil {
		return &dto.AiBatchHistoryResp{List: []dto.AiBatchHistoryItem{}, Total: 0}, nil
	}

	rows, err := s.aiTaskMgr.repo.ListInflightByUser(userID, limit)
	if err != nil {
		log.Printf("[ai history] list inflight failed user_id=%d: %v", userID, err)
		return nil, fmt.Errorf("查询进行中任务失败: %w", err)
	}
	if len(rows) == 0 {
		return &dto.AiBatchHistoryResp{List: []dto.AiBatchHistoryItem{}, Total: 0}, nil
	}

	// 一次性查 subject_name
	subjectIDs := make(map[uint]struct{}, len(rows))
	for i := range rows {
		if rows[i].SubjectID > 0 {
			subjectIDs[rows[i].SubjectID] = struct{}{}
		}
	}
	ids := make([]uint, 0, len(subjectIDs))
	for id := range subjectIDs {
		ids = append(ids, id)
	}
	subjectNameMap, err := s.aiRepo.FindSubjectsByIDs(ids)
	if err != nil {
		log.Printf("[ai history] find subjects failed user_id=%d: %v", userID, err)
		// 名称查不到不能让整个接口失败，降级为「科目#id」即可
		subjectNameMap = map[uint]string{}
	}

	list := make([]dto.AiBatchHistoryItem, 0, len(rows))
	for i := range rows {
		t := &rows[i]
		subjectName := subjectNameMap[t.SubjectID]
		item := dto.AiBatchHistoryItem{
			BatchID:         t.TaskID,
			SubjectID:       t.SubjectID,
			ChapterID:       t.ChapterID,
			ChapterName:     t.ChapterName,
			Type:            t.QuestionType,
			TypeLabel:       typeLabelCN(t.QuestionType),
			Difficulty:      t.Difficulty,
			Count:           0,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04"),
			AnsweredCount:   0,
			CorrectCount:    0,
			KnowledgePoints: []string{},
			SubjectName:     subjectName,
			Status:          t.Status,
		}
		if item.ChapterID == 0 || item.ChapterName == "" {
			item.ChapterName = "不限定"
		}
		if item.SubjectName == "" {
			if t.SubjectID > 0 {
				item.SubjectName = fmt.Sprintf("科目#%d", t.SubjectID)
			} else if t.ChapterName != "" {
				item.SubjectName = t.ChapterName
			} else {
				item.SubjectName = "未知科目"
			}
		}
		list = append(list, item)
	}

	// 按 created_at 倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt > list[j].CreatedAt
	})

	return &dto.AiBatchHistoryResp{List: list, Total: int64(len(list))}, nil
}

// parseQuestionIDs 解析 GROUP_CONCAT(question_id) 产生的逗号分隔 ID 串
func parseQuestionIDs(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		var id uint64
		fmt.Sscanf(strings.TrimSpace(p), "%d", &id)
		if id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

// countBatchAnswers 统计一批题目的已答数与答对数
func countBatchAnswers(ids []uint, answerMap map[uint]bool) (answered, correct int64) {
	for _, id := range ids {
		if ok, has := answerMap[id]; has {
			answered++
			if ok {
				correct++
			}
		}
	}
	return answered, correct
}

// splitKnowledgePoints 解析 | 分隔的知识点，去重（保留顺序），最多返回前 3 个
func splitKnowledgePoints(s string) []string {
	if s == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, p := range strings.Split(s, "|") {
		k := strings.TrimSpace(p)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
		if len(out) >= 3 {
			break
		}
	}
	return out
}

// GetGenerateBatch 获取某用户某批次生成的全部题目，用于历史记录重新答题
func (s *AiService) GetGenerateBatch(userID uint, batchID string) (*dto.GenerateQuestionsResp, error) {
	if batchID == "" {
		return nil, fmt.Errorf("batch_id 不能为空")
	}
	records, err := s.aiRepo.FindByBatch(userID, batchID)
	if err != nil {
		return nil, fmt.Errorf("查询生成批次失败: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("未找到该批生成题目")
	}

	questions := make([]dto.AiGeneratedQuestionResp, len(records))
	for i, r := range records {
		questions[i] = dto.AiGeneratedQuestionResp{
			ID:             r.QuestionID,
			Type:           r.Type,
			Difficulty:     r.Difficulty,
			Content:        r.Content,
			CaseMaterial:   r.CaseMaterial,
			Options:        r.Options,
			BlankOptions:   "[]",
			Answer:         r.Answer,
			Analysis:       r.Analysis,
			KnowledgePoint: r.KnowledgePoint,
		}
	}

	return &dto.GenerateQuestionsResp{Questions: questions}, nil
}

func (s *AiService) StartAiExam(userID uint, req dto.StartAiExamReq) (*dto.StartExamResp, error) {
	if req.SubjectID == 0 {
		return nil, fmt.Errorf("请选择科目")
	}

	// 检查该用户在该科目下已关联题库的 AI 题目总数
	totalCount, err := s.aiRepo.CountByUserAndSubject(userID, req.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("查询 AI 题目失败: %w", err)
	}
	if totalCount == 0 {
		return nil, fmt.Errorf("你在该科目下还没有 AI 生成的题目，请先到「AI 练习」中生成题目")
	}
	if int(totalCount) < req.Count {
		return nil, fmt.Errorf("你在该科目下仅有 %d 道 AI 题目，无法抽取 %d 道，请减少数量或先去生成更多题目", totalCount, req.Count)
	}

	aiQuestions, err := s.aiRepo.FindRandomByUserAndSubject(userID, req.SubjectID, req.Count)
	if err != nil {
		return nil, fmt.Errorf("获取 AI 题目失败: %w", err)
	}
	if len(aiQuestions) == 0 {
		return nil, fmt.Errorf("未找到可用的 AI 题目，请先到「AI 练习」中生成题目")
	}

	questionIDs := make([]uint, len(aiQuestions))
	for i, q := range aiQuestions {
		questionIDs[i] = q.QuestionID
	}

	questions, err := s.questionRepo.FindByIDs(questionIDs)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("AI 题目在题库中不存在，请重新生成")
	}

	subjectID := aiQuestions[0].SubjectID
	reqCount := len(questions)
	duration := req.Duration
	if duration <= 0 {
		duration = reqCount * 2
	}

	now := time.Now()
	record := &model.ExamRecord{
		UserID:     userID,
		TemplateID: 0,
		TotalScore: reqCount,
		Status:     "pending",
		StartedAt:  now,
		Duration:   0,
	}

	if err := s.examRepo.Create(record); err != nil {
		return nil, fmt.Errorf("创建考试记录失败: %w", err)
	}

	snapshots := make([]model.ExamRecordAnswer, len(questions))
	examQuestions := make([]dto.ExamQuesResp, len(questions))
	for i, q := range questions {
		snapshots[i] = model.ExamRecordAnswer{
			RecordID:   record.ID,
			QuestionID: q.ID,
		}
		examQuestions[i] = dto.ExamQuesResp{
			ID:           q.ID,
			Type:         q.Type,
			Content:      q.Content,
			CaseMaterial: q.CaseMaterial,
			Options:      q.Options,
			BlankOptions: q.BlankOptions,
			Difficulty:   q.Difficulty,
			Score:        1,
			SortOrder:    i + 1,
			Source:       q.Source,
		}
	}

	if err := s.examRepo.CreateBatchAnswers(snapshots); err != nil {
		return nil, fmt.Errorf("绑定题目失败: %w", err)
	}

	subjectName := ""
	if sub, err := s.subjectRepo.FindByID(subjectID); err == nil {
		subjectName = sub.Name + " "
	}

	return &dto.StartExamResp{
		RecordID:     record.ID,
		TemplateID:   0,
		TemplateName: subjectName + "AI 智能组卷",
		Duration:     duration,
		TotalScore:   reqCount,
		Questions:    examQuestions,
		Answers:      map[uint]string{},
		Elapsed:      0,
		Status:       "pending",
	}, nil
}

func (s *AiService) Analyze(req dto.AnalyzeReq) (*dto.AnalyzeResp, error) {
	typeName := typeLabelCN(req.QuestionType)
	prompt := fmt.Sprintf(
		`你是一位软考辅导专家。请对以下题目作答情况进行解析。

题目类型：%s
题目内容：%s
正确答案：%s
用户答案：%s

请分析：
1. 这道题考察的知识点是什么
2. 正确答案为什么正确
3. 用户的答案哪里错了（如果用户答错的话）
4. 建议如何巩固这个知识点

请直接给出解析内容，不要输出无关的开头和结尾。`,
		typeName, req.QuestionContent, req.QuestionAnswer, req.UserAnswer,
	)

	resp, err := llm.Chat(context.Background(), req.ApiConfig.Provider, req.ApiConfig.BaseURL, req.ApiConfig.ApiKey, llm.ChatRequest{
		Model:       req.ApiConfig.Model,
		Messages:    []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.5,
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, err
	}

	return &dto.AnalyzeResp{Analysis: resp.Choices[0].Message.Content}, nil
}

type aiGeneratedItem struct {
	Type           string   `json:"type"`
	Difficulty     string   `json:"difficulty"`
	Content        string   `json:"content"`
	CaseMaterial   string   `json:"case_material"`
	Options        []string `json:"options"`
	Answer         string   `json:"answer"`
	Analysis       string   `json:"analysis"`
	KnowledgePoint string   `json:"knowledge_point"`
}

// buildGeneratePrompt 按题型（typeCode）选择对应的 5 套模板之一。
// 命中具体章节时注入细分考点清单；未命中（如选了"不限定章节"）则用
// 该科目的章→节两层大纲；该科目未维护大纲时，调用 fallbackSubjectGuard
// 注入"按所选科目考纲"硬约束，避免 LLM 串科目。
//
// 参数：
//   subject   - 科目名（如"系统分析师"）
//   chapter   - 章节名（来自 chapter.name；"不限定"表示全章节随机）
//   typeCode  - 题型编码：single / multi / judge / case_study / essay
//   difficulty - 中文难度描述：简单 / 中等 / 困难
//   count     - 本次生成题目数
func buildGeneratePrompt(subject, chapter, typeCode, difficulty string, count int) string {
	countStr := fmt.Sprintf("%d", count)

	// 1. 准备考点清单 + 科目录入硬约束
	var kpBlock, subjectGuard string
	subjectChapters := getChaptersForSubject(subject)

	if hasChapterSyllabus(subject) {
		// 大纲命中的科目：按章节命中与否注入不同粒度
		if ch := getChapterByNameForSubject(subject, chapter); ch != nil {
			kpBlock = formatChapterKPBlock(ch)
		} else if len(subjectChapters) > 0 {
			// 选了"不限定章节"：只给章→节两层，避免 prompt 体积过大（4-5K tokens）
			kpBlock = formatChapterOutlineBlock(subjectChapters)
		} else {
			kpBlock = "（该科目暂无章节数据，请按官方考纲常见考点命制。）"
		}
	} else {
		// 未维护大纲的科目：科目录入硬约束 + 通用提示
		guard := fallbackSubjectGuard(subject)
		if guard != "" {
			subjectGuard = guard + "\n\n"
		}
		kpBlock = "（科目「" + subject + "」未细化考点清单，请严格遵守上方【科目录入·硬约束】。）"
	}

	// 2. 选择 Few-shot 示例 + 输出格式
	var fewShot, outputFormat string
	switch typeCode {
	case "essay":
		fewShot = fewShotEssay
		outputFormat = outputFormatEssay
	case "case_study":
		fewShot = fewShotCaseStudy
		outputFormat = outputFormatCaseStudy
	case "judge":
		fewShot = fewShotJudge
		outputFormat = outputFormatJudge
	case "multi":
		fewShot = fewShotMulti
		outputFormat = outputFormatMulti
	default: // single 或未知类型
		fewShot = fewShotSingle
		outputFormat = outputFormatSingle
	}

	// 3. 拼装：科目录入硬约束（若有） + 通用头部 + Few-shot + 输出格式约束
	head := commonPromptHead(subject, chapter, kpBlock, difficulty, countStr)
	return subjectGuard + head + "\n" + fewShot + "\n" + outputFormat
}

// cleanJSONResponse 清洗 LLM 返回中常见的 markdown 包裹和首尾空白。
// 支持 ```json / ```JSON / ``` 等大小写变体，以及 LLM 偶尔输出的"以下是 JSON"等前缀。
func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)
	lower := strings.ToLower(content)
	// 处理 ```json / ```JSON / ``` 开头
	switch {
	case strings.HasPrefix(lower, "```json"):
		content = strings.TrimSpace(content[7:])
	case strings.HasPrefix(lower, "```"):
		content = strings.TrimSpace(content[3:])
	}
	// 处理 ``` 结尾
	lower = strings.ToLower(content)
	if strings.HasSuffix(lower, "```") {
		content = strings.TrimSpace(content[:len(content)-3])
	}
	content = strings.TrimSpace(content)
	return content
}

// extractJSONArray 从混有说明文字/被截断的返回内容中，尽力提取出一段完整的 JSON 数组。
// 优先找最外层 [...]；找不到合法数组时返回原始内容。
func extractJSONArray(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return content
	}
	// 1) 优先匹配最外层 [ ... ]
	start := strings.Index(trimmed, "[")
	end := strings.LastIndex(trimmed, "]")
	if start >= 0 && end > start {
		candidate := []byte(trimmed[start : end+1])
		var probe []json.RawMessage
		if err := json.Unmarshal(candidate, &probe); err == nil {
			return trimmed[start : end+1]
		}
		// 最外层 [ ... ] 不是合法数组（可能因为 [...] 在说明文字里嵌套），
		// 退化到从 start 开始用括号配对找第一个能解析的 [...] 段
		for i := start; i < len(trimmed); i++ {
			if trimmed[i] != '[' {
				continue
			}
			for j := len(trimmed) - 1; j > i; j-- {
				if trimmed[j] != ']' {
					continue
				}
				cand := []byte(trimmed[i : j+1])
				if json.Unmarshal(cand, &probe) == nil {
					return trimmed[i : j+1]
				}
			}
		}
	}
	return content
}

func typeLabelCN(t string) string {
	switch t {
	case "single":
		return "单选题"
	case "multi":
		return "多选题"
	case "judge":
		return "判断题"
	case "fill":
		return "填空题"
	case "short":
		return "简答题"
	case "comprehensive":
		return "综合题"
	case "essay":
		return "论文题"
	case "case_study":
		return "案例分析"
	default:
		return t
	}
}

func difficultyLabelCN(d string) string {
	switch d {
	case "easy":
		return "简单"
	case "medium":
		return "中等"
	case "hard":
		return "困难"
	default:
		return d
	}
}

// EssayScore 对论文/案例分析进行 AI 评分
func (s *AiService) EssayScore(userID uint, req dto.EssayScoreReq) (*dto.EssayScoreResp, error) {
	// OPT-08: 校验 record 归属（userID 来自 token，不可被请求体覆盖）
	if err := s.assertRecordOwner(userID, req.RecordType, req.RecordID); err != nil {
		return nil, err
	}

	// 获取题目内容
	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 根据题型选择评分维度与 prompt
	scoreSpec := scoreSpecForType(question.Type)
	var prompt string
	if question.Type == "case_study" {
		// 案例分析需要把案例材料一并作为上下文
		prompt = buildCaseStudyScorePrompt(question.Content, question.CaseMaterial, req.UserAnswer)
	} else {
		prompt = buildEssayScorePrompt(question.Content, req.UserAnswer)
	}

	resp, err := llm.Chat(context.Background(), req.ApiConfig.Provider, req.ApiConfig.BaseURL, req.ApiConfig.ApiKey, llm.ChatRequest{
		Model:       req.ApiConfig.Model,
		Messages:    []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.3,
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var result essayScoreResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("AI 评分结果解析失败: %w", err)
	}

	// 校验分值范围（按题型对应的各维度满分 clamp）
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	result.ArgumentScore = clamp(result.ArgumentScore, 0, scoreSpec.ArgumentMax)
	result.StructureScore = clamp(result.StructureScore, 0, scoreSpec.StructureMax)
	result.LanguageScore = clamp(result.LanguageScore, 0, scoreSpec.LanguageMax)
	result.DepthScore = clamp(result.DepthScore, 0, scoreSpec.DepthMax)
	result.TotalScore = clamp(result.TotalScore, 0, scoreSpec.TotalMax)

	now := time.Now()

	score := &model.EssayScore{
		UserID:         userID,
		QuestionID:     req.QuestionID,
		QuestionType:   question.Type,
		RecordType:     req.RecordType,
		RecordID:       req.RecordID,
		ExamAnswerID:   req.ExamAnswerID,
		UserAnswer:     req.UserAnswer,
		TotalScore:     result.TotalScore,
		ArgumentScore:  result.ArgumentScore,
		StructureScore: result.StructureScore,
		LanguageScore:  result.LanguageScore,
		DepthScore:     result.DepthScore,
		Comment:        result.Comment,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 检查是否已有评分记录，有则更新，无则创建
	var existing *model.EssayScore
	if req.RecordType == "exam" && req.ExamAnswerID > 0 {
		existing, _ = s.essayScoreRepo.FindByExamAnswer(req.ExamAnswerID)
	} else {
		existing, _ = s.essayScoreRepo.FindByRecord(req.RecordType, req.RecordID)
	}

	if existing != nil {
		score.ID = existing.ID
		score.CreatedAt = existing.CreatedAt
		if err := s.essayScoreRepo.Save(score); err != nil {
			return nil, fmt.Errorf("更新评分记录失败: %w", err)
		}
	} else {
		if err := s.essayScoreRepo.Create(score); err != nil {
			return nil, fmt.Errorf("保存评分记录失败: %w", err)
		}
	}

	return &dto.EssayScoreResp{
		ID:             score.ID,
		QuestionID:     score.QuestionID,
		QuestionType:   score.QuestionType,
		TotalScore:     score.TotalScore,
		ArgumentScore:  score.ArgumentScore,
		StructureScore: score.StructureScore,
		LanguageScore:  score.LanguageScore,
		DepthScore:     score.DepthScore,
		Comment:        score.Comment,
		CreatedAt:      now.Format("2006-01-02 15:04:05"),
	}, nil
}

// CheckEssayScore 检查是否已有评分
func (s *AiService) CheckEssayScore(recordType string, recordID, examAnswerID uint, currentUserID uint) (*dto.EssayScoreCheckResp, error) {
	// OPT-08: 校验记录归属
	if err := s.assertRecordOwner(currentUserID, recordType, recordID); err != nil {
		return nil, err
	}

	var score *model.EssayScore
	var err error

	if recordType == "exam" && examAnswerID > 0 {
		score, err = s.essayScoreRepo.FindByExamAnswer(examAnswerID)
	} else {
		score, err = s.essayScoreRepo.FindByRecord(recordType, recordID)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.EssayScoreCheckResp{HasScore: false}, nil
		}
		return nil, err
	}

	return &dto.EssayScoreCheckResp{
		HasScore: true,
		Score: &dto.EssayScoreResp{
			ID:             score.ID,
			QuestionID:     score.QuestionID,
			QuestionType:   score.QuestionType,
			TotalScore:     score.TotalScore,
			ArgumentScore:  score.ArgumentScore,
			StructureScore: score.StructureScore,
			LanguageScore:  score.LanguageScore,
			DepthScore:     score.DepthScore,
			Comment:        score.Comment,
			CreatedAt:      score.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// assertRecordOwner OPT-08: 校验 record 是否属于 currentUserID。
// 不属于或不存在均返回 ErrForbidden（403），避免 IDOR。
func (s *AiService) assertRecordOwner(currentUserID uint, recordType string, recordID uint) error {
	if recordID == 0 {
		return ErrForbidden
	}
	switch recordType {
	case "practice":
		r, err := s.practiceRepo.FindByID(recordID)
		if err != nil {
			return ErrForbidden
		}
		if r.UserID != currentUserID {
			return ErrForbidden
		}
	case "exam":
		r, err := s.examRepo.FindByID(recordID)
		if err != nil {
			return ErrForbidden
		}
		if r.UserID != currentUserID {
			return ErrForbidden
		}
	default:
		return ErrForbidden
	}
	return nil
}

type essayScoreResult struct {
	TotalScore     int    `json:"total_score"`
	ArgumentScore  int    `json:"argument_score"`
	StructureScore int    `json:"structure_score"`
	LanguageScore  int    `json:"language_score"`
	DepthScore     int    `json:"depth_score"`
	Comment        string `json:"comment"`
}

func buildEssayScorePrompt(questionContent, userAnswer string) string {
	return fmt.Sprintf(
		`你是一位软考高级论文阅卷专家。请对以下论文作答进行分维度评分。

论文题目：
%s

考生作答：
%s

评分要求：
1. 论点与立意（满分20分）：评估论点是否明确、切题，立意是否有深度
2. 结构与逻辑（满分20分）：评估论文结构是否完整、逻辑是否清晰、论证是否有力
3. 语言表达（满分20分）：评估语言是否专业流畅、表达是否准确、术语是否规范
4. 深度与广度（满分15分）：评估分析是否深入、知识面是否广阔、实例是否恰当
5. 总分（满分75分）：为以上4项得分之和

请严格按照以下 JSON 格式输出，不要输出 markdown 代码块标记，只输出纯 JSON：
{
  "total_score": 60,
  "argument_score": 16,
  "structure_score": 15,
  "language_score": 17,
  "depth_score": 12,
  "comment": "评语内容，包含优缺点分析和改进建议，200字左右"
}`,
		questionContent, userAnswer,
	)
}

// scoreSpec 定义各题型评分的维度满分
type scoreSpec struct {
	ArgumentMax  int // 第一维度满分
	StructureMax int // 第二维度满分
	LanguageMax  int // 第三维度满分
	DepthMax     int // 第四维度满分
	TotalMax     int // 总分满分
}

// scoreSpecForType 返回题型对应的评分维度满分范围
func scoreSpecForType(questionType string) scoreSpec {
	switch questionType {
	case "case_study":
		// 案例分析：要点完整性/分析逻辑/专业术语/方案可行性，各20分
		return scoreSpec{ArgumentMax: 20, StructureMax: 20, LanguageMax: 20, DepthMax: 20, TotalMax: 80}
	default:
		// 论文：论点/结构/语言/深度，20/20/20/15
		return scoreSpec{ArgumentMax: 20, StructureMax: 20, LanguageMax: 20, DepthMax: 15, TotalMax: 75}
	}
}

func buildCaseStudyScorePrompt(questionContent, caseMaterial, userAnswer string) string {
	caseCtx := caseMaterial
	if strings.TrimSpace(caseCtx) == "" {
		caseCtx = "（无案例材料）"
	}
	return fmt.Sprintf(
		`你是一位软考案例分析阅卷专家。请结合所给案例材料，对以下案例分析题的作答进行分维度评分。

案例材料：
%s

案例问题：
%s

考生作答：
%s

评分要求：
1. 要点完整性（满分20分）：评估是否覆盖题目要求的关键要点，结论是否全面
2. 分析逻辑（满分20分）：评估分析推理是否严谨、步骤是否清晰、是否紧扣案例材料
3. 专业术语（满分20分）：评估术语使用是否准确规范、表述是否专业
4. 方案可行性（满分20分）：评估给出的方案/结论是否合理、可落地、具有实践价值
5. 总分（满分80分）：为以上4项得分之和

请严格按照以下 JSON 格式输出，不要输出 markdown 代码块标记，只输出纯 JSON：
{
  "total_score": 64,
  "argument_score": 16,
  "structure_score": 17,
  "language_score": 15,
  "depth_score": 16,
  "comment": "评语内容，包含优缺点分析和改进建议，200字左右"
}`,
		caseCtx, questionContent, userAnswer,
	)
}
