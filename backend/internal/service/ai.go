package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/pkg/llm"
	"gorm.io/gorm"
)

type AiService struct {
	aiRepo          *repository.AiRepo
	questionRepo    *repository.QuestionRepo
	subjectRepo     *repository.SubjectRepo
	chapterRepo     *repository.ChapterRepo
	examRepo        *repository.ExamRecordRepo
	essayScoreRepo  *repository.EssayScoreRepo
	practiceRepo    *repository.PracticeRecordRepo
}

func NewAiService(
	aiRepo *repository.AiRepo,
	questionRepo *repository.QuestionRepo,
	subjectRepo *repository.SubjectRepo,
	chapterRepo *repository.ChapterRepo,
	examRepo *repository.ExamRecordRepo,
	essayScoreRepo *repository.EssayScoreRepo,
	practiceRepo *repository.PracticeRecordRepo,
) *AiService {
	return &AiService{
		aiRepo:         aiRepo,
		questionRepo:   questionRepo,
		subjectRepo:    subjectRepo,
		chapterRepo:    chapterRepo,
		examRepo:       examRepo,
		essayScoreRepo: essayScoreRepo,
		practiceRepo:   practiceRepo,
	}
}

func (s *AiService) GetProviders() dto.AiProvidersResp {
	return dto.AiProvidersResp{
		Providers: []dto.AiProvider{
			{Provider: "deepseek", Name: "DeepSeek", BaseURL: "https://api.deepseek.com", Models: []string{"deepseek-chat", "deepseek-reasoner"}},
			{Provider: "openai", Name: "OpenAI", BaseURL: "https://api.openai.com", Models: []string{"gpt-4o-mini", "gpt-4o", "gpt-3.5-turbo"}},
			{Provider: "custom", Name: "自定义", BaseURL: "", Models: []string{}},
		},
	}
}

