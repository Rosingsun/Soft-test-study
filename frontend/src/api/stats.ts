import { get } from './request'
import type { StatsOverviewResp, DailyStatsResp, SubjectProgressResp, CalendarStatsResp, ChapterProgressResp } from '@/types/stats'

export function getStatsOverview() {
  return get<StatsOverviewResp>('/stats/overview')
}

export function getDailyStats(days = 30) {
  return get<DailyStatsResp[]>('/stats/daily', { days })
}

export function getSubjectProgress() {
  return get<SubjectProgressResp[]>('/stats/subject-progress')
}

export function getCalendarStats(month: string) {
  return get<CalendarStatsResp[]>('/stats/calendar', { month })
}

export function getChapterProgress() {
  return get<ChapterProgressResp[]>('/stats/chapter-progress')
}
