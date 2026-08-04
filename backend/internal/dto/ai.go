package dto

type AiApiConfig struct {
	Provider string `json:"provider"`
	ApiKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
}

type GenerateQuestionsReq struct {
	ApiConfig  AiApiConfig `json:"api_config" binding:"required"`
	SubjectID  uint        `json:"subject_id" binding:"required"`
	ChapterID  uint        `json:"chapter_id"`
	Types      []string    `json:"types" binding:"required,min=1"`
	Difficulty string      `json:"difficulty"`
	Count      int         `json:"count" binding:"required,min=1,max=20"`
}

type AiGeneratedQuestionResp struct {
	ID             uint   `json:"id"`
	Type           string `json:"type"`
	Difficulty     string `json:"difficulty"`
	Content        string `json:"content"`
	Options        string `json:"options"`
	Answer         string `json:"answer"`
	Analysis       string `json:"analysis"`
	KnowledgePoint string `json:"knowledge_point"`
}

type GenerateQuestionsResp struct {
	Questions []AiGeneratedQuestionResp `json:"questions"`
}

type AnalyzeReq struct {
	ApiConfig       AiApiConfig `json:"api_config" binding:"required"`
	QuestionContent string      `json:"question_content" binding:"required"`
	QuestionType    string      `json:"question_type" binding:"required"`
	QuestionAnswer  string      `json:"question_answer" binding:"required"`
	UserAnswer      string      `json:"user_answer" binding:"required"`
}

type AnalyzeResp struct {
	Analysis string `json:"analysis"`
}

type AiProvider struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Models   []string `json:"models"`
}

type AiProvidersResp struct {
	Providers []AiProvider `json:"providers"`
}

type StartAiExamReq struct {
	SubjectID uint   `json:"subject_id" binding:"required"`
	Count     int    `json:"count" binding:"required,min=1,max=75"`
	Duration  int    `json:"duration"`
}

// EssayScoreReq 论文 AI 评分请求
type EssayScoreReq struct {
	ApiConfig  AiApiConfig `json:"api_config" binding:"required"`
	QuestionID uint        `json:"question_id" binding:"required"`
	UserAnswer string      `json:"user_answer" binding:"required"`
	// 以下字段用于关联评分记录
	RecordType   string `json:"record_type" binding:"required"`   // practice / exam
	RecordID     uint   `json:"record_id" binding:"required"`     // 练习记录ID或考试记录ID
	ExamAnswerID uint   `json:"exam_answer_id"`                   // 考试答题详情ID（仅考试模式）
}

// EssayScoreResp 论文 AI 评分响应
type EssayScoreResp struct {
	ID             uint   `json:"id"`
	QuestionID     uint   `json:"question_id"`
	TotalScore     int    `json:"total_score"`
	ArgumentScore  int    `json:"argument_score"`
	StructureScore int    `json:"structure_score"`
	LanguageScore  int    `json:"language_score"`
	DepthScore     int    `json:"depth_score"`
	Comment        string `json:"comment"`
	CreatedAt      string `json:"created_at"`
}

// EssayScoreCheckResp 检查论文是否已有评分的响应
type EssayScoreCheckResp struct {
	HasScore bool            `json:"has_score"`
	Score    *EssayScoreResp `json:"score,omitempty"`
}
