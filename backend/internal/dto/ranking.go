package dto

type RankingScopeResp struct {
	LevelID     uint `json:"level_id"`
	SubjectID   uint `json:"subject_id"`
	TotalUsers  int64 `json:"total_users"`
}

type RankingMyResp struct {
	Rank             int     `json:"rank"`
	Value            float64 `json:"value"`
	TotalParticipants int    `json:"total_participants"`
	Percentile       int     `json:"percentile"`
}

type RankingReferenceResp struct {
	Avg           float64 `json:"avg"`
	Top10Threshold float64 `json:"top10_threshold"`
}

// DistributionItem 单个人数分布分段
type DistributionItem struct {
	Label  string `json:"label"`   // 分段标签，如 "80-100%"
	Count  int    `json:"count"`   // 该分段参与人数
	Ratio  float64 `json:"ratio"`  // 占比（0~1，保留 4 位小数）
	IsMine bool   `json:"is_mine"` // 当前用户是否落在该分段
}

type RankingResp struct {
	Category     string               `json:"category"`
	Metric       string               `json:"metric"`
	Scope        RankingScopeResp     `json:"scope"`
	My           *RankingMyResp       `json:"my"`
	Reference    RankingReferenceResp `json:"reference"`
	Distribution []DistributionItem   `json:"distribution"`
}

// SectionChapterItem 分项预估中的单章节明细
type SectionChapterItem struct {
	ChapterID   uint    `json:"chapter_id"`
	ChapterName string  `json:"chapter_name"`
	Weight      float64 `json:"weight"`
	Attempted   int     `json:"attempted"`
	Correct     int     `json:"correct"`
	Accuracy    float64 `json:"accuracy"`
	EstScore    float64 `json:"est_score"`
}

// SectionBreakdown 单题型（选择题/案例分析/论文）分项预估
type SectionBreakdown struct {
	Section     string               `json:"section"`
	SectionName string               `json:"section_name"`
	MaxScore    float64              `json:"max_score"`
	Attempted   int                  `json:"attempted"`
	EstScore    float64              `json:"est_score"`
	Accuracy    float64              `json:"accuracy"`
	SampleSize  int                  `json:"sample_size"`
	Sufficient  bool                 `json:"sufficient"`
	Chapters    []SectionChapterItem `json:"chapters"`
}

// EstimatedScoreResp 分项预估响应（满分 225 = 75 × 3）
type EstimatedScoreResp struct {
	SubjectID   uint               `json:"subject_id"`
	TotalEst    float64            `json:"total_est"`
	TotalMax    float64            `json:"total_max"`
	Sections    []SectionBreakdown `json:"sections"`
	GeneratedAt string             `json:"generated_at"`
}
