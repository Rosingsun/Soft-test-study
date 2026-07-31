export interface StatsOverviewResp {
  total_practiced: number
  total_correct: number
  accuracy: number
  total_exams: number
  avg_exam_score: number
  study_days: number
  wrong_count: number
}

export interface DailyStatsResp {
  date: string
  total_count: number
  correct_count: number
  incorrect_count: number
}

export interface SubjectProgressResp {
  subject_id: number
  subject_name: string
  total_count: number
  correct_count: number
  accuracy: number
}

export interface CalendarStatsResp {
  date: string
  total_count: number
  correct_count: number
  incorrect_count: number
  duration: number
}

export interface ChapterProgressResp {
  chapter_id: number
  chapter_name: string
  sub_subject_id: number
  total_count: number
  correct_count: number
  accuracy: number
}
