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
	// CaseMaterial 案例分析题专用：800~1500 字背景材料
	CaseMaterial   string `json:"case_material"`
	Options        string `json:"options"`
	BlankOptions   string `json:"blank_options"`
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

// EssayScoreReq AI 评分请求（论文/案例分析）
type EssayScoreReq struct {
	ApiConfig  AiApiConfig `json:"api_config" binding:"required"`
	QuestionID uint        `json:"question_id" binding:"required"`
	UserAnswer string      `json:"user_answer" binding:"required"`
	// 以下字段用于关联评分记录
	RecordType   string `json:"record_type" binding:"required"`   // practice / exam
	RecordID     uint   `json:"record_id" binding:"required"`     // 练习记录ID或考试记录ID
	ExamAnswerID uint   `json:"exam_answer_id"`                   // 考试答题详情ID（仅考试模式）
}

// EssayScoreResp AI 评分响应（论文/案例分析）
type EssayScoreResp struct {
	ID             uint   `json:"id"`
	QuestionID     uint   `json:"question_id"`
	QuestionType   string `json:"question_type"`
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

// AsyncGenerateTask 异步 AI 出题任务状态
// Status: pending（已提交） / running（生成中） / success（完成） / failed（失败）
type AsyncGenerateTask struct {
	ID           string                  `json:"id"`
	Status       string                  `json:"status"`
	QuestionType string                  `json:"question_type"`
	ChapterID    uint                    `json:"chapter_id"`
	ChapterName  string                  `json:"chapter_name"`
	Difficulty   string                  `json:"difficulty"`
	Count        int                     `json:"count"`
	CreatedAt    string                  `json:"created_at"`
	UpdatedAt    string                  `json:"updated_at"`
	FinishedAt   string                  `json:"finished_at,omitempty"`
	Error        string                  `json:"error,omitempty"`
	Result       *GenerateQuestionsResp  `json:"result,omitempty"`
}

// AsyncSubmitResp 异步任务提交响应
type AsyncSubmitResp struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// AiBatchHistoryItem AI 生成题历史记录（一次生成 = 一个批次）
type AiBatchHistoryItem struct {
	BatchID        string   `json:"batch_id"`
	SubjectID      uint     `json:"subject_id"`
	SubjectName    string   `json:"subject_name"`
	ChapterID      uint     `json:"chapter_id"`
	ChapterName    string   `json:"chapter_name"`
	Type           string   `json:"type"`
	TypeLabel      string   `json:"type_label"`
	Difficulty     string   `json:"difficulty"`
	Count          int64    `json:"count"`
	CreatedAt      string   `json:"created_at"`
	AnsweredCount  int64    `json:"answered_count"`
	CorrectCount   int64    `json:"correct_count"`
	KnowledgePoints []string `json:"knowledge_points"`
}

// AiBatchHistoryResp AI 生成题历史列表响应
type AiBatchHistoryResp struct {
	List  []AiBatchHistoryItem `json:"list"`
	Total int64                `json:"total"`
}
