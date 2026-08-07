package dto

type QuestionResp struct {
	ID           uint   `json:"id"`
	SubjectID    uint   `json:"subject_id"`
	SubSubjectID uint   `json:"sub_subject_id"`
	ChapterID    uint   `json:"chapter_id"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	CaseMaterial string `json:"case_material"`
	Options      string `json:"options"`
	// BlankOptions 多空题专用：每空独立选项 JSON；其他题型为空字符串
	BlankOptions string `json:"blank_options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Year         int    `json:"year"`
	Source       string `json:"source"`
}

type PracticeSubmitReq struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Mode       string `json:"mode" binding:"required"`
	Answer     string `json:"answer"`
	Duration   int    `json:"duration"`
}

type PracticeRecordResp struct {
	ID         uint   `json:"id"`
	QuestionID uint   `json:"question_id"`
	Mode       string `json:"mode"`
	Answer     string `json:"answer"`
	IsCorrect  int    `json:"is_correct"`
	Duration   int    `json:"duration"`
}
