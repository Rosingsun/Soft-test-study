package dto

type ExamTemplateResp struct {
	ID            uint   `json:"id"`
	SubjectID     uint   `json:"subject_id"`
	Name          string `json:"name"`
	Duration      int    `json:"duration"`
	TotalScore    int    `json:"total_score"`
	Year          int    `json:"year"`
	QuestionCount int    `json:"question_count"`
}

type StartExamResp struct {
	RecordID     uint            `json:"record_id"`
	TemplateID   uint            `json:"template_id"`
	TemplateName string          `json:"template_name"`
	Duration     int             `json:"duration"`
	TotalScore   int             `json:"total_score"`
	Questions    []ExamQuesResp  `json:"questions"`
	Answers      map[uint]string `json:"answers"`
	Elapsed      int             `json:"elapsed"`
	Status       string          `json:"status"`
}

type ExamQuesResp struct {
	ID         uint   `json:"id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Options    string `json:"options"`
	Difficulty string `json:"difficulty"`
	Score      int    `json:"score"`
	SortOrder  int    `json:"sort_order"`
}

type SubmitAnswerReq struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
}

type ExamResultResp struct {
	RecordID        uint                  `json:"record_id"`
	TemplateName    string                `json:"template_name"`
	SubjectID       uint                  `json:"subject_id"`
	Score           int                   `json:"score"`
	TotalScore      int                   `json:"total_score"`
	Duration        int                   `json:"duration"`
	CorrectCount    int                   `json:"correct_count"`
	TotalCount      int                   `json:"total_count"`
	Status          string                `json:"status"`
	StartedAt       string                `json:"started_at"`
	FinishedAt      string                `json:"finished_at"`
	SectionAccuracy []SectionAccuracyResp `json:"section_accuracy"`
	AnswerDetails   []AnswerDetailResp    `json:"answer_details"`
}

type SectionAccuracyResp struct {
	Type         string  `json:"type"`
	TotalCount   int     `json:"total_count"`
	CorrectCount int     `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
}

type AnswerDetailResp struct {
	QuestionID uint   `json:"question_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Options    string `json:"options"`
	YourAnswer string `json:"your_answer"`
	CorrectAns string `json:"correct_answer"`
	IsCorrect  int    `json:"is_correct"`
	Score      int    `json:"score"`
	Analysis   string `json:"analysis"`
}

type ExamRecordResp struct {
	ID           uint   `json:"id"`
	TemplateID   uint   `json:"template_id"`
	TemplateName string `json:"template_name"`
	SubjectID    uint   `json:"subject_id"`
	Score        int    `json:"score"`
	TotalScore   int    `json:"total_score"`
	Duration     int    `json:"duration"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	CreatedAt    string `json:"created_at"`
}