func (s *AiService) GenerateQuestions(userID uint, req dto.GenerateQuestionsReq) (*dto.GenerateQuestionsResp, error) {
	items, err := s.callAIAndParse(req.ApiConfig, req.SubjectID, req.ChapterID, req.Types, req.Difficulty, req.Count)
	if err != nil {
		return nil, err
	}

	return s.saveGenerated(userID, req.SubjectID, req.ChapterID, items)
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

	typeNames := make([]string, len(types))
	for i, t := range types {
		typeNames[i] = typeLabelCN(t)
	}

	difficultyName := "中等"
	if difficulty != "" {
		difficultyName = difficultyLabelCN(difficulty)
	}

	prompt := buildGeneratePrompt(subjectName, chapterName, strings.Join(typeNames, "、"), difficultyName, count)

	resp, err := llm.Chat(context.Background(), config.BaseURL, config.ApiKey, llm.ChatRequest{
		Model:       config.Model,
		Messages:    []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   4096,
	})
	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var generated []aiGeneratedItem
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("AI 返回的题目格式解析失败: %w", err)
	}

	if len(generated) == 0 {
		return nil, fmt.Errorf("AI 未生成任何题目")
	}

	for i := range generated {
		if generated[i].Type == "" {
			generated[i].Type = types[0]
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

func (s *AiService) saveGenerated(userID, subjectID, chapterID uint, items []aiGeneratedItem) (*dto.GenerateQuestionsResp, error) {
	now := time.Now()
	aiRecords := make([]model.AiGeneratedQuestion, len(items))
	questions := make([]model.Question, len(items))

	for i, item := range items {
		optionsJSON, _ := json.Marshal(item.Options)
		questions[i] = model.Question{
			SubjectID:  subjectID,
			ChapterID:  chapterID,
			Type:       item.Type,
			Difficulty: item.Difficulty,
			Content:    item.Content,
			Options:    string(optionsJSON),
			Answer:     item.Answer,
			Analysis:   item.Analysis,
			Source:     "ai",
			Status:     1,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		aiRecords[i] = model.AiGeneratedQuestion{
			UserID:         userID,
			SubjectID:      subjectID,
			ChapterID:      chapterID,
			Type:           item.Type,
			Difficulty:     item.Difficulty,
			Content:        item.Content,
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
			Options:        string(optionsJSON),
			Answer:         item.Answer,
			Analysis:       item.Analysis,
			KnowledgePoint: item.KnowledgePoint,
		}
	}

	return &dto.GenerateQuestionsResp{Questions: result}, nil
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
			ID:         q.ID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    q.Options,
			Difficulty: q.Difficulty,
			Score:      1,
			SortOrder:  i + 1,
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

	resp, err := llm.Chat(context.Background(), req.ApiConfig.BaseURL, req.ApiConfig.ApiKey, llm.ChatRequest{
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
	Options        []string `json:"options"`
	Answer         string   `json:"answer"`
	Analysis       string   `json:"analysis"`
	KnowledgePoint string   `json:"knowledge_point"`
}

func buildGeneratePrompt(subject, chapter, types, difficulty string, count int) string {
	return fmt.Sprintf(
		`你是一位软考（计算机技术与软件专业技术资格水平考试）出题专家。请根据以下要求生成题目：

- 考试科目：%s
- 知识点范围：%s
- 题目类型：%s
- 难度：%s
- 数量：%d 道

要求：
1. 题目内容专业、准确，符合软考考试大纲
2. 每道题的选项控制在 4 个（多选题也是 4 个选项）
3. 解析要详细，说明为什么对、为什么错
4. type 字段可选值为 single 或 multi
5. 多选题的 answer 格式为逗号分隔的字母，如 "A,C" 表示选 A 和 C

请严格按照以下 JSON 格式输出，不要输出 markdown 代码块标记，只输出纯 JSON 数组：
[
  {
    "type": "single",
    "difficulty": "medium",
    "content": "题目内容（支持Markdown格式）",
    "options": ["A. 选项内容", "B. 选项内容", "C. 选项内容", "D. 选项内容"],
    "answer": "A",
    "analysis": "本题考查...正确答案是A，因为...B选项错误的原因是...",
    "knowledge_point": "知识点名称"
  }
]`,
		subject, chapter, types, difficulty, count,
	)
}

func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
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

// EssayScore 对论文进行 AI 评分
func (s *AiService) EssayScore(userID uint, req dto.EssayScoreReq) (*dto.EssayScoreResp, error) {
	// 获取题目内容
	question, err := s.questionRepo.FindByID(req.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("获取题目信息失败: %w", err)
	}

	// 构建评分 prompt
	prompt := buildEssayScorePrompt(question.Content, req.UserAnswer)

	resp, err := llm.Chat(context.Background(), req.ApiConfig.BaseURL, req.ApiConfig.ApiKey, llm.ChatRequest{
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

	// 校验分值范围
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	result.ArgumentScore = clamp(result.ArgumentScore, 0, 20)
	result.StructureScore = clamp(result.StructureScore, 0, 20)
	result.LanguageScore = clamp(result.LanguageScore, 0, 20)
	result.DepthScore = clamp(result.DepthScore, 0, 15)
	result.TotalScore = clamp(result.TotalScore, 0, 75)

	now := time.Now()

	score := &model.EssayScore{
		UserID:         userID,
		QuestionID:     req.QuestionID,
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
		TotalScore:     score.TotalScore,
		ArgumentScore:  score.ArgumentScore,
		StructureScore: score.StructureScore,
		LanguageScore:  score.LanguageScore,
		DepthScore:     score.DepthScore,
		Comment:        score.Comment,
		CreatedAt:      now.Format("2006-01-02 15:04:05"),
	}, nil
}

// CheckEssayScore 检查论文是否已有评分
func (s *AiService) CheckEssayScore(recordType string, recordID, examAnswerID uint) (*dto.EssayScoreCheckResp, error) {
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
