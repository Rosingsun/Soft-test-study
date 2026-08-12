package dto

// ExtractKnowledgePointsReq AI 提取知识点请求
type ExtractKnowledgePointsReq struct {
	ApiConfig        AiApiConfig `json:"api_config" binding:"required"`
	QuestionContent  string      `json:"question_content" binding:"required"`
	QuestionType     string      `json:"question_type" binding:"required"`
	QuestionAnswer   string      `json:"question_answer"`
	QuestionAnalysis string      `json:"question_analysis"`
	SubjectID        uint        `json:"subject_id"`
}

// KnowledgePointItem 知识点条目
type KnowledgePointItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ExtractKnowledgePointsResp AI 提取响应
type ExtractKnowledgePointsResp struct {
	Points []KnowledgePointItem `json:"points"`
}

// AddKnowledgePointReq 加入知识点请求
type AddKnowledgePointReq struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	SubjectID        uint   `json:"subject_id"`
	SourceQuestionID uint   `json:"source_question_id"`
}

// AddKnowledgePointItemResp 单条加入结果
type AddKnowledgePointItemResp struct {
	Name      string                  `json:"name"`
	Status    string                  `json:"status"`
	UserKPID  uint                    `json:"user_kp_id"`
	Knowledge *KnowledgePointResp     `json:"knowledge,omitempty"`
	Existing  *ExistingKnowledgePoint `json:"existing,omitempty"`
}

// ExistingKnowledgePoint 冲突时返回的旧记录
type ExistingKnowledgePoint struct {
	UserKPID      uint   `json:"user_kp_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsImportant   bool   `json:"is_important"`
	IsMastered    bool   `json:"is_mastered"`
	CreatedAt     string `json:"created_at"`
}

// KnowledgePointResp 知识点详情
type KnowledgePointResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SubjectID   uint   `json:"subject_id"`
	Source      string `json:"source"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// UserKnowledgePointResp 用户知识点列表项
type UserKnowledgePointResp struct {
	ID               uint                 `json:"id"`
	UserID           uint                 `json:"user_id"`
	KnowledgePointID uint                 `json:"knowledge_point_id"`
	SourceQuestionID uint                 `json:"source_question_id"`
	IsImportant      bool                 `json:"is_important"`
	IsMastered       bool                 `json:"is_mastered"`
	MasteredAt       string               `json:"mastered_at"`
	Note             string               `json:"note"`
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
	Knowledge        KnowledgePointResp   `json:"knowledge"`
}

// BatchAddKnowledgePointReq 批量加入请求
type BatchAddKnowledgePointReq struct {
	Points           []AddKnowledgePointReq `json:"points" binding:"required,min=1"`
	SourceQuestionID uint                   `json:"source_question_id"`
}

// BatchAddKnowledgePointResp 批量加入响应
type BatchAddKnowledgePointResp struct {
	Added      []AddKnowledgePointItemResp   `json:"added"`
	Duplicates []AddKnowledgePointItemResp   `json:"duplicates"`
}

// ResolveDuplicateReq 冲突解决请求
type ResolveDuplicateReq struct {
	Action      string `json:"action" binding:"required"`
	NewName     string `json:"new_name"`
	NewDesc     string `json:"new_description"`
	SubjectID   uint   `json:"subject_id"`
}

// UpdateUserKnowledgePointReq 更新用户知识点（重点/掌握）
type UpdateUserKnowledgePointReq struct {
	IsImportant *bool `json:"is_important"`
	IsMastered  *bool `json:"is_mastered"`
}

// KnowledgePointStatsResp 统计
type KnowledgePointStatsResp struct {
	Total       int64 `json:"total"`
	Important   int64 `json:"important"`
	Mastered    int64 `json:"mastered"`
	Unmastered  int64 `json:"unmastered"`
	RecentWeek  int64 `json:"recent_week"`
}

// ListKnowledgePointReq 列表查询参数
type ListKnowledgePointReq struct {
	Filter     string `form:"filter"`
	SubjectID  uint   `form:"subject_id"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}
