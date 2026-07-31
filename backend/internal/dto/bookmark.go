package dto

type CreateFolderReq struct {
	Name string `json:"name" binding:"required"`
}

type FolderResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type QuestionFavoriteResp struct {
	ID         uint   `json:"id"`
	QuestionID uint   `json:"question_id"`
	FolderID   uint   `json:"folder_id"`
	CreatedAt  string `json:"created_at"`
}
