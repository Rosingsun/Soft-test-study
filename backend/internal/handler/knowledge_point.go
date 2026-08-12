package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type KnowledgePointHandler struct {
	svc *service.KnowledgePointService
}

func NewKnowledgePointHandler(svc *service.KnowledgePointService) *KnowledgePointHandler {
	return &KnowledgePointHandler{svc: svc}
}

// Extract AI 提取知识点
func (h *KnowledgePointHandler) Extract(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.ExtractKnowledgePointsReq
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

	resp, err := h.svc.ExtractKnowledgePoints(req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "AI 提取失败: "+err.Error())
		return
	}
	response.Success(c, resp)
}

// Add 加入单条
func (h *KnowledgePointHandler) Add(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.AddKnowledgePointReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	result, err := h.svc.Add(userID, req)
	if err != nil {
		if errors.Is(err, service.ErrConflict) && result != nil {
			existing := dto.ExistingKnowledgePoint{
				UserKPID: result.Existing.ID,
				Name:     result.KP.Name,
			}
			if result.KP.Description != "" {
				existing.Description = result.KP.Description
			}
			existing.IsImportant = result.Existing.IsImportant
			existing.IsMastered = result.Existing.IsMastered
			existing.CreatedAt = result.Existing.CreatedAt.Format("2006-01-02 15:04:05")

			c.JSON(200, gin.H{
				"code":    config.CodeDuplicate,
				"message": "该知识点已存在",
				"data": gin.H{
					"status":   "duplicate",
					"existing": existing,
					"incoming": gin.H{
						"name":             req.Name,
						"description":      req.Description,
						"source_question_id": req.SourceQuestionID,
					},
				},
			})
			return
		}
		response.Error(c, config.CodeBadRequest, "加入失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"status":     result.Status,
		"user_kp_id": result.UserKPID,
		"knowledge":  result.KP,
	})
}

// BatchAdd 批量加入
func (h *KnowledgePointHandler) BatchAdd(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.BatchAddKnowledgePointReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	resp, err := h.svc.BatchAdd(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "批量加入失败: "+err.Error())
		return
	}
	response.Success(c, resp)
}

// ResolveDuplicate 解决冲突
func (h *KnowledgePointHandler) ResolveDuplicate(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return
	}

	var req dto.ResolveDuplicateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	result, err := h.svc.ResolveDuplicate(userID, uint(id), req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "解决冲突失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// List 列表
func (h *KnowledgePointHandler) List(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	var req dto.ListKnowledgePointReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.svc.List(userID, req)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取列表失败: "+err.Error())
		return
	}
	response.SuccessPage(c, list, total, req.Page, req.PageSize)
}

// Update 切换重点/掌握
func (h *KnowledgePointHandler) Update(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return
	}

	var req dto.UpdateUserKnowledgePointReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数校验失败: "+err.Error())
		return
	}

	if req.IsImportant == nil && req.IsMastered == nil {
		response.Error(c, config.CodeParamError, "至少传入一个字段")
		return
	}

	if err := h.svc.UpdateUserKP(userID, uint(id), req); err != nil {
		response.Error(c, config.CodeBadRequest, "更新失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Remove 删除
func (h *KnowledgePointHandler) Remove(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "id 参数错误")
		return
	}

	if err := h.svc.Remove(userID, uint(id)); err != nil {
		response.Error(c, config.CodeBadRequest, "删除失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Stats 统计
func (h *KnowledgePointHandler) Stats(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		response.Error(c, config.CodeUnauthorized, "未登录")
		return
	}

	stats, err := h.svc.Stats(userID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取统计失败: "+err.Error())
		return
	}
	response.Success(c, stats)
}
