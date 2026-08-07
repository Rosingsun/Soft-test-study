export type RankingCategory = 'checkin' | 'exam' | 'practice'
export type RankingMetric = 'streak' | 'accuracy' | 'score'

export interface RankingScopeResp {
  level_id: number
  subject_id: number
  total_users: number
}

export interface RankingMyResp {
  rank: number
  value: number
  total_participants: number
  percentile: number
}

export interface RankingReferenceResp {
  avg: number
  top10_threshold: number
}

export interface DistributionItem {
  label: string
  count: number
  ratio: number
  is_mine: boolean
}

export interface RankingResp {
  category: RankingCategory
  metric: RankingMetric
  scope: RankingScopeResp
  my?: RankingMyResp | null
  reference: RankingReferenceResp
  distribution: DistributionItem[]
}
