package dto

type StudyPlanCreateReq struct {
	Title     string `json:"title" binding:"required"`
	SubjectID uint   `json:"subject_id"`
	DailyGoal int    `json:"daily_goal" binding:"required,min=1"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type StudyPlanUpdateReq struct {
	Title     string `json:"title"`
	SubjectID *uint  `json:"subject_id"`
	DailyGoal *int   `json:"daily_goal"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type StudyPlanResp struct {
	ID           uint    `json:"id"`
	SubjectID    uint    `json:"subject_id"`
	Title        string  `json:"title"`
	DailyGoal    int     `json:"daily_goal"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Status       int     `json:"status"`
	TodayDone    int64   `json:"today_done"`
	TodayGoal    int     `json:"today_goal"`
	OverallDone  int64   `json:"overall_done"`
	OverallGoal  int     `json:"overall_goal"`
	CompletionPct float64 `json:"completion_pct"`
	RemainingDays int    `json:"remaining_days"`
	CreatedAt    string  `json:"created_at"`
}
