import type { Question } from './question'

export interface ReviewCardQuestion extends Question {
  card_id: number
  repetition: number
  interval_days: number
  due_date: string
}

export interface ReviewTodayResp {
  total: number
  questions: ReviewCardQuestion[]
}

export interface ReviewAnswerReq {
  question_id: number
  answer: string
  duration: number
}

export interface ReviewAnswerResp {
  question_id: number
  is_correct: number
  correct_answer: string
  analysis: string
  content: string
  options: string
  blank_options?: string
  type: string
  next_due_date: string
  mastered: boolean
}

export interface ReviewOverviewResp {
  due_today: number
  due_tomorrow: number
  mastered: number
  reviewing: number
  total_reviews: number
}
