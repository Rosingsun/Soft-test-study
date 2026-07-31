package dto

type StatsOverviewResp struct {
	TotalPracticed int64   `json:"total_practiced"`
	TotalCorrect   int64   `json:"total_correct"`
	Accuracy       float64 `json:"accuracy"`
	TotalExams     int64   `json:"total_exams"`
	AvgExamScore   float64 `json:"avg_exam_score"`
	StudyDays      int64   `json:"study_days"`
	WrongCount     int64   `json:"wrong_count"`
}

type DailyStatsResp struct {
	Date           string `json:"date"`
	TotalCount     int64  `json:"total_count"`
	CorrectCount   int64  `json:"correct_count"`
	IncorrectCount int64  `json:"incorrect_count"`
}

type SubjectProgressResp struct {
	SubjectID    uint    `json:"subject_id"`
	SubjectName  string  `json:"subject_name"`
	TotalCount   int64   `json:"total_count"`
	CorrectCount int64   `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
}

type CalendarStatsResp struct {
	Date           string `json:"date"`
	TotalCount     int64  `json:"total_count"`
	CorrectCount   int64  `json:"correct_count"`
	IncorrectCount int64  `json:"incorrect_count"`
	Duration       int64  `json:"duration"`
}

type ChapterProgressResp struct {
	ChapterID    uint    `json:"chapter_id"`
	ChapterName  string  `json:"chapter_name"`
	SubSubjectID uint    `json:"sub_subject_id"`
	TotalCount   int64   `json:"total_count"`
	CorrectCount int64   `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
}
