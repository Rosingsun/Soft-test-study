package dto

type QuestionMarkResp struct {
	ID         uint   `json:"id"`
	QuestionID uint   `json:"question_id"`
	CreatedAt  string `json:"created_at"`
}
