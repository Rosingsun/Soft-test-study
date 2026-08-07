import { get, post } from './request'
import type { CheckInTodayResp, CheckInAnswerReq, CheckInResultResp, CheckInHistoryResp, CheckInStatsResp } from '@/types/checkin'
import type { PracticeRecordResp } from '@/types/question'

export function getCheckInToday() {
  return get<CheckInTodayResp>('/check-in/today')
}

export function submitCheckInAnswer(data: CheckInAnswerReq) {
  return post<PracticeRecordResp>('/check-in/answer', data)
}

export function completeCheckIn() {
  return post<CheckInResultResp>('/check-in/complete')
}

export function resetCheckIn() {
  return post<void>('/check-in/reset')
}

export function getCheckInHistory(month: string) {
  return get<CheckInHistoryResp[]>('/check-in/history', { month })
}

export function getCheckInStats() {
  return get<CheckInStatsResp>('/check-in/stats')
}
