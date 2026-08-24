package dto

// StudyMaterialResp 学习资料响应
type StudyMaterialResp struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	SubjectID     uint   `json:"subject_id"`
	SubjectName   string `json:"subject_name"`
	CoverURL      string `json:"cover_url"`
	FileURL       string `json:"file_url"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	DownloadCount int    `json:"download_count"`
	ViewCount     int    `json:"view_count"`
	CreatedAt     string `json:"created_at"`
}

// MindMapNode 思维导图节点（递归结构，children 为空表示叶子）
type MindMapNode struct {
	Title    string        `json:"title"`
	Children []MindMapNode `json:"children,omitempty"`
}
