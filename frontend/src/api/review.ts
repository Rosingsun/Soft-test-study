import { get, post } from './request'
import type { ReviewTodayResp, ReviewAnswerReq, ReviewAnswerResp, ReviewOverviewResp } from '@/types/review'
import type { PracticeRecordResp } from '@/types/question'

export function getReviewToday() {
  return get<ReviewTodayResp>('/review/today')
}

export function submitReviewAnswer(data: ReviewAnswerReq) {
  return post<PracticeRecordResp & ReviewAnswerResp>('/review/answer', data)
}

export function getReviewOverview() {
  return get<ReviewOverviewResp>('/review/overview')
}
