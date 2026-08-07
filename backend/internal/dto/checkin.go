package dto

// CheckInQuestionResp 打卡题目（含答案/解析，供前端本地即时判分）
type CheckInQuestionResp struct {
	ID           uint   `json:"id"`
	SubjectID    uint   `json:"subject_id"`
	SubSubjectID uint   `json:"sub_subject_id"`
	ChapterID    uint   `json:"chapter_id"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	CaseMaterial string `json:"case_material"`
	Options      string `json:"options"`
	// BlankOptions 多空题专用：每空独立选项 JSON
	BlankOptions string `json:"blank_options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Year         int    `json:"year"`
	Source       string `json:"source"`
	Answered     bool   `json:"answered"`
}

type CheckInTodayResp struct {
	CheckDate     string                `json:"check_date"`
	Completed     bool                  `json:"completed"`
	// RetakeAvailable 今日已完成打卡，允许重新作答
	RetakeAvailable bool                `json:"retake_available"`
	TotalCount    int                   `json:"total_count"`
	AnsweredCount int                   `json:"answered_count"`
	Questions     []CheckInQuestionResp `json:"questions"`
	AnswerMap     map[uint]string       `json:"answer_map"`
	Result        *CheckInResultResp    `json:"result,omitempty"`
}

type CheckInAnswerReq struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
	Duration   int    `json:"duration"`
}

type CheckInResultResp struct {
	CheckDate    string  `json:"check_date"`
	TotalCount   int     `json:"total_count"`
	CorrectCount int     `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
	Duration     int     `json:"duration"`
}

type CheckInHistoryResp struct {
	Date         string  `json:"date"`
	TotalCount   int     `json:"total_count"`
	CorrectCount int     `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
}

type CheckInStatsResp struct {
	CurrentStreak int     `json:"current_streak"`
	MaxStreak     int     `json:"max_streak"`
	TotalDays     int     `json:"total_days"`
	AvgAccuracy   float64 `json:"avg_accuracy"`
	TodayDone     bool    `json:"today_done"`
}
