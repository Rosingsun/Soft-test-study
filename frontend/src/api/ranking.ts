import { get } from './request'
import type { RankingResp, RankingCategory, RankingMetric, EstimatedScoreResp } from '@/types/ranking'

export function getRanking(category: RankingCategory, metric: RankingMetric, subjectId = 0) {
  return get<RankingResp>('/rankings', {
    category,
    metric,
    ...(subjectId ? { subject_id: subjectId } : {}),
  })
}

export function getEstimatedSectionScores(subjectId = 0) {
  return get<EstimatedScoreResp>('/rankings/estimated-section-scores', {
    ...(subjectId ? { subject_id: subjectId } : {}),
  })
}
