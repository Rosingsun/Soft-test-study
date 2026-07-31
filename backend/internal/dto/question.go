package dto

type QuestionResp struct {
	ID           uint   `json:"id"`
	SubjectID    uint   `json:"subject_id"`
	SubSubjectID uint   `json:"sub_subject_id"`
	ChapterID    uint   `json:"chapter_id"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	Options      string `json:"options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Year         int    `json:"year"`
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
