package dto

type ReviewCardResp struct {
	CardID       uint   `json:"card_id"`
	QuestionID   uint   `json:"id"`
	SubjectID    uint   `json:"subject_id"`
	SubSubjectID uint   `json:"sub_subject_id"`
	ChapterID    uint   `json:"chapter_id"`
	Type         string `json:"type"`
	Difficulty   string `json:"difficulty"`
	Content      string `json:"content"`
	CaseMaterial string `json:"case_material"`
	Options      string `json:"options"`
	// BlankOptions 多空题专用：每空独立选项 JSON；其他题型为空字符串
	BlankOptions string `json:"blank_options"`
	Answer       string `json:"answer"`
	Analysis     string `json:"analysis"`
	Year         int    `json:"year"`
	Repetition   int    `json:"repetition"`
	IntervalDays int    `json:"interval_days"`
	DueDate      string `json:"due_date"`
}

type ReviewTodayResp struct {
	Total     int               `json:"total"`
	Questions []ReviewCardResp  `json:"questions"`
}

type ReviewAnswerReq struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
	Duration   int    `json:"duration"`
}

type ReviewAnswerResp struct {
	QuestionID  uint   `json:"question_id"`
	IsCorrect   int    `json:"is_correct"`
	CorrectAns  string `json:"correct_answer"`
	Analysis    string `json:"analysis"`
	Content     string `json:"content"`
	Options     string `json:"options"`
	// BlankOptions 多空题专用：每空独立选项 JSON
	BlankOptions string `json:"blank_options"`
	Type        string `json:"type"`
	NextDueDate string `json:"next_due_date"`
	Mastered    bool   `json:"mastered"`
}

type ReviewOverviewResp struct {
	DueToday     int64 `json:"due_today"`
	DueTomorrow  int64 `json:"due_tomorrow"`
	Mastered     int64 `json:"mastered"`
	Reviewing    int64 `json:"reviewing"`
	TotalReviews int64 `json:"total_reviews"`
}
