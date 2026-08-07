import { get } from './request'
import type { RankingResp, RankingCategory, RankingMetric } from '@/types/ranking'

export function getRanking(category: RankingCategory, metric: RankingMetric, subjectId = 0) {
  return get<RankingResp>('/rankings', {
    category,
    metric,
    ...(subjectId ? { subject_id: subjectId } : {}),
  })
}
