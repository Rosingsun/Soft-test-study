package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type QuestionHandler struct {
	svc *service.QuestionService
}

func NewQuestionHandler(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) ListByChapter(c *gin.Context) {
	chapterIDStr := c.Query("chapter_id")
	if chapterIDStr == "" {
		response.Error(c, 10002, "缺少 chapter_id 参数")
		return
	}
	chapterID, err := strconv.ParseUint(chapterIDStr, 10, 64)
	if err != nil {
		response.Error(c, 10002, "chapter_id 参数格式错误")
		return
	}
	difficulty := c.Query("difficulty")
	list, err := h.svc.GetByChapterID(uint(chapterID), difficulty)
	if err != nil {
		response.Error(c, 10001, "获取题目列表失败")
		return
	}
	response.Success(c, list)
}

func (h *QuestionHandler) CaseStudies(c *gin.Context) {
	subSubjectID, ok := parseUintQuery(c, "sub_subject_id")
	if !ok {
		response.Error(c, config.CodeParamError, "缺少或错误的 sub_subject_id 参数")
		return
	}
	year := 0
	if v := c.Query("year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			year = n
		}
	}
	list, err := h.svc.GetCaseStudies(subSubjectID, year)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取案例分析题失败")
		return
	}
	response.Success(c, list)
}

func (h *QuestionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, config.CodeParamError, "参数格式错误")
		return
	}
	q, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, config.CodeNotFound, "题目不存在")
		return
	}
	response.Success(c, q)
}

func (h *QuestionHandler) Random(c *gin.Context) {
	subjectID, ok := parseUintQuery(c, "subject_id")
	if !ok {
		response.Error(c, config.CodeParamError, "缺少或错误的 subject_id 参数")
		return
	}
	count := 10
	if v := c.Query("count"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			count = n
		}
	}
	difficulty := c.Query("difficulty")
	list, err := h.svc.GetRandomQuestions(subjectID, count, difficulty)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取随机题目失败")
		return
	}
	response.Success(c, list)
}

func (h *QuestionHandler) Special(c *gin.Context) {
	subjectID, ok := parseUintQuery(c, "subject_id")
	if !ok {
		response.Error(c, config.CodeParamError, "缺少或错误的 subject_id 参数")
		return
	}
	count := 10
	if v := c.Query("count"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			count = n
		}
	}
	qtype := c.Query("type")
	difficulty := c.Query("difficulty")
	list, err := h.svc.GetSpecialQuestions(subjectID, qtype, difficulty, count)
	if err != nil {
		response.Error(c, config.CodeBadRequest, "获取专项题目失败")
		return
	}
	response.Success(c, list)
}

func parseUintQuery(c *gin.Context, key string) (uint, bool) {
	val := c.Query(key)
	if val == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}
