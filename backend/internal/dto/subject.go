package dto

type SubjectResp struct {
	ID          uint   `json:"id"`
	LevelID     uint   `json:"level_id"`
	Name        string `json:"name"`
	ShortName   string `json:"short_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

type SubSubjectResp struct {
	ID        uint   `json:"id"`
	SubjectID uint   `json:"subject_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type ChapterResp struct {
	ID            uint   `json:"id"`
	SubjectID     uint   `json:"subject_id"`
	SubSubjectID  uint   `json:"sub_subject_id"`
	ParentID      uint   `json:"parent_id"`
	Name          string `json:"name"`
	SortOrder     int    `json:"sort_order"`
	QuestionCount int64  `json:"question_count"`
	MaterialID    uint   `json:"material_id"`
}
