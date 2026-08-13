package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type AiHandler struct {
	svc *service.AiService
}

func NewAiHandler(svc *service.AiService) *AiHandler {
	return &AiHandler{svc: svc}
}

func (h *AiHandler) GetProviders(c *gin.Context) {
	providers := h.svc.GetProviders()
	response.Success(c, providers)
}

func (h *AiHandler) GenerateQuestions(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.GenerateQuestionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if len(req.Types) == 0 {
		response.Error(c, config.CodeParamError, "请至少选择一种题型")
		return
	}

	for _, t := range req.Types {
		if t != "single" && t != "multi" && t != "essay" {
			response.Error(c, config.CodeParamError, "AI 出题仅支持 single、multi 和 essay 类型")
			return
		}
	}

	if req.ApiConfig.ApiKey == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API Key")
		return
	}

	if req.ApiConfig.BaseURL == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API 地址")
		return
	}

	if req.ApiConfig.Model == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI 模型")
		return
	}

	resp, err := h.svc.GenerateQuestions(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "AI 出题失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// SubmitGenerateAsync 提交 AI 出题异步任务，立即返回 task_id
func (h *AiHandler) SubmitGenerateAsync(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.GenerateQuestionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if len(req.Types) == 0 {
		response.Error(c, config.CodeParamError, "请至少选择一种题型")
		return
	}
	// 异步出题支持 5 种题型
	for _, t := range req.Types {
		if t != "single" && t != "multi" && t != "judge" &&
			t != "case_study" && t != "essay" {
			response.Error(c, config.CodeParamError, "AI 出题仅支持 single / multi / judge / case_study / essay 类型")
			return
		}
	}

	if req.ApiConfig.ApiKey == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API Key")
		return
	}
	if req.ApiConfig.BaseURL == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API 地址")
		return
	}
	if req.ApiConfig.Model == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI 模型")
		return
	}

	task, err := h.svc.SubmitGenerateAsync(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "提交 AI 出题任务失败: "+err.Error())
		return
	}

	response.Success(c, dto.AsyncSubmitResp{TaskID: task.ID, Status: task.Status})
}

// GetGenerateTask 查询异步任务状态
// 必须从 JWT context 取 userID 并传给 service 做归属校验。
// 越权访问（任务存在但不属于当前用户）与"不存在"统一返回 10005，避免泄露任务存在性。
func (h *AiHandler) GetGenerateTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.Error(c, config.CodeParamError, "task_id 不能为空")
		return
	}
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	task, err := h.svc.GetGenerateTask(taskID, userID)
	if err != nil {
		response.Error(c, config.CodeNotFound, err.Error())
		return
	}
	response.Success(c, task)
}

// ListGenerateTasks 列出当前用户最近 20 条 AI 出题任务（按 userID 严格过滤）
func (h *AiHandler) ListGenerateTasks(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}
	tasks := h.svc.ListGenerateTasks(userID)
	response.Success(c, gin.H{"list": tasks})
}

// ListHistory 列出当前用户的 AI 生成题历史批次（已完成）
func (h *AiHandler) ListHistory(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	resp, err := h.svc.ListGenerateHistory(userID, 50)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListInflightTasks 列出当前用户所有进行中的 AI 出题任务（pending/running）。
// 与 /ai/history 分开：进行中任务不进 ai_generated_questions，独立查 ai_generated_tasks。
// 失败时返回 500，不静默吞错——否则用户刷新页面会看到「记录凭空消失」。
func (h *AiHandler) ListInflightTasks(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	resp, err := h.svc.ListInflightTasks(userID, 50)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetBatchQuestions 获取某用户某批次生成的全部题目（历史记录重新答题）
func (h *AiHandler) GetBatchQuestions(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	batchID := c.Param("batch_id")
	resp, err := h.svc.GetGenerateBatch(userID, batchID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AiHandler) Analyze(c *gin.Context) {
	var req dto.AnalyzeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if req.ApiConfig.ApiKey == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API Key")
		return
	}

	resp, err := h.svc.Analyze(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "AI 解析失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AiHandler) StartExam(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.StartAiExamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if req.Count < 1 || req.Count > 75 {
		response.Error(c, config.CodeParamError, "题目数量需在 1-75 之间")
		return
	}

	resp, err := h.svc.StartAiExam(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "AI 组卷失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// EssayScore 论文 AI 评分
func (h *AiHandler) EssayScore(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.EssayScoreReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if req.ApiConfig.ApiKey == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API Key")
		return
	}

	if req.ApiConfig.BaseURL == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI API 地址")
		return
	}

	if req.ApiConfig.Model == "" {
		response.Error(c, config.CodeParamError, "请先配置 AI 模型")
		return
	}

	if req.RecordType != "practice" && req.RecordType != "exam" {
		response.Error(c, config.CodeParamError, "record_type 只能为 practice 或 exam")
		return
	}

	resp, err := h.svc.EssayScore(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "AI 评分失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// CheckEssayScore 检查论文是否已有评分
func (h *AiHandler) CheckEssayScore(c *gin.Context) {
	recordType := c.Query("record_type")
	recordIDStr := c.Query("record_id")
	examAnswerIDStr := c.Query("exam_answer_id")

	if recordType != "practice" && recordType != "exam" {
		response.Error(c, config.CodeParamError, "record_type 只能为 practice 或 exam")
		return
	}

	var recordID, examAnswerID uint
	if recordIDStr != "" {
		var rID uint64
		fmt.Sscanf(recordIDStr, "%d", &rID)
		recordID = uint(rID)
	}
	if examAnswerIDStr != "" {
		var eaID uint64
		fmt.Sscanf(examAnswerIDStr, "%d", &eaID)
		examAnswerID = uint(eaID)
	}

	resp, err := h.svc.CheckEssayScore(recordType, recordID, examAnswerID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "查询评分失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}
