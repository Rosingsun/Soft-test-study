package dto

type WrongQuestionResp struct {
	ID           uint   `json:"id"`
	QuestionID   uint   `json:"question_id"`
	WrongCount   int    `json:"wrong_count"`
	CorrectCount int    `json:"correct_count"`
	LastWrongAt  string `json:"last_wrong_at"`
	CreatedAt    string `json:"created_at"`
	Content      string `json:"content"`
	Type         string `json:"type"`
	SubjectID    uint   `json:"subject_id"`
}

type WrongPracticeSubmitReq struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
	Duration   int    `json:"duration"`
}
