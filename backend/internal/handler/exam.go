package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type ExamHandler struct {
	svc *service.ExamService
}

func NewExamHandler(svc *service.ExamService) *ExamHandler {
	return &ExamHandler{svc: svc}
}

func (h *ExamHandler) ListTemplates(c *gin.Context) {
	list, err := h.svc.ListTemplates()
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取试卷列表失败")
		return
	}
	response.Success(c, list)
}

func (h *ExamHandler) StartExam(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	resp, err := h.svc.StartExam(userID, uint(id))
	if err != nil {
		respondError(c, err, "开始考试失败")
		return
	}
	response.Success(c, resp)
}

func (h *ExamHandler) SubmitAnswer(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}

	var req dto.SubmitAnswerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, config.CodeParamError, "参数错误")
		return
	}
	if err := h.svc.SubmitAnswer(userID, uint(id), req); err != nil {
		respondError(c, err, "提交答案失败")
		return
	}
	response.Success(c, nil)
}

func (h *ExamHandler) SubmitExam(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	result, err := h.svc.SubmitExam(userID, uint(id))
	if err != nil {
		respondError(c, err, "交卷失败")
		return
	}
	response.Success(c, result)
}

func (h *ExamHandler) GetResult(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	result, err := h.svc.GetResult(userID, uint(id))
	if err != nil {
		respondError(c, err, "获取成绩失败")
		return
	}
	response.Success(c, result)
}

func (h *ExamHandler) ListRecords(c *gin.Context) {
	userID, ok := mustUserID(c)
	if !ok {
		return
	}
	list, err := h.svc.ListRecords(userID)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取考试记录失败")
		return
	}
	response.Success(c, list)
}
