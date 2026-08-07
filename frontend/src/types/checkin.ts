import type { Question } from './question'

export interface CheckInQuestion extends Question {
  answered: boolean
}

export interface CheckInTodayResp {
  check_date: string
  completed: boolean
  // 今日已完成打卡，允许重新作答
  retake_available: boolean
  total_count: number
  answered_count: number
  questions: CheckInQuestion[]
  answer_map: Record<number, string>
  result?: CheckInResultResp
}

export interface CheckInAnswerReq {
  question_id: number
  answer: string
  duration: number
}

export interface CheckInResultResp {
  check_date: string
  total_count: number
  correct_count: number
  accuracy: number
  duration: number
}

export interface CheckInHistoryResp {
  date: string
  total_count: number
  correct_count: number
  accuracy: number
}

export interface CheckInStatsResp {
  current_streak: number
  max_streak: number
  total_days: number
  avg_accuracy: number
  today_done: boolean
}
